---
sidebar_position: 3
---

# Observability

## Logging

FerretDB writes structured logs to the standard error (`stderr`) stream.
The most recent entries are also available via `getLog` command.

:::note

<!-- https://github.com/FerretDB/FerretDB/issues/3421 -->

Structured log format is not stable yet; field names and formatting of values might change in minor releases.
:::

FerretDB provides the following log formats:

- `console` is a human-readable format with optional colors.
  It colorizes log levels if running terminal supports it.
  To disable ANSI colors, [`NO_COLOR` environment variable](https://no-color.org) could be set to any value.
- `text` is machine-readable [logfmt](https://brandur.org/logfmt)-like format
  (powered by [Go's `slog.TextHandler`](https://pkg.go.dev/log/slog#TextHandler)).
- `json` is machine-readable JSON format
  (powered by [Go's `slog.JSONHandler`](https://pkg.go.dev/log/slog#JSONHandler)).
- `mongo` is machine-readable structured JSON format, similar to the one used in MongoDB.
  It follows the Relaxed Extended JSON specification.
  Fields required in the output format but not yet implemented, will not be included.

There are four logging levels:

<!-- https://github.com/FerretDB/FerretDB/issues/4439 -->

- `error` is used for errors that can't be handled gracefully
  and typically result in client connection being closed;
- `warn` is used for errors that can be handled gracefully
  and typically result in an error being returned to the client (without closing the connection);
- `info` is used for various information messages;
- `debug` should only be used for debugging.

The default level is `info`, except for [development builds](https://pkg.go.dev/github.com/FerretDB/FerretDB/v2/build/version#hdr-Development_builds) that default to `debug`.

:::caution
`debug`-level messages include complete query and response bodies, full error messages, authentication credentials,
and other sensitive information.

Since logs are often retained by the infrastructure
(and FerretDB itself makes recent entries available via the `getLog` command),
that poses a security risk.
Additionally, writing out a significantly larger number of log messages affects FerretDB performance.
For those reasons, the `debug` level should not be enabled in production environments.
:::

The format and level can be adjusted by [configuration flags](flags.md#miscellaneous).

### Docker logs

If Docker was launched with [our quick local setup with Docker Compose](../installation/ferretdb/docker.md#run-production-image),
the following command can be used to fetch the logs.

```sh
docker compose logs ferretdb
```

Otherwise, you can check a list of running Docker containers with `docker ps`
and get logs with `docker logs`.

## OpenTelemetry traces

FerretDB can be configured to send OpenTelemetry traces to the specified HTTP/OTLP URL (e.g. `http://host:4318/v1/traces`).
It can be changed with [`--otel-traces-url` flag](flags.md#miscellaneous).

:::note

<!-- https://github.com/FerretDB/FerretDB/issues/3422 -->

Trace format is not stable yet; attribute names and values might change in minor releases.
:::

## Debug handler

FerretDB exposes various HTTP endpoints with the debug handler on `http://127.0.0.1:8088/debug/` by default.
The host and port can be changed with [`--debug-addr` flag](flags.md#interfaces).

The complete list of handlers is logged on startup
and can also be seen on the http://127.0.0.1:8088/debug/ page itself.

### Archive

FerretDB serves a zip archive with debugging information on the `/debug/archive` endpoint.
Information in the archive helps us debug performance and compatibility problems.

:::caution
Please do not publish the whole archive in our [public community places](/#community),
as it contains sensitive information.
:::

### Metrics

FerretDB exposes metrics in Prometheus format on the `/debug/metrics` endpoint.
There is no need to use an external exporter.

:::note

<!-- https://github.com/FerretDB/FerretDB/issues/3420 -->

The set of metrics is not stable yet; metric and label names and value formatting might change in minor releases.
:::

### Probes

FerretDB exposes the following probes that can be used for
[Kubernetes health checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
and similar use cases.
They return HTTP code 2xx if a probe is successful and 5xx otherwise.
The response body is always empty, but additional information may be present in logs.

- `/debug/livez` is a liveness probe.
  It succeeds if FerretDB is ready to accept new connections from MongoDB protocol clients.
  It does not check the PostgreSQL connection.
  An error response or timeout indicates (after a small initial startup delay) a serious problem.
  Generally, FerretDB should be restarted in that case.
  Additionally, the error is returned during the FerretDB shutdown while it waits for established connections to be closed.
- `/debug/readyz` is a readiness probe.
  It checks that the MongoDB protocol client connection can be established by sending the `ping` command to FerretDB.
  That ensures that the PostgreSQL connection can be established and DocumentDB is installed correctly.
  An error response or timeout indicates a problem with the PostgreSQL or DocumentDB configuration.
