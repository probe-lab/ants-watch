package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/urfave/cli/v3"
)

var healthConfig = struct {
	MetricsHost string
	MetricsPort int
}{
	MetricsHost: "127.0.0.1",
	MetricsPort: 5999, // one below the FirstPort to not accidentally override it
}

func healthCommand() *cli.Command {
	return &cli.Command{
		Name:  "health",
		Usage: "Checks the health of the service via the metrics endpoint",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "metrics.host",
				Usage:       "The host serving the metrics endpoint",
				Sources:     cli.EnvVars("ANTS_METRICS_HOST"),
				Destination: &healthConfig.MetricsHost,
				Value:       healthConfig.MetricsHost,
			},
			&cli.IntFlag{
				Name:        "metrics.port",
				Usage:       "The port serving the metrics endpoint",
				Sources:     cli.EnvVars("ANTS_METRICS_PORT"),
				Destination: &healthConfig.MetricsPort,
				Value:       healthConfig.MetricsPort,
			},
		},
		Action: healthCheck,
	}
}

func healthCheck(ctx context.Context, c *cli.Command) error {
	addr := net.JoinHostPort(healthConfig.MetricsHost, strconv.Itoa(healthConfig.MetricsPort))
	endpoint := fmt.Sprintf("http://%s/healthz", addr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unhealthy: status code %d", resp.StatusCode)
}
