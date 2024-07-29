package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	dto "github.com/prometheus/client_model/go"

	"github.com/DataDog/datadog-go/statsd"
)

var bufferFilePath = "/tmp/openmetricsbuffer.txt"

func send(client *statsd.Client, cb *countBuffer,
	targetName string, metricConf metricConf, metric *dto.MetricFamily) error {

	switch *metric.Type {
	case dto.MetricType_COUNTER:
		for _, metric := range metric.Metric {
			value := int64(*metric.Counter.Value)
			name := normalizeName(targetName, metric)
			if metricConf["change"] != "" {
				old := cb.int64(name)
				value -= old
				if value < 0 {
					value = 0
				}
			}
			tags := getTag(metric.Label)
			if err := client.Count(targetName, value, tags, 1); err != nil {
				return err
			}
			cb.updateInt64(name, int64(*metric.Counter.Value))
			slog.Debug("sending",
				slog.String("targetName", targetName),
				slog.String("type", "counter"),
				slog.Int64("value", value),
				slog.Any("tags", tags),
			)
		}
	case dto.MetricType_GAUGE:
		for _, metric := range metric.Metric {
			value := *metric.Gauge.Value
			name := normalizeName(targetName, metric)
			if metricConf["change"] != "" {
				old := cb.float64(name)
				value -= old
				if value < 0 {
					value = 0
				}
			}
			tags := getTag(metric.Label)
			if err := client.Gauge(targetName, value, tags, 1); err != nil {
				return err
			}
			cb.updateFloat64(name, *metric.Gauge.Value)
			slog.Debug("sending",
				slog.String("targetName", targetName),
				slog.String("type", "gauge"),
				slog.Float64("value", value),
				slog.Any("tags", tags),
			)
		}
	case dto.MetricType_HISTOGRAM:
		// not implemented yet
	case dto.MetricType_GAUGE_HISTOGRAM:
		// not implemented yet
	}

	return nil
}

func normalizeName(targetName string, metric *dto.Metric) string {
	// implement me

	elems := make([]string, 0)
	elems = append(elems, targetName)
	elems = append(elems, getTag(metric.Label)...)

	return strings.Join(elems, "_")
}

func getTag(label []*dto.LabelPair) []string {
	tags := make([]string, 0)
	for _, labelPair := range label {
		tags = append(tags,
			fmt.Sprintf("%s:%s", labelPair.GetName(), labelPair.GetValue()),
		)
	}
	return tags
}

type countBuffer struct {
	buf map[string]string
}

func (cb *countBuffer) updateFloat64(key string, value float64) {
	cb.buf[key] = fmt.Sprintf("%f", value)
}
func (cb *countBuffer) updateInt64(key string, value int64) {
	cb.buf[key] = fmt.Sprintf("%d", value)
}

func (cb *countBuffer) flush(path string) error {
	var buf strings.Builder

	for k, v := range cb.buf {
		buf.WriteString(fmt.Sprintf("%s %s\n", k, v))
	}

	return os.WriteFile(path, []byte(buf.String()), os.ModePerm)
}

func (cb *countBuffer) float64(k string) float64 {
	v, exists := cb.buf[k]
	if !exists {
		return 0
	}
	ret, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return ret
}
func (cb *countBuffer) int64(k string) int64 {
	v, exists := cb.buf[k]
	if !exists {
		return 0
	}
	ret, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return ret
}

func newCountBuffer(path string) (*countBuffer, error) {
	cb := countBuffer{
		buf: make(map[string]string),
	}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(data), "\n") {
			s := strings.Fields(line)
			if len(s) != 2 || line[0] == '#' {
				continue
			}
			key := s[0]
			cb.buf[key] = s[1]
		}
	}

	return &cb, nil

}
