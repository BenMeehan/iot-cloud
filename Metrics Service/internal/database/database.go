package database

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB interface with methods for GORM operations related to metrics
type DB interface {
	Connect(connStr string) error
	Close() error
	GetConn() *gorm.DB
	EnsureMetricsTables()
	createMetricsTables()
	isHypertable(tableName string) bool
	convertToHypertable(tableName string)
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

	// Ensure the system_metrics and process_metrics tables exist and are hypertables
	d.EnsureMetricsTables()

	d.Logger.Info("Connected to TimescaleDB")
	return nil
}

// EnsureMetricsTables ensures that both system and process metrics tables exist and are hypertables
func (d *Database) EnsureMetricsTables() {
	tables := []string{"system_metrics", "process_metrics"}

	for _, tableName := range tables {
		// Check if the table exists
		if !d.tableExists(tableName) {
			d.createMetricsTables()
		} else {
			d.Logger.Infof("Table '%s' already exists", tableName)
		}

		// Check if it's a hypertable
		if !d.isHypertable(tableName) {
			d.convertToHypertable(tableName)
		} else {
			d.Logger.Infof("Table '%s' is already a hypertable", tableName)
		}
	}
}

// Check if the table exists
func (d *Database) tableExists(tableName string) bool {
	var exists bool
	d.Conn.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists)
	return exists
}

// Create both the system_metrics and process_metrics tables if they don't exist
func (d *Database) createMetricsTables() {
	// Create system_metrics table
	createSystemMetricsSQL := `
        CREATE TABLE IF NOT EXISTS system_metrics (
            device_id TEXT NOT NULL,
            timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            cpu_usage FLOAT,
            memory FLOAT,
            disk FLOAT,
            network FLOAT,
            PRIMARY KEY (device_id, timestamp)
        );`
	if err := d.Conn.Exec(createSystemMetricsSQL).Error; err != nil {
		d.Logger.Fatalf("Error creating system_metrics table: %v", err)
	}
	d.Logger.Info("Created 'system_metrics' table")

	// Create process_metrics table (without foreign key constraint)
	createProcessMetricsSQL := `
        CREATE TABLE IF NOT EXISTS process_metrics (
            device_id TEXT NOT NULL,
            timestamp TIMESTAMPTZ NOT NULL,
            process_name TEXT NOT NULL,
            cpu_usage FLOAT,
            memory FLOAT,
            PRIMARY KEY (device_id, timestamp, process_name)
        );`
	if err := d.Conn.Exec(createProcessMetricsSQL).Error; err != nil {
		d.Logger.Fatalf("Error creating process_metrics table: %v", err)
	}
	d.Logger.Info("Created 'process_metrics' table")
}

// Check if the table is a hypertable
func (d *Database) isHypertable(tableName string) bool {
	var exists bool
	d.Conn.Raw("SELECT EXISTS (SELECT 1 FROM timescaledb_information.hypertables WHERE hypertable_name = ?)", tableName).Scan(&exists)
	return exists
}

// Convert the table to a hypertable
func (d *Database) convertToHypertable(tableName string) {
	convertSQL := "SELECT create_hypertable(?, 'timestamp');"
	if err := d.Conn.Exec(convertSQL, tableName).Error; err != nil {
		d.Logger.Fatalf("Error converting table %s to hypertable: %v", tableName, err)
	}
	d.Logger.Infof("Converted '%s' table to hypertable", tableName)
}

// Close closes the database connection
func (d *Database) Close() error {
	d.Logger.Info("Database connection closed (handled by GORM)")
	return nil
}

// GetConn returns the underlying GORM database connection
func (d *Database) GetConn() *gorm.DB {
	return d.Conn
}
