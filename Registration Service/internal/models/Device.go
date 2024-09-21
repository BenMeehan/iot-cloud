package models

import (
	"gorm.io/gorm"
)

// Device represents a device registered in the system
type Device struct {
	gorm.Model        // Embedding gorm.Model to include CreatedAt, UpdatedAt, DeletedAt
	ID         string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	// Additional fields can be added here in the future
}

// NewDevice creates a new Device instance with a generated UUID
func NewDevice(deviceID string) *Device {
	return &Device{
		ID: deviceID,
	}
}
