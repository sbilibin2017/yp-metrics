package facades

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricUpdateFacade struct {
	client     *resty.Client
	serverAddr string
	secretKey  string
}

func NewMetricUpdateFacade(client *resty.Client, serverAddr string, secretKey string) *MetricUpdateFacade {
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
		secretKey:  secretKey,
	}
}

func (f *MetricUpdateFacade) Updates(ctx context.Context, req []types.Metrics) error {
	addr := f.serverAddr
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	addr += "/updates/"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	var hashHeader string
	if f.secretKey != "" {
		sum := sha256.Sum256(append(jsonData, []byte(f.secretKey)...))
		hashHeader = hex.EncodeToString(sum[:])
	}

	var compressedBuf bytes.Buffer
	gzw := gzip.NewWriter(&compressedBuf)
	if _, err := gzw.Write(jsonData); err != nil {
		return fmt.Errorf("gzip write failed: %w", err)
	}
	if err := gzw.Close(); err != nil {
		return fmt.Errorf("gzip close failed: %w", err)
	}
	compressedBody := compressedBuf.Bytes()

	reqBuilder := f.client.R().
		SetContext(ctx).
		SetBody(compressedBody).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	if hashHeader != "" {
		reqBuilder.SetHeader("HashSHA256", hashHeader)
	}

	resp, err := reqBuilder.Post(addr)
	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
