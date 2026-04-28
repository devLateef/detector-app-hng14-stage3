# HNG Anomaly Detection Engine

A real-time HTTP traffic anomaly detector built in Go, running alongside Nextcloud on Docker.

## Server Details

- **Server IP**: `<YOUR_SERVER_IP>`
- **Metrics Dashboard**: `http://<YOUR_DOMAIN>:8081`
- **Nextcloud**: `http://<YOUR_SERVER_IP>` (IP only)

## Language Choice

Go was chosen for:
- Native goroutines for concurrent log tailing, baseline updates, and HTTP serving
- Low memory footprint suitable for a 2GB RAM VPS
- Static binary — no runtime dependencies in the container
- Strong standard library (net/http, sync, os/exec for iptables)

## How the Sliding Window Works

Each request is timestamped and appended to a **deque** (slice used as a queue):
- One deque per IP (`perIP map[string]*IPStats`)
- One global deque

On every `Add()` call, entries older than 60 seconds are evicted from the front of each deque. The rate is simply `len(deque)` — the number of timestamps still in the window.

This gives an accurate per-second rate over the last 60 seconds without any rate-limiting libraries.

## How the Baseline Works

- **Window**: 30-minute rolling window of per-second global request counts
- **Recalculation**: Every 60 seconds, mean and stddev are recomputed from all samples in the window
- **Per-hour slots**: Each recalculation updates the current hour's slot (0–23). When the current hour has ≥ 30 samples, its slot is preferred over the global rolling mean
- **Floor value**: Mean is floored at 0.1 to prevent division-by-zero on cold start

## Anomaly Detection Logic

An IP is flagged if **either** condition fires:
1. **Z-score**: `(rate - mean) / stddev > 3.0`
2. **Multiplier**: `rate > 5 × mean`

If the IP has an **error surge** (4xx/5xx rate ≥ 3× baseline error mean), both thresholds are halved, making detection more sensitive.

Global traffic spikes use the same logic but trigger a Slack alert only (no IP block).

## How iptables Blocking Works

When an anomaly is detected:
1. `iptables -C INPUT -s <ip> -j DROP` checks if a rule already exists
2. If not, `iptables -A INPUT -s <ip> -j DROP` adds the DROP rule
3. A goroutine sleeps for the ban duration, then calls `iptables -D` in a loop to remove all matching rules
4. Unban durations follow a backoff: 10 min → 30 min → 2 hours → permanent

## Setup Instructions (Fresh VPS)

```bash
# 1. Install Docker and Docker Compose
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 2. Clone the repository
git clone https://github.com/<YOUR_USERNAME>/<YOUR_REPO>.git
cd <YOUR_REPO>

# 3. Create .env with your Slack webhook
echo "SLACK_WEBHOOK=https://hooks.slack.com/services/..." > .env

# 4. Start the stack
docker compose up -d --build

# 5. Verify all containers are running
docker compose ps

# 6. Check detector logs
docker logs detector -f
```

## GitHub Repository

[https://github.com/<YOUR_USERNAME>/<YOUR_REPO>](https://github.com/<YOUR_USERNAME>/<YOUR_REPO>)

## Blog Post

[Link to blog post](<YOUR_BLOG_URL>)
