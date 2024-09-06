package database

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize TimescaleDB connection
func InitDB(user, password, host, dbName, SSLMode string, port int) {
	dsn := "postgres://" + user + ":" + password +
		"@" + host + ":" +
		fmt.Sprint(port) + "/" +
		dbName +
		"?sslmode=" + SSLMode

	logrus.Info(dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("Error connecting to TimescaleDB: %v", err)
	}
	DB = db
	logrus.Info("Connected to TimescaleDB")

	// Ensure the heartbeats table exists and is a hypertable
	EnsureHeartbeatTable()
}

// EnsureHeartbeatTable ensures that the heartbeats table and hypertable exist
func EnsureHeartbeatTable() {
	// Check if the heartbeats table exists
	if !tableExists("heartbeats") {
		createHeartbeatTable()
	} else {
		logrus.Info("Table 'heartbeats' already exists")
	}

	// Check if it's already a hypertable
	if !isHypertable("heartbeats") {
		convertToHypertable()
	} else {
		logrus.Info("Table 'heartbeats' is already a hypertable")
	}
}

// Check if the table exists
func tableExists(tableName string) bool {
	var exists bool
	DB.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists)
	return exists
}

// Create the heartbeats table
func createHeartbeatTable() {
	createTableSQL := `
        CREATE TABLE heartbeats (
            device_id TEXT NOT NULL,
            timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
            status TEXT
        );`
	if err := DB.Exec(createTableSQL).Error; err != nil {
		logrus.Fatalf("Error creating heartbeats table: %v", err)
	}
	logrus.Info("Created 'heartbeats' table")
}

// Check if the table is a hypertable
func isHypertable(tableName string) bool {
	var exists bool
	DB.Raw("SELECT EXISTS (SELECT 1 FROM timescaledb_information.hypertables WHERE hypertable_name = ?)", tableName).Scan(&exists)
	return exists
}

// Convert the table to a hypertable
func convertToHypertable() {
	convertSQL := "SELECT create_hypertable('heartbeats', 'timestamp');"
	if err := DB.Exec(convertSQL).Error; err != nil {
		logrus.Fatalf("Error converting table to hypertable: %v", err)
	}
	logrus.Info("Converted 'heartbeats' table to hypertable")
}
