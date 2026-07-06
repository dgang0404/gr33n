package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultWarmupPoll = 2 * time.Second

// WarmupFarmCounsel triggers POST /guardian/warmup and polls until ready or timeout.
func (c *APIClient) WarmupFarmCounsel(ctx context.Context, timeout time.Duration) error {
	if c == nil || c.HTTP == nil {
		return fmt.Errorf("eval API client not configured")
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	body, _ := json.Marshal(map[string]any{
		"mode":    "farm_counsel",
		"farm_id": c.FarmID,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/guardian/warmup", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("warmup HTTP %d", resp.StatusCode)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state, err := c.healthAwakeningState(ctx)
		if err != nil {
			return err
		}
		if state == "ready" {
			return nil
		}
		if state == "unavailable" {
			return fmt.Errorf("guardian unavailable during warmup")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(defaultWarmupPoll):
		}
	}
	return fmt.Errorf("warmup timed out after %s", timeout)
}

func (c *APIClient) healthAwakeningState(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/v1/chat/health?farm_id=%d&mode=farm_counsel", strings.TrimRight(c.BaseURL, "/"), c.FarmID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("health HTTP %d: %s", resp.StatusCode, truncate(string(body), 120))
	}
	var parsed struct {
		Awakening struct {
			State string `json:"state"`
		} `json:"awakening"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	return strings.TrimSpace(parsed.Awakening.State), nil
}
