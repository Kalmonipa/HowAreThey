package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron"
	_ "github.com/robfig/cron/v3"

	"howarethey/pkg/handler"
	"howarethey/pkg/logger"
	"howarethey/pkg/models"
)

func createOrOpenSQLiteDB(sqlFilePath string) (*sql.DB, error) {
	// Check if the database file exists
	if _, err := os.Stat(sqlFilePath); os.IsNotExist(err) {
		file, err := os.Create(sqlFilePath)
		if err != nil {
			return nil, err
		}
		file.Close()
		logger.LogMessage(logger.LogLevelInfo, "Database file created")
	} else {
		logger.LogMessage(logger.LogLevelDebug, "Database file already exists")
	}

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", sqlFilePath)
	if err != nil {
		return nil, err
	}
	logger.LogMessage(logger.LogLevelDebug, "Database file opened")
	return db, nil
}

// createTable creates a table in the SQLite database to store your data.
func createTable(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS friends (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		lastContacted TEXT NOT NULL,
		birthday TEXT NOT NULL,
		notes TEXT NOT NULL
    );`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

// CheckBirthdaysToday is used daily
func CheckBirthdaysToday() {
	resp, err := http.Get("http://localhost:8080/birthdays")
	if err != nil {
		logger.LogMessage(logger.LogLevelError, "Error calling GetRandomFriend: %v", err)
		return
	}
	defer resp.Body.Close()
}

// GetRandomFriendScheduled is used for scheduled calls, without a Gin context
func GetRandomFriendScheduled() {
	// Example of making an HTTP request to the endpoint
	resp, err := http.Get("http://localhost:8080/friends/random")
	if err != nil {
		logger.LogMessage(logger.LogLevelError, "Error calling GetRandomFriend: %v", err)
		return
	}
	defer resp.Body.Close()
}

func main() {

	var (
		dbFilePath               = "sql/friends.db"
		friend_selector_schedule string
		bday_check_time          string
		bday_schedule            string
	)

	// Sets up the logger
	logger.SetupLogger()

	logger.LogMessage(logger.LogLevelInfo, "Starting app")

	// Open the database connection
	db, err := createOrOpenSQLiteDB(dbFilePath)
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to open database: %v", err)
		panic(err)
	}
	defer db.Close()

	logger.LogMessage(logger.LogLevelInfo, "Database opened")

	// Create the table
	if err := createTable(db); err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to create table: %v", err)
		panic(err)
	}

	friendsList, err := models.BuildFriendsList(db)
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to build slice: %v", err)
		panic(err)
	}

	friendsHandler := handler.NewFriendsHandler(friendsList, db)

	router := handler.SetupRouter(friendsHandler)

	c := cron.New()

	// Get the cron schedule for when to select friends to contact or default to @weekly if it's not passed in
	if os.Getenv("FRIEND_SELECTOR_CRON_SCHEDULE") != "" {
		friend_selector_schedule = os.Getenv("FRIEND_SELECTOR_CRON_SCHEDULE")
	} else {
		friend_selector_schedule = "@weekly"
	}

	logger.LogMessage(logger.LogLevelInfo, "Running on the schedule: %s", friend_selector_schedule)

	err = c.AddFunc(friend_selector_schedule, func() {
		GetRandomFriendScheduled()
	})
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "error: %v", err)
		panic(err)
	}

	if os.Getenv("IGNORE_BIRTHDAYS") != "true" {
		// Get the time of day to check if today is anyones birthday. Defaults to 8am.
		// Must be between 0-23
		if os.Getenv("BIRTHDAY_CHECK_TIME") != "" {
			bday_check_time = os.Getenv("BIRTHDAY_CHECK_TIME")
			bday_schedule = "0 " + bday_check_time + " * * *"
		} else {
			bday_schedule = "0 8 * * *"
		}
		err = c.AddFunc(bday_schedule, func() {
			CheckBirthdaysToday()
		})
		if err != nil {
			logger.LogMessage(logger.LogLevelFatal, "error: %v", err)
			panic(err)
		}
	}

	// Start the cron scheduler
	c.Start()

	notification_type := os.Getenv("NOTIFICATION_SERVICE")

	switch notification_type {
	case "DISCORD":
		notification_type = "Discord"
	case "NTFY":
		notification_type = "ntfy"
	case "TELEGRAM":
		notification_type = "Telegram"
	default:
		notification_type = ""
	}

	if notification_type != "" {
		logger.LogMessage(logger.LogLevelInfo, "Sending notifications to "+notification_type)
	}

	logger.LogMessage(logger.LogLevelInfo, "Starting webserver")

	err = router.Run()
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "error: %v", err)
		panic(err)
	}
}
