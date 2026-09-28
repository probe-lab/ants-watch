package ants

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

// FetchPublicIPv4 returns this host's public IPv4, or "" if all sources fail.
func FetchPublicIPv4(ctx context.Context) string {
	if ip := fetchPublicIPv4FromCheckip(ctx); ip != "" {
		return ip
	}
	slog.Info("falling back to ECS metadata for public IPv4")
	return fetchPublicIPv4FromECS(ctx)
}

// fetchPublicIPv4FromCheckip resolves the public IPv4 via checkip.amazonaws.com.
func fetchPublicIPv4FromCheckip(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	slog.Info("resolving public IPv4 via checkip.amazonaws.com")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://checkip.amazonaws.com", nil)
	if err != nil {
		slog.Warn("build public IP request", "err", err)
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("fetch public IP", "err", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("read public IP response", "err", err)
		return ""
	}

	ip := strings.TrimSpace(string(body))
	if net.ParseIP(ip).To4() == nil {
		slog.Warn("public IP is not IPv4", "value", ip)
		return ""
	}

	slog.Info("resolved public IPv4", "ip", ip)
	return ip
}

// fetchPublicIPv4FromECS resolves the task's public IPv4 via the ECS/EC2 APIs,
// returning "" on any failure.
func fetchPublicIPv4FromECS(ctx context.Context) string {
	cluster, taskARN, err := ecsTaskIdentity(ctx)
	if err != nil {
		slog.Warn("read ECS task metadata", "err", err)
		return ""
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Warn("load AWS config", "err", err)
		return ""
	}

	eniID, err := taskENI(ctx, ecs.NewFromConfig(cfg), cluster, taskARN)
	if err != nil {
		slog.Warn("describe ECS task", "err", err)
		return ""
	}

	ip, err := eniPublicIP(ctx, ec2.NewFromConfig(cfg), eniID)
	if err != nil {
		slog.Warn("describe network interface", "err", err)
		return ""
	}

	if net.ParseIP(ip).To4() == nil {
		slog.Warn("ECS public IP is not IPv4", "value", ip)
		return ""
	}

	slog.Info("resolved public IPv4 via ECS", "ip", ip)
	return ip
}

func ecsTaskIdentity(ctx context.Context) (cluster, taskARN string, err error) {
	uri := os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
	if uri == "" {
		return "", "", fmt.Errorf("ECS_CONTAINER_METADATA_URI_V4 not set")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri+"/task", nil)
	if err != nil {
		return "", "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var meta struct {
		Cluster string `json:"Cluster"`
		TaskARN string `json:"TaskARN"`
	}
	if err := json.Unmarshal(body, &meta); err != nil {
		return "", "", err
	}
	if meta.TaskARN == "" {
		return "", "", fmt.Errorf("no TaskARN in ECS metadata")
	}
	return meta.Cluster, meta.TaskARN, nil
}

func taskENI(ctx context.Context, client *ecs.Client, cluster, taskARN string) (string, error) {
	out, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: &cluster,
		Tasks:   []string{taskARN},
	})
	if err != nil {
		return "", err
	}
	if len(out.Tasks) == 0 {
		return "", fmt.Errorf("task %s not found", taskARN)
	}

	for _, att := range out.Tasks[0].Attachments {
		if att.Type == nil || *att.Type != "ElasticNetworkInterface" {
			continue
		}
		for _, d := range att.Details {
			if d.Name != nil && *d.Name == "networkInterfaceId" && d.Value != nil {
				return *d.Value, nil
			}
		}
	}
	return "", fmt.Errorf("no ENI attachment on task %s", taskARN)
}

func eniPublicIP(ctx context.Context, client *ec2.Client, eniID string) (string, error) {
	out, err := client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{eniID},
	})
	if err != nil {
		return "", err
	}
	if len(out.NetworkInterfaces) == 0 {
		return "", fmt.Errorf("ENI %s not found", eniID)
	}

	assoc := out.NetworkInterfaces[0].Association
	if assoc == nil || assoc.PublicIp == nil {
		return "", fmt.Errorf("ENI %s has no public IP association", eniID)
	}
	return *assoc.PublicIp, nil
}
