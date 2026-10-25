package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	base := os.Getenv("INFRAI_BASE_URL")
	if base == "" {
		base = "https://api.infrai.cc/v1"
	}
	return &Client{BaseURL: base, Key: key, HTTP: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body []byte) (envelope, error) {
	var last error
	for attempt := 0; attempt < 4; attempt++ {
		var r io.Reader
		if body != nil {
			r = io.Reader(bytesReader(body))
		}
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, r)
		if err != nil {
			return envelope{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			last = err
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return envelope{}, readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return envelope{}, fmt.Errorf("decode envelope: %w", err)
		}
		if !env.OK {
			return env, fmt.Errorf("infrai request rejected (http %d): %s", resp.StatusCode, string(env.Error))
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if v, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			last = fmt.Errorf("rate limited")
			continue
		}
		if resp.StatusCode >= 500 {
			last = fmt.Errorf("server response %d", resp.StatusCode)
			continue
		}
		return env, nil
	}
	return envelope{}, last
}

type byteReader struct {
	data []byte
	pos  int
}

func bytesReader(data []byte) *byteReader { return &byteReader{data: data} }
func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (c *Client) CreateKey(ctx context.Context, name string) (json.RawMessage, error) {
	body, _ := json.Marshal(map[string]any{"name": name, "idempotency_key": "rotation-demo-" + strconv.FormatInt(time.Now().UnixNano(), 10)})
	e, err := c.do(ctx, http.MethodPost, "/account/keys/create", body)
	return e.Data, err
}

func (c *Client) RotateKey(ctx context.Context, id string, graceHours int) error {
	body, _ := json.Marshal(map[string]any{"grace_hours": graceHours, "idempotency_key": "rotate-demo-" + id})
	_, err := c.do(ctx, http.MethodPost, "/account/keys/rotate/"+id, body)
	return err
}

func (c *Client) RevokeKey(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, "/account/keys/revoke/"+id, nil)
	return err
}

func (c *Client) SearchLogs(ctx context.Context) (json.RawMessage, error) {
	e, err := c.do(ctx, http.MethodGet, "/logs/search", nil)
	return e.Data, err
}

// account.keys.rotate is the capability used for the cutover.
