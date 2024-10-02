package database

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB interface with new methods for GORM operations
type DB interface {
	Connect(connStr string) error
	Close() error
	GetConn() *gorm.DB
	EnsureHeartbeatTable()
	createHeartbeatTable()
	isHypertable(tableName string) bool
	convertToHypertable()
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

	// Ensure the heartbeats table exists and is a hypertable
	d.EnsureHeartbeatTable()

	d.Logger.Info("Connected to TimescaleDB")
	return nil
}

// EnsureHeartbeatTable ensures that the heartbeats table and hypertable exist
func (d *Database) EnsureHeartbeatTable() {
	// Check if the heartbeats table exists
	if !d.tableExists("heartbeats") {
		d.createHeartbeatTable()
	} else {
		d.Logger.Info("Table 'heartbeats' already exists")
	}

	// Check if it's already a hypertable
	if !d.isHypertable("heartbeats") {
		d.convertToHypertable()
	} else {
		d.Logger.Info("Table 'heartbeats' is already a hypertable")
	}
}

// Check if the table exists
func (d *Database) tableExists(tableName string) bool {
	var exists bool
	d.Conn.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists)
	return exists
}

// Create the heartbeats table
func (d *Database) createHeartbeatTable() {
	createTableSQL := `
        CREATE TABLE heartbeats (
            device_id TEXT NOT NULL,
            timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            status TEXT
        );`
	if err := d.Conn.Exec(createTableSQL).Error; err != nil {
		d.Logger.Fatalf("Error creating heartbeats table: %v", err)
	}
	d.Logger.Info("Created 'heartbeats' table")
}

// Check if the table is a hypertable
func (d *Database) isHypertable(tableName string) bool {
	var exists bool
	d.Conn.Raw("SELECT EXISTS (SELECT 1 FROM timescaledb_information.hypertables WHERE hypertable_name = ?)", tableName).Scan(&exists)
	return exists
}

// Convert the table to a hypertable
func (d *Database) convertToHypertable() {
	convertSQL := "SELECT create_hypertable('heartbeats', 'timestamp');"
	if err := d.Conn.Exec(convertSQL).Error; err != nil {
		d.Logger.Fatalf("Error converting table to hypertable: %v", err)
	}
	d.Logger.Info("Converted 'heartbeats' table to hypertable")
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
