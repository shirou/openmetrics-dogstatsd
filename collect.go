package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"gopkg.in/yaml.v3"
)

type appConfig struct {
	Instances []instance `yaml:"instances"`
}

type instance struct {
	ApplicationName     string       `yaml:"application_name"`
	OpenMetricsEndpoint string       `yaml:"openmetrics_endpoint"`
	Metrics             []metricConf `yaml:"metrics"`
}

type metricConf map[string]string

func readConfig(path string) (*appConfig, error) {
	var conf appConfig

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if len(conf.Instances) == 0 {
		return nil, fmt.Errorf("can not parse config file")
	}
	return &conf, nil
}

func readOpenMetrics(ins instance) (map[string]*dto.MetricFamily, error) {
	ctx := context.Background()

	data, err := readEndpoint(ctx, ins.OpenMetricsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to read endpoint: %w", err)
	}

	reader := bytes.NewReader(data)

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func readEndpoint(ctx context.Context, urlStr string) ([]byte, error) {
	slog.Info("connecting to endpoint", slog.String("url", urlStr))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error read data: %w", err)
	}

	return data, nil
}
