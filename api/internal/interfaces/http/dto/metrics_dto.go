package dto

import (
	"time"
)

// === Metrics DTOs ===

// MetricsQueryInput represents input for querying metrics
type MetricsQueryInput struct {
	ContainerID string `path:"container_id" doc:"Container ID to get metrics for"`
	StartTime   string `query:"start" doc:"Start time (RFC3339 or relative like -1h, -24h, -3d)"`
	EndTime     string `query:"end" doc:"End time (RFC3339 or relative)"`
	Resolution  string `query:"resolution" doc:"Data resolution: 1m, 5m, 15m, 30m, 1h, 6h, 1d"`
}

// ContainerMetricsOutput represents output for container metrics
type ContainerMetricsOutput struct {
	Body ContainerMetricsResponse
}

// ContainerMetricsResponse represents the response body
type ContainerMetricsResponse struct {
	ContainerID   string                 `json:"container_id"`
	ContainerName string                 `json:"container_name"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Resolution    string                 `json:"resolution,omitempty"`
	DataPoints    int                    `json:"data_points"`
	Metrics       []ContainerMetricPoint `json:"metrics"`
}

// ContainerMetricPoint represents a single metric data point
type ContainerMetricPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	CPUPercent     float64   `json:"cpu_percent"`
	MemoryUsage    uint64    `json:"memory_usage"`
	MemoryLimit    uint64    `json:"memory_limit"`
	MemoryPercent  float64   `json:"memory_percent"`
	NetworkRxBytes uint64    `json:"network_rx_bytes"`
	NetworkTxBytes uint64    `json:"network_tx_bytes"`
	BlockRead      uint64    `json:"block_read"`
	BlockWrite     uint64    `json:"block_write"`
	PIDs           uint64    `json:"pids"`
}

// AllContainerMetricsInput represents input for all container metrics
type AllContainerMetricsInput struct {
	StartTime  string `query:"start" doc:"Start time (RFC3339 or relative)"`
	EndTime    string `query:"end" doc:"End time (RFC3339 or relative)"`
	Resolution string `query:"resolution" doc:"Data resolution: 1m, 5m, 15m, 30m, 1h, 6h, 1d"`
}

// AllContainerMetricsOutput represents output for all container metrics
type AllContainerMetricsOutput struct {
	Body AllContainerMetricsResponse
}

// AllContainerMetricsResponse represents the response body
type AllContainerMetricsResponse struct {
	StartTime  time.Time                           `json:"start_time"`
	EndTime    time.Time                           `json:"end_time"`
	Resolution string                              `json:"resolution,omitempty"`
	Containers map[string]ContainerMetricsResponse `json:"containers"`
}

// LatestMetricsInput represents input for latest metrics
type LatestMetricsInput struct{}

// LatestMetricsOutput represents output for latest metrics
type LatestMetricsOutput struct {
	Body LatestMetricsResponse
}

// LatestMetricsResponse represents the response body
type LatestMetricsResponse struct {
	Timestamp  time.Time                       `json:"timestamp"`
	System     *SystemMetricsSummary           `json:"system,omitempty"`
	Containers map[string]ContainerMetricPoint `json:"containers"`
}

// SystemMetricsSummary represents system metrics summary
type SystemMetricsSummary struct {
	CPUCores          int    `json:"cpu_cores"`
	MemoryTotal       uint64 `json:"memory_total"`
	ContainersRunning int    `json:"containers_running"`
	ContainersPaused  int    `json:"containers_paused"`
	ContainersStopped int    `json:"containers_stopped"`
	ImagesCount       int    `json:"images_count"`
	VolumesCount      int    `json:"volumes_count"`
	NetworksCount     int    `json:"networks_count"`
}

// SystemMetricsInput represents input for system metrics
type SystemMetricsInput struct {
	StartTime string `query:"start" doc:"Start time (RFC3339 or relative)"`
	EndTime   string `query:"end" doc:"End time (RFC3339 or relative)"`
}

// SystemMetricsOutput represents output for system metrics
type SystemMetricsOutput struct {
	Body SystemMetricsResponse
}

// SystemMetricsResponse represents the response body
type SystemMetricsResponse struct {
	StartTime  time.Time           `json:"start_time"`
	EndTime    time.Time           `json:"end_time"`
	DataPoints int                 `json:"data_points"`
	Metrics    []SystemMetricPoint `json:"metrics"`
}

// SystemMetricPoint represents a single system metric data point
type SystemMetricPoint struct {
	Timestamp         time.Time `json:"timestamp"`
	CPUCores          int       `json:"cpu_cores"`
	MemoryTotal       uint64    `json:"memory_total"`
	MemoryUsed        uint64    `json:"memory_used,omitempty"`
	MemoryPercent     float64   `json:"memory_percent,omitempty"`
	ContainersRunning int       `json:"containers_running"`
	ContainersPaused  int       `json:"containers_paused"`
	ContainersStopped int       `json:"containers_stopped"`
	ImagesCount       int       `json:"images_count"`
	VolumesCount      int       `json:"volumes_count"`
	NetworksCount     int       `json:"networks_count"`
}

// === Log Query DTOs ===

// LogQueryInput represents input for querying logs
type LogQueryInput struct {
	ContainerIDs string `query:"containers" doc:"Comma-separated container IDs"`
	StartTime    string `query:"start" doc:"Start time (RFC3339 or relative)"`
	EndTime      string `query:"end" doc:"End time (RFC3339 or relative)"`
	Search       string `query:"search" doc:"Search term in log message"`
	Level        string `query:"level" doc:"Log level filter: error, warn, info, debug"`
	Limit        int    `query:"limit" doc:"Maximum number of entries to return"`
	Offset       int    `query:"offset" doc:"Offset for pagination"`
}

// LogQueryOutput represents output for log query
type LogQueryOutput struct {
	Body LogQueryResponse
}

// LogQueryResponse represents the response body
type LogQueryResponse struct {
	Query      LogQueryParams `json:"query"`
	Entries    []LogEntryDTO  `json:"entries"`
	TotalCount int            `json:"total_count"`
	HasMore    bool           `json:"has_more"`
}

// LogQueryParams represents query parameters in response
type LogQueryParams struct {
	ContainerIDs []string  `json:"container_ids,omitempty"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Search       string    `json:"search,omitempty"`
	Level        string    `json:"level,omitempty"`
	Limit        int       `json:"limit,omitempty"`
	Offset       int       `json:"offset,omitempty"`
}

// LogEntryDTO represents a log entry in response
type LogEntryDTO struct {
	Timestamp     time.Time      `json:"timestamp"`
	ContainerID   string         `json:"container_id"`
	ContainerName string         `json:"container_name"`
	Stream        string         `json:"stream"`
	Message       string         `json:"message"`
	Level         string         `json:"level,omitempty"`
	Fields        map[string]any `json:"fields,omitempty"`
}

// LogAggregationInput represents input for log aggregation
type LogAggregationInput struct {
	ContainerIDs string `query:"containers" doc:"Comma-separated container IDs"`
	StartTime    string `query:"start" doc:"Start time (RFC3339 or relative)"`
	EndTime      string `query:"end" doc:"End time (RFC3339 or relative)"`
	Interval     string `query:"interval" doc:"Aggregation interval: 1m, 5m, 15m, 30m, 1h, 6h, 12h, 1d"`
	Search       string `query:"search" doc:"Search term in log message"`
	Level        string `query:"level" doc:"Log level filter"`
	GroupBy      string `query:"group_by" doc:"Group by: level, stream, source"`
}

// LogAggregationOutput represents output for log aggregation
type LogAggregationOutput struct {
	Body LogAggregationResponse
}

// LogAggregationResponse represents the response body
type LogAggregationResponse struct {
	Query      LogQueryParams     `json:"query"`
	Interval   string             `json:"interval"`
	TotalCount int64              `json:"total_count"`
	Buckets    []LogBucketDTO     `json:"buckets"`
	GroupData  map[string][]int64 `json:"group_data,omitempty"`
}

// LogBucketDTO represents a time bucket in response
type LogBucketDTO struct {
	Timestamp time.Time        `json:"timestamp"`
	Count     int64            `json:"count"`
	ByLevel   map[string]int64 `json:"by_level,omitempty"`
	ByStream  map[string]int64 `json:"by_stream,omitempty"`
	BySource  map[string]int64 `json:"by_source,omitempty"`
}

// === Metrics Store Stats ===

// MetricsStatsInput represents input for store stats
type MetricsStatsInput struct{}

// MetricsStatsOutput represents output for store stats
type MetricsStatsOutput struct {
	Body MetricsStatsResponse
}

// MetricsStatsResponse represents the response body
type MetricsStatsResponse struct {
	ContainersTracked     int    `json:"containers_tracked"`
	ContainerMetricPoints int    `json:"container_metric_points"`
	SystemMetricPoints    int    `json:"system_metric_points"`
	LogEntries            int    `json:"log_entries"`
	RetentionPeriod       string `json:"retention_period"`
}
