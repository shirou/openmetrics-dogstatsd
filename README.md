# OpenMetrics-DogStatsD

`OpenMetrics-DogStatsD` is a program for sending metrics in OpenMetrics format to [DogStatsD](https://docs.datadoghq.com/developers/dogstatsd/) of [Datadog](https://www.datadoghq.com/).

This feature is already available in the regular Datadog agent, so this program is unnecessary for that use case. However, Datadog's IoT Agent does not have this capability, which is why this program was created.

# How to use

## Config file

First, prepare a configuration file. If you name the file `config.yaml`, it will be automatically loaded without needing to specify it as an option.

Here is an example:

```
instances:
  - application_name: node_exporter
    openmetrics_endpoint: http://localhost:9100/metrics
    metrics:
      - node_network_transmit_bytes_total: node.network.transmit.bytes_total
        change: true
      - node_network_receive_bytes_total: node.network.receive.bytes_total
        change: true
      - node_filesystem_free_bytes: node.fs.free_bytes
      - promhttp_metric_handler_requests_total: node.promhttp.requests_count
  - application_name: my_application
    openmetrics_endpoint: http://localhost:8888/metrics
    metrics:
      - node_network_transmit_bytes_total: app.my_application.network.transmit.bytes_total
        change: true
      - node_network_receive_bytes_total: app.my_application.network.receive.bytes_total
        change: true
```

- instances
  - This is a top-level element. This program can read metrics from multiple OpenMetrics sources.
- application_name
  - This is the identifier for each instance. It is only used for readability and is not used within the program.
- openmetrics_endpoint
  -  This is the endpoint for OpenMetrics.
- metrics
  -  List the OpenMetrics metric names as keys and the corresponding DogStatsD metric names as values.
  - If there is a key named `change` and its value is true, the metric will represent the difference since the last value.

Metric types such as counter or gauge will inherit the metric type from OpenMetrics.

## change: true

If change: true is specified within a metric, the value sent to DogStatsD will be the difference since the last value. The previously obtained information is stored under /tmp.

## usage

When run normally, the program executes once and then terminates. If you specify the `-d` option, it will run in daemon mode and continue to execute at the specified interval in seconds.

Running with the `-systemd-service` option will output a Unit file for use as a Systemd Service. Please modify it as needed before use.

```
-addr string
      DogStatsD address (default "127.0.0.1:8125")
-conf string
      config file path (default "./config.yaml")
-d int
      daemon mode with N second (default off)
-systemd-service
      print systemd service file
-v    verbose logging
```

## License

Apache License Version 2.0
