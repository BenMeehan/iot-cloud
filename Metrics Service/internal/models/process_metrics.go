package models

import "time"

// ProcessMetrics contains metrics for an individual process
type ProcessMetrics struct {
	Timestamp   time.Time `json:"timestamp" gorm:"column:timestamp;index;not null"`
	DeviceID    string    `json:"device_id" gorm:"column:device_id;size:255;index;not null"`
	ProcessName string    `json:"process_name,omitempty" gorm:"column:process_name"`
	CPUUsage    *float64  `json:"cpu_usage,omitempty" gorm:"column:cpu_usage"`
	Memory      *float64  `json:"memory,omitempty" gorm:"column:memory"`
}
