package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ws_active_connections",
		Help: "Number of active WebSocket connections",
	})

	MessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ws_messages_total",
		Help: "Total messages processed",
	}, []string{"room"})

	MessageLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "ws_message_latency_seconds",
		Help:    "Message processing latency",
		Buckets: prometheus.DefBuckets,
	})

	RedisPublishErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_redis_publish_errors_total",
		Help: "Total Redis publish failures",
	})
	MessagesDropped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ws_messages_dropped_total",
		Help: "Messages dropped due to full client buffers",
	})
)
