package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type loggingTransport struct {
	log logrus.FieldLogger
}

// Executes HTTP request with request/response logging.
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// before
	startTime := time.Now()
	requestBody := readBody(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))

	fields := logrus.Fields{
		"method": req.Method,
		"url":    req.URL,
		"body":   redactSensitiveRequestBody(req, string(requestBody)),
	}
	t.log.WithFields(fields).Debugf("request")

	// Real request happens here.
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// After.
	duration := time.Since(startTime)
	responseBody := readBody(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(responseBody))

	fields = logrus.Fields{
		"method": resp.Request.Method,
		"url":    resp.Request.URL,
		"code":   resp.StatusCode,
		// We want to see that request-ID
		"headers": resp.Header,
		"time":    duration.Seconds(),
		"body":    redactSensitiveResponseBody(resp, string(responseBody)),
	}

	t.log.WithFields(fields).Debugf("response")

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
