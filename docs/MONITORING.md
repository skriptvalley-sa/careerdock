# CareerDock — Monitoring & Observability

> **Version:** 1.0
> **Status:** Draft (Phase 8)
> **Last updated:** 2026-03-12
> **Depends on:** [ARCHITECTURE.md](./ARCHITECTURE.md), [DEPLOYMENT.md](./DEPLOYMENT.md), [SECURITY.md](./SECURITY.md), [ADMIN-PANEL.md](./ADMIN-PANEL.md)

---

## 1. Overview

CareerDock's observability stack is built on free-tier services — sufficient for an MVP serving up to 1,000 concurrent users.

| Pillar | Tool | Free Tier Limit |
|--------|------|----------------|
| **Logging** | CloudWatch Logs (via Docker awslogs driver) | 5GB ingestion/month |
| **Metrics** | Grafana Cloud (Prometheus-compatible) | 10K active series |
| **Error tracking** | Sentry | 5K events/month |
| **Uptime** | UptimeRobot | 50 monitors |
| **Dashboards** | Grafana Cloud + Admin panel (`/admin`) | Included in free tier |
| **AI cost tracking** | Custom (admin API + Asynq periodic task) | N/A (self-built) |

### 1.1 What's Covered Elsewhere

| Topic | Document |
|-------|----------|
| Log format, CloudWatch groups, retention, viewing logs | [DEPLOYMENT.md §8](./DEPLOYMENT.md) |
| Health check endpoint (`/api/health`) | [DEPLOYMENT.md §7](./DEPLOYMENT.md) |
| Log redaction (what must never be logged) | [SECURITY.md §9.4](./SECURITY.md) |
| Incident response playbook (P1-P3 severity) | [SECURITY.md §12](./SECURITY.md) |
| AI cost dashboard UI and admin API | [ADMIN-PANEL.md §8](./ADMIN-PANEL.md) |

This document covers: **application instrumentation, metric definitions, alerting rules, Sentry integration, SLOs, Grafana dashboards, and operational queries.**

---

## 2. Application Instrumentation

### 2.1 HTTP Metrics Middleware

Every request emits metrics via a Chi middleware. Metrics are exposed as Prometheus counters and histograms on a `/metrics` endpoint (internal only — not routed through Nginx).

```go
// internal/middleware/metrics.go

import (
    "net/http"
    "strconv"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_http_requests_total",
        Help: "Total HTTP requests",
    }, []string{"method", "route", "status"})

    httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "careerdock_http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    }, []string{"method", "route"})

    httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "careerdock_http_requests_in_flight",
        Help: "Current number of HTTP requests being processed",
    })
)

func Metrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        httpRequestsInFlight.Inc()

        ww := newResponseWriter(w) // captures status code
        next.ServeHTTP(ww, r)

        httpRequestsInFlight.Dec()
        duration := time.Since(start).Seconds()

        route := chi.RouteContext(r.Context()).RoutePattern()
        if route == "" {
            route = "unknown"
        }
        status := strconv.Itoa(ww.statusCode)

        httpRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
        httpRequestDuration.WithLabelValues(r.Method, route).Observe(duration)
    })
}
```

**Metrics endpoint (internal, not exposed through Nginx):**

```go
// cmd/api/main.go — add internal metrics server

import "github.com/prometheus/client_golang/prometheus/promhttp"

// Internal metrics server on :9090 (not proxied by Nginx)
go func() {
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":9090", mux)
}()
```

### 2.2 Database Metrics

```go
// internal/middleware/db_metrics.go

var (
    dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "careerdock_db_query_duration_seconds",
        Help:    "Database query duration",
        Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
    }, []string{"operation"}) // operation: select, insert, update, delete

    dbConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "careerdock_db_connections_active",
        Help: "Active database connections",
    })

    dbConnectionsIdle = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "careerdock_db_connections_idle",
        Help: "Idle database connections",
    })

    dbQueryErrors = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_db_query_errors_total",
        Help: "Total database query errors",
    }, []string{"operation"})
)
```

pgxpool exposes connection pool stats — poll them every 15 seconds:

```go
// cmd/api/main.go — pool stats collector

go func() {
    ticker := time.NewTicker(15 * time.Second)
    for range ticker.C {
        stats := dbPool.Stat()
        dbConnectionsActive.Set(float64(stats.AcquiredConns()))
        dbConnectionsIdle.Set(float64(stats.IdleConns()))
    }
}()
```

### 2.3 Redis Metrics

```go
var (
    redisOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "careerdock_redis_operation_duration_seconds",
        Help:    "Redis operation duration",
        Buckets: []float64{0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
    }, []string{"operation"}) // get, set, del, etc.

    redisOperationErrors = promauto.NewCounter(prometheus.CounterOpts{
        Name: "careerdock_redis_operation_errors_total",
        Help: "Total Redis operation errors",
    })

    redisMemoryUsed = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "careerdock_redis_memory_used_bytes",
        Help: "Redis memory usage in bytes",
    })
)
```

Poll Redis `INFO memory` every 30 seconds for memory metrics.

### 2.4 AI Provider Metrics

```go
var (
    aiRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "careerdock_ai_request_duration_seconds",
        Help:    "AI provider request duration",
        Buckets: []float64{1, 2.5, 5, 10, 15, 30, 60},
    }, []string{"provider", "operation"}) // provider: claude/openai, operation: resume_parse/ats_general/etc.

    aiRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_ai_requests_total",
        Help: "Total AI provider requests",
    }, []string{"provider", "operation", "status"}) // status: success/error/fallback

    aiTokensUsed = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_ai_tokens_used_total",
        Help: "Total tokens consumed",
    }, []string{"provider", "operation", "direction"}) // direction: input/output

    aiCacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_ai_cache_total",
        Help: "AI result cache hits and misses",
    }, []string{"result"}) // result: hit/miss
)
```

### 2.5 Asynq Worker Metrics

```go
var (
    jobProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_job_processed_total",
        Help: "Total jobs processed",
    }, []string{"task_type", "status"}) // status: success/error

    jobProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "careerdock_job_processing_duration_seconds",
        Help:    "Job processing duration",
        Buckets: []float64{0.5, 1, 2.5, 5, 10, 30, 60, 120},
    }, []string{"task_type"})

    jobQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "careerdock_job_queue_depth",
        Help: "Current job queue depth",
    }, []string{"queue"}) // queue: critical/default/low
)
```

Poll Asynq inspector every 30 seconds for queue depth:

```go
go func() {
    inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: cfg.RedisURL})
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        for _, queue := range []string{"critical", "default", "low"} {
            info, err := inspector.GetQueueInfo(queue)
            if err == nil {
                jobQueueDepth.WithLabelValues(queue).Set(float64(info.Pending))
            }
        }
    }
}()
```

### 2.6 Business Metrics

```go
var (
    userSignupsTotal = promauto.NewCounter(prometheus.CounterOpts{
        Name: "careerdock_user_signups_total",
        Help: "Total user registrations",
    })

    premiumConversionsTotal = promauto.NewCounter(prometheus.CounterOpts{
        Name: "careerdock_premium_conversions_total",
        Help: "Total free-to-premium conversions",
    })

    paymentRevenuePaise = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_payment_revenue_paise_total",
        Help: "Total revenue in paise",
    }, []string{"product_type"})

    resumeUploadsTotal = promauto.NewCounter(prometheus.CounterOpts{
        Name: "careerdock_resume_uploads_total",
        Help: "Total resume uploads",
    })

    atsChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "careerdock_ats_checks_total",
        Help: "Total ATS checks performed",
    }, []string{"check_type"}) // general/company/job
)
```

These counters are incremented in the relevant service methods (e.g., `userSignupsTotal.Inc()` in `AuthService.Register()`).

---

## 3. Metric Collection — Grafana Cloud

### 3.1 Why Grafana Cloud (Not Self-Hosted Prometheus)

| Option | Pros | Cons |
|--------|------|------|
| Self-hosted Prometheus + Grafana | Full control | Needs EC2 resources, disk, maintenance |
| **Grafana Cloud (free tier)** | Zero maintenance, 10K metrics, 14-day retention | Limited series count |
| AWS CloudWatch Metrics | Native AWS integration | Expensive per custom metric |

Grafana Cloud free tier is the best fit: zero ops overhead, enough capacity for MVP, and natively supports Prometheus metrics.

### 3.2 Setup — Grafana Alloy (Agent)

Grafana Alloy (formerly Grafana Agent) runs on EC2 and scrapes the `/metrics` endpoint, then remote-writes to Grafana Cloud.

```yaml
# /opt/careerdock/alloy-config.yaml

prometheus.scrape "api" {
  targets = [
    { "__address__" = "127.0.0.1:9090", "instance" = "api" },
  ]
  scrape_interval = "30s"
  forward_to     = [prometheus.remote_write.grafana_cloud.receiver]
}

prometheus.remote_write "grafana_cloud" {
  endpoint {
    url = "https://prometheus-prod-01-ap-south-1.grafana.net/api/prom/push"
    basic_auth {
      username = "<grafana-cloud-user-id>"
      password = "<grafana-cloud-api-key>"
    }
  }
}
```

**Docker Compose addition** (added to `docker-compose.prod.yml`):

```yaml
  alloy:
    image: grafana/alloy:latest
    restart: unless-stopped
    volumes:
      - ./alloy-config.yaml:/etc/alloy/config.alloy:ro
    network_mode: host    # Access API metrics on localhost:9090
    deploy:
      resources:
        limits:
          memory: 128M
          cpus: '0.25'
```

### 3.3 Metric Budget

With 10K active series on Grafana Cloud free tier, here's the allocation:

| Category | Estimated Series | Key Metrics |
|----------|:---:|-------------|
| HTTP requests | ~200 | method × route × status combinations |
| HTTP duration | ~100 | method × route × bucket combinations |
| Database | ~30 | query duration, connections, errors |
| Redis | ~20 | operation duration, memory, errors |
| AI provider | ~80 | provider × operation × status/direction |
| Job queue | ~30 | task_type × status, queue depth |
| Business | ~20 | signups, conversions, revenue, uploads, ATS checks |
| Go runtime | ~50 | goroutines, GC, memory (auto from promhttp) |
| **Total** | **~530** | Well within 10K limit |

---

## 4. Error Tracking — Sentry

### 4.1 Backend Integration

```go
// cmd/api/main.go

import "github.com/getsentry/sentry-go"

func main() {
    if err := sentry.Init(sentry.ClientOptions{
        Dsn:              cfg.SentryDSN,
        Environment:      cfg.Environment,
        Release:          buildVersion,     // e.g., "v1.2.0"
        TracesSampleRate: 0.1,              // 10% of transactions for performance monitoring
        AttachStacktrace: true,
    }); err != nil {
        slog.Error("sentry init failed", "error", err)
    }
    defer sentry.Flush(2 * time.Second)

    // ...
}
```

### 4.2 Sentry Middleware

```go
// internal/middleware/sentry.go

import (
    "github.com/getsentry/sentry-go"
    sentryhttp "github.com/getsentry/sentry-go/http"
)

func SentryMiddleware() func(http.Handler) http.Handler {
    handler := sentryhttp.New(sentryhttp.Options{
        Repanic: true,
    })
    return handler.Handle
}

// Add user context after auth middleware
func SentryUserContext(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if userID := middleware.UserIDFromContext(r.Context()); userID != uuid.Nil {
            hub := sentry.GetHubFromContext(r.Context())
            if hub != nil {
                hub.Scope().SetUser(sentry.User{ID: userID.String()})
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

### 4.3 Worker Sentry Integration

```go
// cmd/worker/main.go

// Wrap each task handler to capture errors
func withSentry(handler asynq.Handler) asynq.Handler {
    return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
        defer func() {
            if r := recover(); r != nil {
                sentry.CurrentHub().Recover(r)
                sentry.Flush(2 * time.Second)
                panic(r) // Re-panic for Asynq retry logic
            }
        }()

        err := handler.ProcessTask(ctx, t)
        if err != nil {
            sentry.CaptureException(err)
        }
        return err
    })
}
```

### 4.4 Frontend Sentry

```typescript
// frontend/src/app/layout.tsx

import * as Sentry from '@sentry/nextjs';

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,
  environment: process.env.NODE_ENV,
  release: process.env.NEXT_PUBLIC_APP_VERSION,
  tracesSampleRate: 0.1,
  replaysSessionSampleRate: 0,      // Disable session replay (privacy)
  replaysOnErrorSampleRate: 0,
});
```

**Sentry Next.js config** (`sentry.client.config.ts`, `sentry.server.config.ts`, `sentry.edge.config.ts`) auto-captures:
- Unhandled exceptions
- Unhandled promise rejections
- Navigation errors
- API call failures

### 4.5 Sentry Budget (5K Events/Month)

| Source | Estimated Events/Month | Notes |
|--------|:---:|-------|
| Backend errors | ~500 | 5xx responses, LLM failures, DB errors |
| Worker errors | ~200 | Job failures (retried 3x each) |
| Frontend errors | ~300 | JS errors, API call failures |
| Performance transactions (10%) | ~2,000 | Sampled API + frontend traces |
| **Total** | **~3,000** | Within 5K limit |

If approaching the limit, reduce `TracesSampleRate` to 0.05 (5%).

---

## 5. Alerting Rules

### 5.1 Alert Definitions

| Alert | Severity | Condition | Channel | Action |
|-------|----------|-----------|---------|--------|
| **API down** | P1 | UptimeRobot: `/api/health` fails 2 consecutive checks | Email + push | Investigate immediately. Check EC2, Docker, Nginx |
| **Frontend down** | P1 | UptimeRobot: frontend fails 2 consecutive checks | Email + push | Check Vercel dashboard |
| **Error rate spike** | P1 | `rate(careerdock_http_requests_total{status=~"5.."}[5m]) / rate(careerdock_http_requests_total[5m]) > 0.05` | Grafana → email | Check Sentry for new errors. Review CloudWatch logs |
| **Database down** | P1 | `/api/health` returns `checks.database = "unhealthy"` | UptimeRobot | Check RDS console. Verify security group. Check connections |
| **High error rate** | P2 | 5xx rate > 2% over 15 minutes | Grafana → email | Review Sentry. Check for deployment issues |
| **High latency** | P2 | `histogram_quantile(0.95, careerdock_http_request_duration_seconds) > 2` for 10 min | Grafana → email | Check slow queries. Review AI provider latency |
| **AI cost spike** | P2 | Daily cost > ₹500 (feature flag threshold) | Asynq task → email | Review AI operations. Check for retry loops. Consider OpenAI fallback |
| **Queue backup** | P2 | `careerdock_job_queue_depth > 50` for 15 min | Grafana → email | Check worker process. Review failed jobs |
| **Redis memory high** | P2 | `careerdock_redis_memory_used_bytes > 200MB` (80% of 256MB) | Grafana → email | Review eviction policy. Clear stale cache entries |
| **DB connections high** | P2 | `careerdock_db_connections_active > 15` (75% of pool) | Grafana → email | Check for connection leaks. Review slow queries |
| **AI cost critical** | P2 | Daily cost > ₹1,000 | Asynq task → email | Consider auto-disabling AI via feature flag |
| **Failed jobs spike** | P3 | `increase(careerdock_job_processed_total{status="error"}[1h]) > 10` | Grafana → email | Review failed job types. Check for provider outages |
| **Payment stuck** | P3 | Payments in `created` status > 30 min | Asynq task → email | Check Razorpay dashboard. Manual reconciliation |

### 5.2 Grafana Alert Configuration

```yaml
# Grafana alerting rule (configured via Grafana Cloud UI)
# Example: Error rate spike

name: "API Error Rate > 5%"
condition:
  query: |
    rate(careerdock_http_requests_total{status=~"5.."}[5m])
    /
    rate(careerdock_http_requests_total[5m])
    > 0.05
  for: 5m
labels:
  severity: P1
annotations:
  summary: "API error rate is {{ $value | humanizePercentage }}"
  runbook: "Check Sentry for new errors. Review CloudWatch logs."
```

### 5.3 Custom Alert Tasks (Asynq Periodic)

For alerts that require business logic (not pure metric thresholds):

```go
// internal/worker/task_alert_check.go

// Runs every hour via Asynq scheduler
func (w *AlertWorker) CheckAlerts(ctx context.Context) error {
    // 1. AI cost check
    todayCost, _ := w.adminService.GetTodayAICostPaise(ctx)
    warnThreshold, _ := w.featureFlags.GetInt(ctx, "ai_cost_warn_threshold_paise")
    critThreshold, _ := w.featureFlags.GetInt(ctx, "ai_cost_critical_threshold_paise")

    if todayCost > critThreshold {
        w.email.Send(ctx, adminAlertEmail("AI cost critical", fmt.Sprintf("Today's AI cost: ₹%.2f", float64(todayCost)/100)))
        w.logger.Error("AI cost critical threshold breached", "cost_paise", todayCost)
    } else if todayCost > warnThreshold {
        w.email.Send(ctx, adminAlertEmail("AI cost warning", fmt.Sprintf("Today's AI cost: ₹%.2f", float64(todayCost)/100)))
        w.logger.Warn("AI cost warn threshold breached", "cost_paise", todayCost)
    }

    // 2. Stuck payments check
    stuckPayments, _ := w.paymentRepo.CountStuckPayments(ctx, 30*time.Minute)
    if stuckPayments > 0 {
        w.email.Send(ctx, adminAlertEmail("Stuck payments", fmt.Sprintf("%d payments in 'created' status for 30+ min", stuckPayments)))
    }

    // 3. Failed jobs check
    failedJobs, _ := w.adminService.GetFailedJobCount(ctx, 24*time.Hour)
    if failedJobs > 10 {
        w.email.Send(ctx, adminAlertEmail("Failed jobs spike", fmt.Sprintf("%d failed jobs in last 24h", failedJobs)))
    }

    return nil
}
```

---

## 6. Service Level Objectives (SLOs)

MVP SLOs — informal targets, not contractual. Formalize after reaching 1K+ users.

| SLO | Target | Measurement | Acceptable Downtime |
|-----|--------|-------------|-------------------|
| **API availability** | 99% | UptimeRobot success rate | ~7.3 hours/month |
| **API latency (p95)** | < 1 second | `histogram_quantile(0.95, ...)` | Excludes AI-triggered async ops |
| **API latency (p99)** | < 5 seconds | `histogram_quantile(0.99, ...)` | Includes file uploads |
| **Error rate** | < 1% | 5xx / total requests | Measured over 24h rolling window |
| **Resume processing** | 99% success | Successful / total parse+score jobs | Retried 3x via Asynq |
| **Payment webhook** | 99.9% success | Captured payments / total webhook deliveries | Idempotent, auto-retried by Razorpay |

### 6.1 SLO Monitoring

Track SLO compliance using Prometheus recording rules:

```yaml
# Recording rules (Grafana Cloud)
groups:
  - name: slo
    interval: 5m
    rules:
      - record: careerdock:api_availability:ratio_5m
        expr: 1 - (rate(careerdock_http_requests_total{status=~"5.."}[5m]) / rate(careerdock_http_requests_total[5m]))

      - record: careerdock:api_latency_p95:seconds_5m
        expr: histogram_quantile(0.95, rate(careerdock_http_request_duration_seconds_bucket[5m]))

      - record: careerdock:api_latency_p99:seconds_5m
        expr: histogram_quantile(0.99, rate(careerdock_http_request_duration_seconds_bucket[5m]))
```

---

## 7. Grafana Dashboards

### 7.1 Overview Dashboard

**Panels:**

| Panel | Type | Query |
|-------|------|-------|
| Request rate | Stat | `sum(rate(careerdock_http_requests_total[5m]))` |
| Error rate | Gauge | `rate(5xx) / rate(total)` — green < 1%, yellow < 5%, red > 5% |
| P95 latency | Stat | `histogram_quantile(0.95, rate(...[5m]))` |
| Requests in flight | Gauge | `careerdock_http_requests_in_flight` |
| Request rate by status | Time series | `sum by (status) (rate(careerdock_http_requests_total[5m]))` |
| Latency distribution | Heatmap | `rate(careerdock_http_request_duration_seconds_bucket[5m])` |
| Top endpoints by latency | Table | P95 latency by route, sorted descending |
| Error count by route | Table | `sum by (route) (rate(careerdock_http_requests_total{status=~"5.."}[5m]))` |

### 7.2 Infrastructure Dashboard

| Panel | Type | Query |
|-------|------|-------|
| DB active connections | Time series | `careerdock_db_connections_active` |
| DB idle connections | Time series | `careerdock_db_connections_idle` |
| DB query latency (p95) | Stat | `histogram_quantile(0.95, rate(careerdock_db_query_duration_seconds_bucket[5m]))` |
| DB errors | Time series | `rate(careerdock_db_query_errors_total[5m])` |
| Redis memory used | Gauge | `careerdock_redis_memory_used_bytes` — warn at 200MB, crit at 240MB |
| Redis operation latency | Time series | `histogram_quantile(0.95, rate(careerdock_redis_operation_duration_seconds_bucket[5m]))` |
| Queue depth | Time series | `careerdock_job_queue_depth` by queue |
| Go goroutines | Time series | `go_goroutines` |
| Go memory | Time series | `go_memstats_alloc_bytes` |

### 7.3 AI Operations Dashboard

| Panel | Type | Query |
|-------|------|-------|
| AI requests (rate) | Time series | `sum by (operation) (rate(careerdock_ai_requests_total[5m]))` |
| AI latency by operation | Time series | `histogram_quantile(0.95, rate(careerdock_ai_request_duration_seconds_bucket[5m]))` by operation |
| AI errors | Time series | `rate(careerdock_ai_requests_total{status="error"}[5m])` by provider |
| Token usage | Time series | `rate(careerdock_ai_tokens_used_total[1h])` by provider + direction |
| Cache hit rate | Stat | `rate(careerdock_ai_cache_total{result="hit"}[1h]) / rate(careerdock_ai_cache_total[1h])` |
| Fallback rate | Stat | `rate(careerdock_ai_requests_total{status="fallback"}[1h]) / rate(careerdock_ai_requests_total[1h])` |

### 7.4 Business Dashboard

| Panel | Type | Query |
|-------|------|-------|
| Daily signups | Stat | `increase(careerdock_user_signups_total[24h])` |
| Premium conversions | Stat | `increase(careerdock_premium_conversions_total[24h])` |
| Revenue today (₹) | Stat | `increase(careerdock_payment_revenue_paise_total[24h]) / 100` |
| Revenue trend | Time series | `increase(careerdock_payment_revenue_paise_total[24h])` over 30 days |
| Resume uploads | Stat | `increase(careerdock_resume_uploads_total[24h])` |
| ATS checks by type | Time series | `rate(careerdock_ats_checks_total[1h])` by check_type |

---

## 8. CloudWatch Log Queries

Common CloudWatch Insights queries for debugging:

### 8.1 Recent Errors

```
fields @timestamp, msg, level, request_id, path, status, error
| filter level = "ERROR"
| sort @timestamp desc
| limit 50
```

### 8.2 Slow Requests (>1s)

```
fields @timestamp, method, path, status, duration_ms, user_id
| filter duration_ms > 1000
| sort duration_ms desc
| limit 50
```

### 8.3 Errors by Endpoint

```
fields path, status
| filter status >= 500
| stats count() as error_count by path
| sort error_count desc
```

### 8.4 AI Operation Logs

```
fields @timestamp, msg, operation, provider, duration_ms, tokens_used
| filter msg like /ai_operation/
| sort @timestamp desc
| limit 50
```

### 8.5 Payment Webhook Issues

```
fields @timestamp, msg, razorpay_order_id, razorpay_payment_id, status, error
| filter msg like /webhook/ or msg like /payment/
| filter level in ["ERROR", "WARN"]
| sort @timestamp desc
| limit 50
```

### 8.6 Request Trace (by Request ID)

```
fields @timestamp, level, msg, request_id, path, status, duration_ms, user_id
| filter request_id = "abc-123-def"
| sort @timestamp asc
```

---

## 9. Structured Log Events

Standard log events emitted by the application (for consistent querying):

| Event (`msg`) | Level | Key Fields | Emitted By |
|---------------|-------|-----------|------------|
| `request_completed` | INFO | method, path, status, duration_ms, user_id, request_id | HTTP middleware |
| `request_failed` | ERROR | method, path, status, duration_ms, error, request_id | HTTP middleware |
| `auth_login_success` | INFO | user_id, email | Auth service |
| `auth_login_failed` | WARN | email, reason (invalid_password / account_not_found) | Auth service |
| `auth_lockout` | WARN | email, ip, attempts | Brute-force middleware |
| `resume_uploaded` | INFO | user_id, resume_id, size_bytes | Resume handler |
| `resume_parsed` | INFO | resume_id, duration_ms, tokens_used | Worker |
| `ats_check_completed` | INFO | check_type, resume_id, score, duration_ms, tokens_used, provider | Worker |
| `ats_check_failed` | ERROR | check_type, resume_id, error, provider | Worker |
| `ai_operation_completed` | INFO | operation, provider, duration_ms, input_tokens, output_tokens, cached | AI service |
| `ai_operation_failed` | ERROR | operation, provider, error, fallback_attempted | AI service |
| `ai_fallback_triggered` | WARN | operation, primary_provider, fallback_provider, primary_error | AI service |
| `payment_order_created` | INFO | user_id, order_id, product_type, amount_paise | Payment service |
| `payment_captured` | INFO | user_id, order_id, payment_id, amount_paise, credits_allocated | Payment webhook |
| `payment_failed` | WARN | user_id, order_id, error | Payment webhook |
| `refund_issued` | INFO | admin_id, payment_id, amount_paise, reason | Admin service |
| `job_completed` | INFO | task_type, duration_ms | Worker |
| `job_failed` | ERROR | task_type, error, retry_count | Worker |
| `feature_flag_toggled` | INFO | admin_id, flag_key, enabled | Admin service |
| `user_suspended` | INFO | admin_id, user_id | Admin service |

---

## 10. Sprint 0 Monitoring Checklist

### Logging
- [ ] CloudWatch log groups created (`careerdock-api`, `careerdock-worker`)
- [ ] Retention set to 30 days
- [ ] Docker awslogs driver configured in `docker-compose.prod.yml`
- [ ] Log redaction middleware active (never logs passwords, tokens, etc.)
- [ ] Request ID middleware generates and propagates correlation IDs

### Metrics
- [ ] Prometheus metrics exposed on `:9090/metrics` (internal only)
- [ ] Grafana Cloud account created (free tier)
- [ ] Grafana Alloy agent running on EC2
- [ ] HTTP, DB, Redis, AI, job queue, and business metrics emitting
- [ ] Overview, Infrastructure, AI, and Business dashboards created

### Error Tracking
- [ ] Sentry project created (backend + frontend)
- [ ] Go Sentry SDK initialised with environment + release tags
- [ ] Frontend Sentry initialised with source maps
- [ ] Worker tasks wrapped with Sentry error capture
- [ ] User context attached after auth (user_id only, no PII)

### Uptime
- [ ] UptimeRobot monitor: API health endpoint (every 5 min)
- [ ] UptimeRobot monitor: Frontend (every 5 min)
- [ ] Alert notifications configured (email)

### Alerting
- [ ] Grafana alerts: error rate, latency, DB connections, Redis memory, queue depth
- [ ] Asynq periodic task: AI cost check (hourly)
- [ ] Asynq periodic task: stuck payments check (hourly)
- [ ] Cost alert feature flags seeded (`ai_cost_warn_threshold_paise`, `ai_cost_critical_threshold_paise`)

### Health Checks
- [ ] `/api/health` returns status + version + component checks
- [ ] Build version injected via `-ldflags` in Docker build
- [ ] Docker Compose health checks active for API and Redis

---

## 11. Deferred to v2

| Feature | Reason |
|---------|--------|
| Distributed tracing (OpenTelemetry) | Overkill for single-server MVP. Add when moving to ECS multi-container |
| Log-based metrics (CloudWatch Metric Filters) | Grafana Cloud metrics are simpler and cheaper |
| APM (Application Performance Monitoring) | Sentry performance monitoring (10% sample) is sufficient for now |
| Custom Grafana alerting (PagerDuty/Slack) | Email alerts sufficient for solo founder |
| Real-time dashboards (WebSocket) | Admin panel with polling is sufficient |
| User behaviour analytics (PostHog/Mixpanel) | Integrate when product-market fit requires funnel analysis |
| Synthetic monitoring (Checkly/Datadog) | UptimeRobot health checks are sufficient |
| SLO burn-rate alerting | Formal SLO enforcement when user count warrants it |
