package models

import "time"

// SystemMetrics represents the system metrics collected at a specific time
type SystemMetrics struct {
	Timestamp time.Time                  `json:"timestamp" gorm:"column:timestamp;index;not null"`
	DeviceID  string                     `json:"device_id" gorm:"column:device_id;size:255;index;not null"`
	CPUUsage  *float64                   `json:"cpu_usage,omitempty" gorm:"column:cpu_usage"`
	Memory    *float64                   `json:"memory,omitempty" gorm:"column:memory"`
	Disk      *float64                   `json:"disk,omitempty" gorm:"column:disk"`
	Network   *float64                   `json:"network,omitempty" gorm:"column:network"`
	Processes map[string]*ProcessMetrics `json:"processes,omitempty" gorm:"-"`
}
