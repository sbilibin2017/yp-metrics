package facades

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
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
	hashHeader string
	hashKey    string
}

func NewMetricUpdateFacade(client *resty.Client, serverAddr string, hashHeader string, hashKey string) *MetricUpdateFacade {
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
		hashHeader: hashHeader,
		hashKey:    hashKey,
	}
}

func (f *MetricUpdateFacade) Updates(ctx context.Context, req []types.Metrics) error {
	addr := f.serverAddr
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	addr += "/updates/"

	compressedBody, err := compressBody(req)
	if err != nil {
		return err
	}

	r := f.client.R().
		SetContext(ctx).
		SetBody(compressedBody).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	if f.hashKey != "" {
		hash := computeHMAC(compressedBody, f.hashKey)
		r.SetHeader(f.hashHeader, hash)
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

func compressBody(data []types.Metrics) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)

	_, err = gzw.Write(jsonData)
	if err != nil {
		return nil, err
	}

	if err := gzw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func computeHMAC(data []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
