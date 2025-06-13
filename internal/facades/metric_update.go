package facades

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricUpdateFacade struct {
	client     *resty.Client
	serverAddr string
	hashKey    string
	hashHeader string
}

func NewMetricUpdateFacade(client *resty.Client, serverAddr string, hashKey, hashHeader string) *MetricUpdateFacade {
	client.
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || (r != nil && r.StatusCode() >= 400)
		})

	return &MetricUpdateFacade{
		client:     client,
		serverAddr: serverAddr,
		hashKey:    hashKey,
		hashHeader: hashHeader,
	}
}

func (f *MetricUpdateFacade) Updates(ctx context.Context, req []types.Metrics) error {
	addr := f.serverAddr
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	addr += "/updates/"

	jsonBytes, err := f.client.JSONMarshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	r := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	if f.hashKey != "" && f.hashHeader != "" {
		h := hmac.New(sha256.New, []byte(f.hashKey))
		h.Write(jsonBytes)
		signature := hex.EncodeToString(h.Sum(nil))
		r.SetHeader(f.hashHeader, signature)
	}

	resp, err := r.Post(addr)
	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
