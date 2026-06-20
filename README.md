# Mentat inter-node network exporter

Mentat runs as a Kubernetes DaemonSet and measures the network path from every
node to every other node. It exports Prometheus metrics for:

- ICMP round-trip latency
- ICMP packet loss
- Effective TCP bandwidth

Each agent exposes Prometheus metrics on port `2112` and serves a fixed-size
bandwidth probe payload on port `2113`. The DaemonSet must publish the bandwidth
port on each node (for example with `hostPort`).

## Metrics

| Metric | Type | Unit |
| --- | --- | --- |
| `node_latency` | Histogram | seconds |
| `node_packet_loss_ratio` | Gauge | ratio from `0` to `1` |
| `node_bandwidth_bytes_per_second` | Gauge | bytes/second |
| `node_bandwidth_probe_failures_total` | Counter | failures |

All metrics use `origin_node` and `destination_node` labels.

## Configuration

| Variable | Default | Meaning |
| --- | --- | --- |
| `NODE_NAME` | Pod hostname | Kubernetes node running this agent |
| `SLEEP_SECONDS` | `5` | Interval between ICMP probe rounds |
| `PING_ATTEMPTS` | `5` | ICMP packets sent to each peer per round |
| `PING_TIMEOUT_SECONDS` | `1` | Timeout for each ICMP packet |
| `BANDWIDTH_PORT` | `2113` | TCP bandwidth endpoint port |
| `BANDWIDTH_BYTES` | `16777216` | Bytes transferred by each bandwidth probe |
| `BANDWIDTH_INTERVAL_SECONDS` | `60` | Interval between bandwidth probe rounds |
| `BANDWIDTH_TIMEOUT_SECONDS` | `30` | Timeout for each bandwidth probe |

Bandwidth probes are sequential per agent but agents operate independently.
Consequently, simultaneous probes can contend with each other; the metric is
effective available bandwidth under current cluster load, not isolated link
capacity.
