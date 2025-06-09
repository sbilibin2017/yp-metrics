package types

// Counter represents a metric type that accumulates counts.
const (
	Counter = "counter"
	// Gauge represents a metric type that measures a value at a point in time.
	Gauge = "gauge"
)

// Metrics represents a single metric data point with its identifier, type, and value.
// Depending on the metric type, either Delta or Value is set.
// Delta is used for Counter metrics (int64).
// Value is used for Gauge metrics (float64).
type Metrics struct {
	ID    string   `json:"id"`              // Metric identifier (name)
	MType string   `json:"type"`            // Metric type: "counter" or "gauge"
	Delta *int64   `json:"delta,omitempty"` // Counter metric value (optional)
	Value *float64 `json:"value,omitempty"` // Gauge metric value (optional)
	Hash  string   `json:"hash,omitempty"`  // Optional hash for metric integrity verification
}

// MetricID represents the identifier and type of a metric, typically used for referencing a metric.
type MetricID struct {
	ID    string `json:"id"`   // Metric identifier (name)
	MType string `json:"type"` // Metric type: "counter" or "gauge"
}
