package models

import "time"

type LogResponse struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Action       string    `json:"action"`
	Details      *string   `json:"details"`
	ServiceGroup string    `json:"service_group"`
	Level        string    `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
}

type LogRequestData struct {
	RequestId string        `json:"Request_id"`
	UserId    int64         `json:"User_id"`
	Method    string        `json:"Method"`
	Path      string        `json:"Path"`
	Uri       string        `json:"Uri"`
	Status    int           `json:"Status"`
	Latency   time.Duration `json:"Latency"`
	RemoteIp  string        `json:"Remote_ip"`
	UserAgent string        `json:"User_agent"`
	Protocol  string        `json:"Protocol"`
	Bytes     int64         `json:"Bytes"`
	Error     any           `json:"Error"`
}
type GetLogsResponse struct {
	Logs       []LogResponse `json:"logs"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int64         `json:"total_pages"`
}

type LogStatsResponse struct {
	Total         int64     `json:"total"`
	Days          int       `json:"days"`
	ServiceGroups []string  `json:"service_groups"`
	Levels        []string  `json:"levels"`
	Since         time.Time `json:"since"`
}

type ClearOldLogsResponse struct {
	RetentionDays int       `json:"retention_days"`
	Before        time.Time `json:"before"`
	Message       string    `json:"message"`
}

// ============ Metrics Response Structs ============

type MetricsPeriod struct {
	Days  int       `json:"days"`
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

type ErrorRateByService struct {
	ServiceGroup string  `json:"service_group"`
	ErrorCount   int64   `json:"error_count"`
	TotalCount   int64   `json:"total_count"`
	ErrorRate    float64 `json:"error_rate"`
}

type LogCountByHour struct {
	Hour     string `json:"hour"`
	LogCount int64  `json:"log_count"`
}

type LogLevelStats struct {
	Level      string  `json:"level"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type ServiceStats struct {
	ServiceGroup string  `json:"service_group"`
	Count        int64   `json:"count"`
	Percentage   float64 `json:"percentage"`
}

type MetricsResponse struct {
	ErrorRateByService       []ErrorRateByService `json:"error_rate_by_service"`
	LogCountByHour           []LogCountByHour     `json:"log_count_by_hour"`
	LogLevelDistribution     []LogLevelStats      `json:"log_level_distribution"`
	ServiceGroupDistribution []ServiceStats       `json:"service_group_distribution"`
	Period                   MetricsPeriod        `json:"period"`
}

type ServiceMetricsResponse struct {
	ServiceGroup      string          `json:"service_group"`
	Days              int             `json:"days"`
	Since             time.Time       `json:"since"`
	TotalLogs         int64           `json:"total_logs"`
	ErrorCount        int64           `json:"error_count"`
	ErrorRate         float64         `json:"error_rate"`
	LevelDistribution []LogLevelStats `json:"level_distribution"`
}

type ServiceHealth struct {
	ServiceGroup string  `json:"service_group"`
	ErrorRate    float64 `json:"error_rate"`
	ErrorCount   int64   `json:"error_count"`
	TotalCount   int64   `json:"total_count"`
	Status       string  `json:"status"`
}

type HealthMetricsResponse struct {
	Timestamp     time.Time       `json:"timestamp"`
	Period        string          `json:"period"`
	Services      []ServiceHealth `json:"services"`
	OverallStatus string          `json:"overall_status"`
	Alerts        []string        `json:"alerts"`
}
