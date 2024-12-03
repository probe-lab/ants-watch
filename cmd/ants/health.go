package main

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

var healthConfig = struct {
	MetricsHost string
	MetricsPort int
}{
	MetricsHost: "127.0.0.1",
	MetricsPort: 5999, // one below the FirstPort to not accidentally override it
}

func HealthCheck(c *cli.Context) error {
	endpoint := fmt.Sprintf(
		"http://%s:%d/metrics",
		healthConfig.MetricsHost, healthConfig.MetricsPort,
	)
	req, err := http.NewRequestWithContext(c.Context, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unhealthy: status code %d", resp.StatusCode)
}
