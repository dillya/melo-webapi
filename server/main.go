package main

import (
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	// Internal
	"github.com/dillya/melo-webapi/internal/device"
	"github.com/dillya/melo-webapi/internal/discover_legacy"
	"github.com/dillya/melo-webapi/internal/utils"

	// REST / OpenAPI
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	// Use MySQL as database
	_ "github.com/go-sql-driver/mysql"

	// Logs
	log "github.com/sirupsen/logrus"
)

func initDatabaseTables(db *sql.DB) bool {
	// Create Version table
	if err := utils.InitializeVersionTable(db); err != nil {
		log.Errorf("failed to initialize Version table: %s", err)
		return false
	}

	// Create Device tables
	if !device.InitializeTables(db) {
		log.Error("failed to initialize Device tables")
		return false
	}

	return true
}

func main() {
	// Get URL from environment
	url := os.Getenv("MELO_WEBAPI_URL")

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

	// Try to connect to database
	for {
		err = db.Ping()
		if err == nil {
			break
		} else if !strings.Contains(err.Error(), "connection refused") {
			log.Errorf("failed to ping database: %s", err)
			return
		}

		// Retry to connect
		log.Error("failed to ping database: retry...")
		time.Sleep(10 * time.Second)
	}

	// Initialize Database tables
	if !initDatabaseTables(db) {
		log.Error("failed to initialize tables")
		return
	}

	// Setup API name / version
	api_name := "Melo Web API"
	api_version := "1.0.0"
	log.Info(api_name + " " + api_version)
	config := huma.DefaultConfig(api_name, api_version)

	// Setup main URL
	if url != "" {
		config.Servers = []*huma.Server{{URL: url}}
	}

	// Create a new router & API.
	router := chi.NewMux()

	// Setup CORS
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
	}))

	// Create API
	api := humachi.New(router, config)

	// Register Device API
	device.Register(api, db)

	// Register deprecated Discover API
	discover_legacy.Register(api, db)

	// Start the server
	http.ListenAndServe("0.0.0.0:8888", router)
}
