package database

import (
	"github.com/benmeehan/iot-registration-service/internal/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB interface with new methods for GORM operations
type DB interface {
	Connect(connStr string) error
	Close() error
	GetConn() *gorm.DB
	SaveDevice(device *models.Device) error
}

// Database holds the connection pool to the PostgreSQL database using GORM
type Database struct {
	Conn *gorm.DB
}

// NewDatabase initializes a new Database instance
func NewDatabase() *Database {
	return &Database{}
}

// Connect establishes the database connection using GORM
func (d *Database) Connect(connStr string) error {
	var err error
	d.Conn, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		logrus.WithError(err).Fatal("Failed to open database connection")
		return err
	}

	// Automatically create the devices table if it doesn't exist
	if err := d.Conn.AutoMigrate(&models.Device{}); err != nil {
		logrus.WithError(err).Fatal("Failed to auto-migrate database schema")
		return err
	}

	logrus.Info("Database connection established")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	// GORM handles connection pooling, so Close is not implemented like in sql.DB
	logrus.Info("Database connection closed (handled by GORM)")
	return nil
}

// GetConn returns the underlying GORM database connection
func (d *Database) GetConn() *gorm.DB {
	return d.Conn
}

// SaveDevice saves the device to the database
func (d *Database) SaveDevice(device *models.Device) error {
	if err := d.Conn.Create(device).Error; err != nil {
		logrus.WithError(err).Error("Failed to save device")
		return err
	}
	logrus.Infof("Device saved with ID: %s", device.ID)
	return nil
}
