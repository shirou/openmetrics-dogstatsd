package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", "./config.yaml", "config file path")
	var dogAddr string
	flag.StringVar(&dogAddr, "addr", "127.0.0.1:8125", "DogStatsD address")
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "verbose logging")

	flag.Parse()

	lvl := slog.LevelInfo
	if verbose {
		lvl = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})
	slog.SetDefault(slog.New(handler))

	if err := run(confPath, dogAddr); err != nil {
		slog.Error("run error", slog.Any("error", err))
	}
}

func run(confPath, dogAddr string) error {
	conf, err := readConfig(confPath)
	if err != nil {
		return err
	}

	for i, instance := range conf.Instances {
		dogstatsdClient, err := statsd.New(dogAddr)
		if err != nil {
			return err
		}
		defer dogstatsdClient.Close()
		dogstatsdClient.SetWriteTimeout(10 * time.Second)

		mf, err := readOpenMetrics(instance)
		if err != nil {
			dogstatsdClient.SimpleServiceCheck(fmt.Sprintf("application.%s", instance.ApplicationName), 0)
			slog.Warn("read failed",
				slog.String("application_name", instance.ApplicationName),
				slog.String("endpoint", instance.OpenMetricsEndpoint),
			)
			continue
		}

		dogstatsdClient.SimpleServiceCheck(fmt.Sprintf("application.%s", instance.OpenMetricsEndpoint), 1)

		bufFile := fmt.Sprintf("/tmp/openmetrics_buf-%d.txt", i)
		slog.Debug("bufFile", slog.String("bufFile", bufFile))
		cb, err := newCountBuffer(bufFile)
		if err != nil {
			return err
		}

		for name, metric := range mf {
			// process metric
			for _, metricConf := range instance.Metrics {
				targetName, exist := metricConf[name]
				if !exist {
					continue
				}
				if err := send(dogstatsdClient, cb, targetName, metricConf, metric); err != nil {
					slog.Error("send error", slog.Any("error", err))
					// continue
				}
			}
		}
		if err := cb.flush(bufFile); err != nil {
			return err
		}
	}

	return nil
}
