package facades

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricUpdateFacade struct {
	client         *resty.Client
	serverAddr     string
	serverEndpoint string
}

func NewMetricUpdateFacade(client *resty.Client, serverAddr string, serverEndpoint string) *MetricUpdateFacade {
	return &MetricUpdateFacade{
		client:         client,
		serverAddr:     serverAddr,
		serverEndpoint: serverEndpoint,
	}
}

func (f *MetricUpdateFacade) Update(ctx context.Context, req types.MetricsUpdatePathRequest) error {
	addr := f.serverAddr
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	baseURL, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("invalid server address: %w", err)
	}

	baseURL.Path = path.Join(baseURL.Path, f.serverEndpoint, req.MType, req.ID, req.Value)

	fullURL := baseURL.String()

	resp, err := f.client.R().
		SetContext(ctx).
		Post(fullURL)

	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
