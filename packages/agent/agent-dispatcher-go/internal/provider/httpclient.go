package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
)

const maxProviderResponseBytes = 4 << 20

type Authentication string

const (
	AuthenticationGitHub Authentication = "github"
	AuthenticationGitLab Authentication = "gitlab"
)

type Request struct {
	Method string
	Path   string
	Query  url.Values
	Body   any
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       json.RawMessage
}

type Client struct {
	baseURL        *url.URL
	token          string
	authentication Authentication
	client         *http.Client
}

func NewClient(
	baseURL *url.URL,
	token string,
	authentication Authentication,
	timeout time.Duration,
) *Client {
	return &Client{
		baseURL:        baseURL,
		token:          token,
		authentication: authentication,
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func NewClientWithHTTP(
	baseURL *url.URL,
	token string,
	authentication Authentication,
	client *http.Client,
) *Client {
	return &Client{baseURL: baseURL, token: token, authentication: authentication, client: client}
}

func (client *Client) Preview(request Request) (domain.RequestPreview, error) {
	target, body, err := client.prepare(request)
	if err != nil {
		return domain.RequestPreview{}, err
	}
	return domain.RequestPreview{
		Requests: []domain.ProviderRequestPreview{{
			Method: request.Method,
			URL:    target.String(),
			Body:   body,
		}},
	}, nil
}

func (client *Client) Do(ctx context.Context, request Request) (Response, error) {
	target, body, err := client.prepare(request)
	if err != nil {
		return Response{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, target.String(), bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("create provider request: %w", err)
	}
	if len(body) > 0 {
		httpRequest.Header.Set("Content-Type", "application/json")
	}
	switch client.authentication {
	case AuthenticationGitHub:
		httpRequest.Header.Set("Authorization", "Bearer "+client.token)
		httpRequest.Header.Set("Accept", "application/vnd.github+json")
		httpRequest.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	case AuthenticationGitLab:
		httpRequest.Header.Set("PRIVATE-TOKEN", client.token)
	}
	response, err := client.client.Do(httpRequest)
	if err != nil {
		return Response{}, &RetryableError{Message: err.Error()}
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, maxProviderResponseBytes+1))
	closeError := response.Body.Close()
	if err != nil && closeError != nil {
		err = errors.Join(err, closeError)
	}
	if err != nil {
		return Response{}, &RetryableError{StatusCode: response.StatusCode, Message: err.Error()}
	}
	if len(content) > maxProviderResponseBytes {
		return Response{}, &PermanentError{StatusCode: response.StatusCode, Message: "provider response exceeds limit"}
	}
	result := Response{
		StatusCode: response.StatusCode,
		Header:     response.Header.Clone(),
		Body:       append(json.RawMessage(nil), content...),
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return result, nil
	}
	githubRateLimit := client.authentication == AuthenticationGitHub &&
		response.StatusCode == http.StatusForbidden &&
		(response.Header.Get("Retry-After") != "" || response.Header.Get("X-RateLimit-Remaining") == "0")
	gitlabResourceLock := client.authentication == AuthenticationGitLab &&
		response.StatusCode == http.StatusConflict
	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 ||
		githubRateLimit || gitlabResourceLock {
		return Response{}, &RetryableError{
			StatusCode: response.StatusCode,
			RetryAfter: retryAfter(response.Header),
			Message:    providerErrorMessage(content),
		}
	}
	return Response{}, &PermanentError{
		StatusCode: response.StatusCode,
		Message:    providerErrorMessage(content),
		Body:       append(json.RawMessage(nil), content...),
	}
}

func (client *Client) prepare(request Request) (*url.URL, json.RawMessage, error) {
	if !strings.HasPrefix(request.Path, "/") || strings.Contains(request.Path, "..") {
		return nil, nil, fmt.Errorf("provider request path is not absolute or contains traversal: %q", request.Path)
	}
	target := *client.baseURL
	target.Path = strings.TrimRight(client.baseURL.Path, "/") + request.Path
	target.RawQuery = request.Query.Encode()
	var body json.RawMessage
	if request.Body != nil {
		encoded, err := json.Marshal(request.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("encode provider request: %w", err)
		}
		body = encoded
	}
	return &target, body, nil
}

func retryAfter(header http.Header) time.Duration {
	value := header.Get("Retry-After")
	if value == "" {
		value = header.Get("X-RateLimit-Reset")
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
			delay := time.Until(time.Unix(seconds, 0))
			if delay > 0 {
				return delay
			}
		}
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if retryAt, err := http.ParseTime(value); err == nil {
		delay := time.Until(retryAt)
		if delay > 0 {
			return delay
		}
	}
	return 0
}

func providerErrorMessage(content []byte) string {
	var response struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if json.Unmarshal(content, &response) == nil {
		if response.Message != "" {
			return response.Message
		}
		if response.Error != "" {
			return response.Error
		}
	}
	message := strings.TrimSpace(string(content))
	if len(message) > 512 {
		return message[:512]
	}
	if message == "" {
		return "provider returned an error"
	}
	return message
}
