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
	Conn   *gorm.DB
	Logger *logrus.Logger
}

// NewDatabase initializes a new Database instance
func NewDatabase(logger *logrus.Logger) *Database {
	return &Database{
		Logger: logger,
	}
}

// Connect establishes the database connection using GORM
func (d *Database) Connect(connStr string) error {
	var err error
	d.Conn, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		d.Logger.WithError(err).Fatal("Failed to open database connection")
		return err
	}

	// Automatically create the devices table if it doesn't exist
	if err := d.Conn.AutoMigrate(&models.Device{}); err != nil {
		d.Logger.WithError(err).Fatal("Failed to auto-migrate database schema")
		return err
	}

	d.Logger.Info("Database connection established")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	// GORM handles connection pooling, so Close is not implemented like in sql.DB
	d.Logger.Info("Database connection closed (handled by GORM)")
	return nil
}

// GetConn returns the underlying GORM database connection
func (d *Database) GetConn() *gorm.DB {
	return d.Conn
}

// SaveDevice saves the device to the database
func (d *Database) SaveDevice(device *models.Device) error {
	if err := d.Conn.Create(device).Error; err != nil {
		d.Logger.WithError(err).Error("Failed to save device")
		return err
	}
	d.Logger.Infof("Device saved with ID: %s", device.ID)
	return nil
}
