package ipify

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type Client interface {
	GetPublicIP(ctx context.Context) (string, error)
}

func NewClient() Client {
	return &client{}
}

type client struct {
}

func (c *client) GetPublicIP(ctx context.Context) (string, error) {
	b := backoff.WithContext(backoff.WithMaxRetries(backoff.NewConstantBackOff(2*time.Second), 3), ctx)
	var ip string
	err := backoff.Retry(func() error {
		c := http.Client{Timeout: 30 * time.Second}
		resp, err := c.Get("https://api.ipify.org")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		ip = strings.TrimSpace(string(b))
		return nil
	}, b)
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", errors.New("public ip is empty")
	}
	return ip, err
}
