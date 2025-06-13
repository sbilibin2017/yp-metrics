package facades

import (
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
	hashHeader string
}

func NewMetricUpdateFacade(
	client *resty.Client,
	serverAddr string,
	secretKey string,
	hashHeader string,
) *MetricUpdateFacade {
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
		hashHeader: hashHeader,
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

	var computedHash string
	if f.secretKey != "" {
		sum := sha256.Sum256(append(jsonData, []byte(f.secretKey)...))
		computedHash = hex.EncodeToString(sum[:])
	}

	reqBuilder := f.client.R().
		SetContext(ctx).
		SetBody(jsonData).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	if computedHash != "" {
		reqBuilder.SetHeader(f.hashHeader, computedHash)
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
