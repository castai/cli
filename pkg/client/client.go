package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/castai/cast-cli/pkg/config"
	"github.com/castai/cast-cli/pkg/sdk"
)

const (
	globalTimeout = 1 * time.Minute
)

func New() (*sdk.ClientWithResponses, error) {
	apiToken, err := config.GetCredentials()
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport
	if config.Debug() {
		tr = &loggingTransport{}
	}
	httpClientOption := func(client *sdk.Client) error {
		client.Client = &http.Client{
			Transport: tr,
			Timeout:   globalTimeout,
		}
		return nil
	}

	apiTokenOption := sdk.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("X-API-Key", apiToken.AccessToken)
		return nil
	})

	defaultApiUrl := config.ApiURL()
	var apiUrl string
	if apiToken.ApiUrl != "" {
		apiUrl = apiToken.ApiUrl
	} else {
		apiUrl = defaultApiUrl
	}

	apiClient, err := sdk.NewClientWithResponses(apiUrl, httpClientOption, apiTokenOption)
	if err != nil {
		return nil, err
	}

	return apiClient, nil
}

func DefaultContext() (context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), globalTimeout)
	return ctx, cancel
}

func CheckResponse(response sdk.Response, err error, expectedStatus int) error {
	return checkResponse(response, err, expectedStatus)
}
func checkResponse(response sdk.Response, err error, expectedStatus int) error {
	if err != nil {
		return err
	}

	if response.StatusCode() != expectedStatus {
		return fmt.Errorf("expected status code %d, received: status=%d body=%s", expectedStatus, response.StatusCode(), string(response.GetBody()))
	}

	return nil
}

type loggingTransport struct {
}

// Executes HTTP request with request/response logging.
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// before
	startTime := time.Now()
	requestBody := readBody(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

	fields := log.Fields{
		"method": req.Method,
		"url":    req.URL,
		"body":   redactSensitiveRequestBody(req, string(requestBody)),
	}
	log.WithFields(fields).Debugf("request")

	// real request happens here
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// after
	duration := time.Since(startTime)
	responseBody := readBody(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(responseBody))

	fields = log.Fields{
		"method": resp.Request.Method,
		"url":    resp.Request.URL,
		"code":   resp.StatusCode,
		// We want to see that request-ID
		"headers": resp.Header,
		"time":    duration.Seconds(),
		"body":    redactSensitiveResponseBody(resp, string(responseBody)),
	}

	log.WithFields(fields).Debugf("response")

	return resp, err
}

func redactSensitiveRequestBody(req *http.Request, requestBody string) string {
	requestsToRedact := []func(*http.Request) bool{
		func(req *http.Request) bool {
			return req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/credentials")
		},
	}

	for _, shouldRedact := range requestsToRedact {
		if shouldRedact(req) {
			return redact(requestBody)
		}
	}

	return requestBody
}

func redactSensitiveResponseBody(resp *http.Response, responseBody string) string {
	responsesToRedact := []func(*http.Response) bool{
		func(req *http.Response) bool {
			return resp.Request.Method == http.MethodGet && strings.Contains(resp.Request.URL.Path, "/kubeconfig")
		},
	}

	for _, shouldRedact := range responsesToRedact {
		if shouldRedact(resp) {
			return redact(responseBody)
		}
	}

	return responseBody
}

func redact(text string) string {
	redacted := "<redacted>"
	if len(text) < 10 {
		return redacted
	}
	return fmt.Sprintf("%v...%s", text[:10], redacted)
}

func readBody(body io.ReadCloser) []byte {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = ioutil.ReadAll(body)
	}

	return bodyBytes
}
