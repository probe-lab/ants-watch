package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

func HealthCheck(ctx *context.Context) error {
	endpoint := fmt.Sprintf(
		"http://%s:%s/health",
		os.Getenv("METRICS_HOST"),
		os.Getenv("METRICS_PORT"),
	)
	resp, err := http.Get(endpoint)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("unhealthy")
}
