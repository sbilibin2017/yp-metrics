package facades

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/yp-metrics/internal/types"
)

type MetricUpdateFacade struct {
	client         *resty.Client
	serverAddr     string
	serverEndpoint string
}

func NewMetricUpdateFacade(client *resty.Client, serverAddr string) *MetricUpdateFacade {
	return &MetricUpdateFacade{
		client:     client,
		serverAddr: serverAddr,
	}
}

func (f *MetricUpdateFacade) Update(ctx context.Context, req types.Metrics) error {
	addr := f.serverAddr
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	addr += "/update/"

	resp, err := f.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("Content-Type", "application/json").
		Post(addr)

	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}
