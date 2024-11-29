package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

func HealthCheck(c *cli.Context) error {
	endpoint := fmt.Sprintf(
		"http://%s:%s/health",
		os.Getenv("METRICS_HOST"),
		os.Getenv("METRICS_PORT"),
	)
	req, err := http.NewRequestWithContext(c, http.MethodGet, endpoint, nil)
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
