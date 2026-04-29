# How I Built a Real-Time Anomaly Detection Engine for a Cloud Storage Platform

## The Problem

Imagine you're running a cloud storage platform used by thousands of people around the world. It's live 24/7. One day, someone starts hammering your server with thousands of requests per second — trying to brute-force accounts, scrape data, or just take the whole thing down.

You need to catch it fast. Not in an hour. Not in a minute. In seconds.

That's exactly what I built for this project: a real-time anomaly detection engine that watches every HTTP request coming into a Nextcloud server, learns what normal traffic looks like, and automatically blocks suspicious IPs using Linux firewall rules — all without any hardcoded thresholds.

---

## What the Project Does

The system runs as a background daemon (a program that runs continuously in the background) alongside a Nextcloud server. Here's what it does at a high level:

1. **Watches** the nginx access log in real time, line by line
2. **Tracks** how many requests each IP is making per second
3. **Learns** what normal traffic looks like over time
4. **Detects** when an IP's behavior deviates significantly from normal
5. **Blocks** the IP using Linux firewall rules (iptables)
6. **Alerts** your team on Slack
7. **Automatically unbans** the IP after a cooldown period
8. **Shows** everything on a live web dashboard

---

## Why This Matters

Traditional security tools like Fail2Ban use fixed rules — "block any IP that makes more than 100 requests per minute." The problem? What's normal for one server is an attack on another. A busy e-commerce site might see 500 requests per minute normally. A small blog might see 5.

Our system doesn't use fixed numbers. It watches your actual traffic, builds a statistical model of what's normal, and only flags things that are genuinely unusual for *your* server. This is called **adaptive anomaly detection**.

---

## The Tech Stack

- **Language**: Go (Golang)
- **Why Go?** Go has native goroutines — lightweight threads that make it easy to tail a log file, update a baseline, serve a dashboard, and process requests all at the same time, with very low memory usage. It compiles to a single static binary with no runtime dependencies.
- **Infrastructure**: Docker Compose, Nginx (reverse proxy), Nextcloud
- **Alerting**: Slack webhooks
- **Blocking**: Linux iptables

---

## How the Sliding Window Works

The core of the rate tracking is a **sliding window** — a data structure that tells you how many requests an IP made in the last 60 seconds, at any point in time.

Think of it like a conveyor belt. Every request gets a timestamp and is placed on the belt. The belt is 60 seconds long. Anything that falls off the end (older than 60 seconds) is discarded. The number of items currently on the belt is the request rate.

In code, this is implemented as a **deque** (double-ended queue) — essentially a slice of timestamps:

```go
type IPStats struct {
    Timestamps []time.Time  // deque of request timestamps
    Errors     int          // count of 4xx/5xx responses
}
```

Every time a new request arrives, we:
1. Append the current timestamp to the end of the deque
2. Evict (remove) all timestamps from the front that are older than 60 seconds
3. The rate is simply `len(Timestamps)`

```go
func (w *Window) evict(now time.Time) {
    cutoff := now.Add(-60 * time.Second)
    i := 0
    for i < len(stats.Timestamps) && stats.Timestamps[i].Before(cutoff) {
        i++
    }
    stats.Timestamps = stats.Timestamps[i:]
}
```

We maintain two deques: one per IP address, and one global (all IPs combined). This lets us detect both individual IP attacks and global traffic spikes.

**Why not just count per minute?** A per-minute counter resets every 60 seconds, creating blind spots. If an attack starts at second 59, you only see 1 second of it before the counter resets. A sliding window always gives you the true last-60-seconds count, regardless of when you check.

---

## How the Baseline Learns from Traffic

The baseline answers the question: "What does normal traffic look like right now?"

Every second, we record the current global request rate as a sample. We keep a **rolling 30-minute window** of these samples — that's up to 1,800 data points. Every 60 seconds, we recalculate the **mean** (average) and **standard deviation** from all samples in the window.

```
mean   = sum(samples) / count(samples)
stddev = sqrt(sum((sample - mean)²) / count(samples))
```

The mean tells us the typical rate. The standard deviation tells us how much it normally varies.

**Per-hour slots**: The system also maintains a separate baseline for each hour of the day (0–23). This is important because traffic at 3am looks very different from traffic at 3pm. When the current hour has accumulated at least 30 samples, the system prefers that hour's baseline over the global rolling mean.

**Floor value**: On a fresh start, there's no traffic yet, so the mean would be 0. Dividing by zero causes problems. We apply a floor of 0.1 — the mean is never allowed to go below this value.

---

## How the Detection Logic Makes a Decision

Once we have a rate and a baseline, detection is simple. An IP is flagged as anomalous if **either** of these conditions fires:

**Condition 1 — Z-score check:**
```
z = (rate - mean) / stddev
if z > 3.0 → anomaly
```
The z-score measures how many standard deviations above the mean the rate is. A z-score above 3.0 means the rate is statistically very unlikely under normal conditions (less than 0.3% probability).

**Condition 2 — Multiplier check:**
```
if rate > 5 × mean → anomaly
```
This catches cases where the standard deviation is very small (very consistent traffic) and the z-score might not fire quickly enough.

**Error surge**: If an IP is generating a lot of 4xx/5xx errors (failed login attempts, scanning for vulnerabilities), both thresholds are automatically halved — making the system more sensitive to that IP's behavior.

**Global anomaly**: The same logic is applied to the global request rate. If the entire server is under a distributed attack, a Slack alert fires — but no single IP is blocked (since the attack is coming from many IPs).

---

## How iptables Blocks an IP

`iptables` is Linux's built-in firewall. It processes every network packet coming into the server and can drop packets from specific IP addresses before they even reach nginx or Nextcloud.

When an anomaly is detected:

```bash
# Add a DROP rule for the offending IP
iptables -A INPUT -s 1.2.3.4 -j DROP
```

Any packet from `1.2.3.4` is now silently dropped at the kernel level — the attacker gets no response, not even an error. From their perspective, the server has disappeared.

We first check if the rule already exists to avoid duplicates:
```bash
iptables -C INPUT -s 1.2.3.4 -j DROP  # -C = check
```

**Auto-unban**: A background goroutine (lightweight thread) sleeps for the ban duration, then removes the rule:
```bash
iptables -D INPUT -s 1.2.3.4 -j DROP  # -D = delete
```

The ban durations follow a backoff schedule:
- 1st offence: 10 minutes
- 2nd offence: 30 minutes  
- 3rd offence: 2 hours
- 4th+ offence: permanent

This means repeat offenders get progressively longer bans, while one-time anomalies (maybe a legitimate user's script went haywire) get a short timeout and can come back.

---

## The Live Dashboard

The system serves a web dashboard at port 8081 that refreshes every 3 seconds, showing:

- Global request rate (last 60 seconds)
- Current baseline mean and standard deviation
- List of currently banned IPs with ban timestamps
- Top 10 source IPs by request volume
- Memory usage and uptime
- Baseline mean by hour (bar chart showing how traffic varies across the day)

---

## Putting It All Together

Here's the full flow when an attack happens:

1. Attacker sends 500 requests/second from IP `1.2.3.4`
2. The sliding window records 500 timestamps for that IP in the last 60 seconds
3. The baseline says normal traffic is ~2 req/s with stddev 0.5
4. Z-score: `(500 - 2) / 0.5 = 996` — way above 3.0
5. `iptables -A INPUT -s 1.2.3.4 -j DROP` fires
6. Slack alert sent: "🚨 BAN | IP: 1.2.3.4 | z-score | Rate: 500 | Baseline: 2.00"
7. Audit log entry written
8. After 10 minutes, the ban is automatically lifted and another Slack alert fires

The whole detection-to-block cycle happens within milliseconds of the anomalous log line being written.

---

## Key Takeaways

- **Adaptive thresholds** beat fixed rules — the system learns your traffic, not someone else's
- **Sliding windows** give accurate real-time rates without blind spots
- **Per-hour baselines** handle natural traffic variation across the day
- **iptables** is the right tool for blocking at the kernel level — fast, reliable, no application overhead
- **Go goroutines** make it easy to do many things concurrently with minimal resources

The full source code is available at: [https://github.com/devLateef/detector-app-hng14-stage3](https://github.com/devLateef/detector-app-hng14-stage3)
