package fystack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseSize = 10 << 20

// ClientOptions configures a Client.
type ClientOptions struct {
	Credentials   Credentials
	WorkspaceID   string
	Environment   Environment
	Domain        string
	BaseURL       string
	HTTPClient    *http.Client
	Logger        *slog.Logger
	Debug         bool
	PollerOptions PollerOptions
}

// RequestOptions configures headers shared by mutating operations.
type RequestOptions struct {
	IdempotencyKey string
}

// APIRequest describes a low-level request to a Fystack API path. Path must be
// relative to the client's base URL and begin with a slash.
type APIRequest struct {
	Method string
	Path   string
	Body   any
	Header http.Header
}

// Client is a Fystack wallet and automation API client.
type Client struct {
	credentials   Credentials
	workspaceID   string
	baseURL       string
	httpClient    *http.Client
	logger        *slog.Logger
	debug         bool
	pollerOptions PollerOptions
	now           func() time.Time
}

// APIError is returned for an unsuccessful API response.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.StatusCode == 0 {
		return e.Message
	}
	return fmt.Sprintf("fystack API error (%d): %s", e.StatusCode, e.Message)
}

// NewClient creates a Fystack API client.
func NewClient(options ClientOptions) (*Client, error) {
	resolvedBaseURL := strings.TrimRight(options.BaseURL, "/")
	if resolvedBaseURL == "" {
		var err error
		resolvedBaseURL, err = baseURL(options.Environment, options.Domain)
		if err != nil {
			return nil, err
		}
	} else {
		parsed, err := url.Parse(resolvedBaseURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, errors.New("base URL must be an absolute HTTP or HTTPS URL")
		}
	}

	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{
		credentials:   options.Credentials,
		workspaceID:   options.WorkspaceID,
		baseURL:       resolvedBaseURL,
		httpClient:    httpClient,
		logger:        logger,
		debug:         options.Debug,
		pollerOptions: options.PollerOptions,
		now:           time.Now,
	}, nil
}

func (c *Client) log(message string, args ...any) {
	if c.debug {
		c.logger.Info(message, args...)
	}
}

type apiEnvelope struct {
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
	Message string          `json:"message"`
}

func (c *Client) do(ctx context.Context, method, endpoint string, body any, headers http.Header, output any) error {
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header = make(http.Header)
	for key, values := range headers {
		request.Header[key] = append([]string(nil), values...)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if err := c.authenticate(ctx, request, bodyBytes); err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send %s %s: %w", method, endpoint, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	responseBytes, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if len(responseBytes) > maxResponseSize {
		return errors.New("fystack API response exceeds 10 MiB")
	}
	if response.StatusCode == http.StatusNoContent && response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(responseBytes, &envelope); err != nil {
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return &APIError{StatusCode: response.StatusCode, Message: http.StatusText(response.StatusCode)}
		}
		return fmt.Errorf("decode API response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || !envelope.Success {
		message := envelope.Message
		if message == "" {
			message = http.StatusText(response.StatusCode)
			if message == "" {
				message = "request failed"
			}
		}
		return &APIError{StatusCode: response.StatusCode, Message: message}
	}
	if output == nil || len(envelope.Data) == 0 || bytes.Equal(envelope.Data, []byte("null")) {
		return nil
	}
	if err := json.Unmarshal(envelope.Data, output); err != nil {
		return fmt.Errorf("decode API response data: %w", err)
	}
	return nil
}

// Do sends a low-level API request. Most callers should use the typed methods
// on Client or a feature subpackage instead.
func (c *Client) Do(ctx context.Context, request APIRequest, output any) error {
	if request.Method == "" {
		return errors.New("request method is required")
	}
	if !strings.HasPrefix(request.Path, "/") || strings.HasPrefix(request.Path, "//") {
		return errors.New("request path must begin with one slash")
	}
	return c.do(ctx, request.Method, request.Path, request.Body, request.Header, output)
}

func (c *Client) authenticate(ctx context.Context, request *http.Request, body []byte) error {
	credentials := c.credentials
	if credentials.Signer == nil && credentials.APISecret == "" {
		if credentials.AuthToken != "" {
			request.Header.Set("Authorization", credentials.AuthToken)
		}
		return nil
	}
	if strings.TrimSpace(credentials.APIKey) == "" {
		return errors.New("API key is required for signed requests")
	}
	timestamp := strconv.FormatInt(c.now().Unix(), 10)
	canonical := CanonicalString(request.Method, request.URL.EscapedPath(), timestamp, body)
	var signature string
	var err error
	if credentials.Signer != nil {
		signature, err = credentials.Signer.Sign(ctx, canonical)
		if err != nil {
			return fmt.Errorf("sign API request: %w", err)
		}
	} else {
		signature = hmacAccessSignature(credentials.APISecret, canonical)
	}
	request.Header.Set("ACCESS-API-KEY", credentials.APIKey)
	request.Header.Set("ACCESS-TIMESTAMP", timestamp)
	request.Header.Set("ACCESS-SIGN", signature)
	return nil
}

func requestHeaders(options RequestOptions) http.Header {
	headers := make(http.Header)
	if options.IdempotencyKey != "" {
		headers.Set("Idempotency-Key", options.IdempotencyKey)
	}
	return headers
}
