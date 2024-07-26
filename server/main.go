package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	// Internal
	"github.com/dillya/melo-webapi/internal/device"
	"github.com/dillya/melo-webapi/internal/discover_legacy"

	// REST / OpenAPI
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	// Use MySQL as database
	_ "github.com/go-sql-driver/mysql"

	// Logs
	log "github.com/sirupsen/logrus"
)

func initDatabaseTables(db *sql.DB) bool {
	// Create device table
	device := `CREATE TABLE IF NOT EXISTS device (
  id int(11) NOT NULL AUTO_INCREMENT,
  serial varchar(17) NOT NULL,
  ip int(10) unsigned NOT NULL,
  name varchar(128) NOT NULL,
  http_port mediumint(9) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY serial_2 (serial,ip),
  KEY serial (serial),
  KEY ip (ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;`
	_, err := db.Exec(device)
	if err != nil {
		log.Errorf("failed to create device table: %s", err)
		return false
	}

	// Create device_iface table
	device_iface := `CREATE TABLE IF NOT EXISTS device_iface (
  id int(11) NOT NULL AUTO_INCREMENT,
  device_id int(11) NOT NULL,
  ip int(10) unsigned NOT NULL,
  mac bigint(20) unsigned NOT NULL,
  name varchar(128) NOT NULL DEFAULT 'Unknown',
  type int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY device_id (device_id,mac),
  CONSTRAINT device_iface_ibfk_2 FOREIGN KEY (device_id) REFERENCES device (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;`
	_, err = db.Exec(device_iface)
	if err != nil {
		log.Errorf("failed to create device interface table: %s", err)
		return false
	}

	return true
}

func main() {
	// Get MySQL login and database from environment
	hostname := os.Getenv("MELO_WEBAPI_MYSQL_HOSTNAME")
	user := os.Getenv("MELO_WEBAPI_MYSQL_USER")
	password := os.Getenv("MELO_WEBAPI_MYSQL_PASSWORD")
	db_name := os.Getenv("MELO_WEBAPI_MYSQL_DATABASE")

	// Create SQL connection
	db, err := sql.Open("mysql", user+":"+password+"@tcp("+hostname+")/"+db_name)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to open database")
	}
	defer db.Close()

	// Setup default database connections
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Initialize Database tables
	if !initDatabaseTables(db) {
		log.Error("failed to initialize tables")
		return
	}

	// Setup API name / version
	api_name := "Melo Web API"
	api_version := "1.0.0"
	log.Info(api_name + " " + api_version)

	// Create a new router & API.
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig(api_name, api_version))

	// Register Device API
	device.Register(api)

	// Register deprecated Discover API
	discover_legacy.Register(api, db)

	// Start the server
	http.ListenAndServe("0.0.0.0:8888", router)
}
