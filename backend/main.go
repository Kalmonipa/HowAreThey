package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	cron "github.com/robfig/cron/v3"

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
		notes TEXT NOT NULL
    );`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

// GetRandomFriendScheduled is used for scheduled calls, without a Gin context
func GetRandomFriendScheduled() {
	// You can either make an internal call to the existing Gin handler
	// Or extract the logic into a separate function and call it here
	// ...

	// Example of making an HTTP request to the endpoint
	resp, err := http.Get("http://localhost:8080/friends/random")
	if err != nil {
		log.Println("Error calling GetRandomFriend:", err)
		return
	}
	defer resp.Body.Close()
	// Handle the response as needed
}

func main() {

	var dbFilePath = "sql/friends.db"

	// Sets up the logger
	logger.SetupLogger()

	// Open the database connection
	db, err := createOrOpenSQLiteDB(dbFilePath)
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to open database: %v", err)
	}
	defer db.Close()

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

	// Create an instance of your handler struct with the friendsList
	friendsHandler := handler.NewFriendsHandler(friendsList, db)

	router := handler.SetupRouter(friendsHandler)

	// Initialize the cron scheduler
	c := cron.New
	// Schedule your task: "@weekly" runs once every week
	_, err = c.AddFunc("@weekly", func() {
		// Call your GetRandomFriend function
		// You might need to adjust this call based on your implementation
		GetRandomFriendScheduled()
	})
	if err != nil {
		log.Fatal("Error scheduling the task: ", err)
	}

	// Start the cron scheduler
	c.Start()

	err = router.Run()
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "error: %v", err)
		panic(err)
	}
}
