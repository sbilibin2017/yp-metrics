package facades

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricUpdateFacade struct {
	client     *resty.Client
	serverAddr string
}

func NewMetricUpdateFacade(client *resty.Client, serverAddr string) *MetricUpdateFacade {
	return &MetricUpdateFacade{
		client:     client,
		serverAddr: serverAddr,
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

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(compressedBody).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		Post(addr)

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
