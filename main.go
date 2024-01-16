package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
)

type Friend struct {
	ID            string `yaml:"id"`
	Name          string `yaml:"name"`
	LastContacted string `yaml:"lastContacted"`
}

type FriendsList []Friend

func createOrOpenSQLiteDB(sqlFilePath string) (*sql.DB, error) {
	// Check if the database file exists
	if _, err := os.Stat(sqlFilePath); os.IsNotExist(err) {
		file, err := os.Create(sqlFilePath)
		if err != nil {
			return nil, err
		}
		file.Close()
		LogMessage(LogLevelInfo, "Database file created")
	} else {
		LogMessage(LogLevelDebug, "Database file already exists")
	}

	// Connect to the SQLite database
	db, err := sql.Open("sqlite3", sqlFilePath)
	if err != nil {
		return nil, err
	}
	LogMessage(LogLevelDebug, "Database file opened")
	return db, nil
}

// createTable creates a table in the SQLite database to store your data.
func createTable(db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS friends (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        lastContacted TEXT NOT NULL
    );`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

// Builds the list from the yaml file specified
func buildFriendsList(db *sql.DB) (FriendsList, error) {
	// Query the database
	rows, err := db.Query("SELECT id, name, lastContacted FROM friends")
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var friends FriendsList
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.Name, &f.LastContacted); err != nil {
			LogMessage(LogLevelFatal, "error: %v", err)
			return nil, err
		}
		friends = append(friends, f)
	}

	if err := rows.Err(); err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		return nil, err
	}

	return friends, nil
}

// Calculates the weight of each name based on how many days since last contacted
// The longer the time since last contact, the higher the chance of them coming up in the selection
func calculateWeight(lastContacted string, currDate time.Time) (int, error) {
	layout := "02/01/2006" // Go layout string (use the reference date)

	lastContactedDate, err := time.Parse(layout, lastContacted)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return 0, err
	}

	// Normalize both dates to the start of the day
	lastContactedNormalised := time.Date(lastContactedDate.Year(), lastContactedDate.Month(), lastContactedDate.Day(), 0, 0, 0, 0, time.Local)
	currDateNormalised := time.Date(currDate.Year(), currDate.Month(), currDate.Day(), 0, 0, 0, 0, currDate.Location())

	// Check if lastContacted date is in the future
	if lastContactedNormalised.After(currDateNormalised) {
		return 0, fmt.Errorf("lastContacted date %s is in the future. It must be in the past", lastContacted)
	}

	// Calculate the difference in days
	difference := currDateNormalised.Sub(lastContactedNormalised)
	days := int(difference.Hours() / 24)

	return days, nil
}

func pickRandomFriend(friends FriendsList) (Friend, error) {
	totalWeight := 0
	weights := make([]int, len(friends))

	for i, friend := range friends {

		weight, err := calculateWeight(friend.LastContacted, time.Now())
		if err != nil {
			return Friend{}, err
		}

		if weight <= 0 {
			continue
		}
		weights[i] = weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return Friend{}, errors.New("total weight is zero")
	}

	randIndex := rand.Intn(totalWeight)
	for i, weight := range weights {
		if randIndex < weight {
			return friends[i], nil
		}
		randIndex -= weight
	}

	return Friend{}, errors.New("unable to select a random friend")
}

func updateLastContacted(friend Friend, todaysDate time.Time) Friend {
	friend.LastContacted = todaysDate.Format("02/01/2006")

	return friend
}

func getFriendCount(friends FriendsList) int {
	return len(friends)
}

func getFriendByID(id string, friends FriendsList) (*Friend, error) {
	for _, friend := range friends {
		if friend.ID == id {
			return &friend, nil
		}
	}
	return nil, errors.New("friend not found")
}

func getFriendByName(name string, friends FriendsList) (*Friend, error) {
	for _, friend := range friends {
		lowercased := strings.ToLower(friend.Name)
		hyphenated := strings.ReplaceAll(lowercased, " ", "-")
		if hyphenated == name {
			return &friend, nil
		}
	}
	return nil, errors.New("friend not found")
}

// addFriend inserts a new friend into the database
func addFriend(db *sql.DB, newFriend Friend) error {
	stmt, err := db.Prepare("INSERT INTO friends(id, name, lastContacted) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newFriend.ID, newFriend.Name, newFriend.LastContacted)
	if err != nil {
		return err
	}

	LogMessage(LogLevelInfo, "New friend added successfully")
	return nil
}

// Lists all the friends names in the friendsList
func listFriendsNames(friends FriendsList) []string {
	var friendsNames []string
	for _, friend := range friends {
		friendsNames = append(friendsNames, friend.Name)
	}
	return friendsNames
}

func setupRouter(handler *FriendsHandler) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	// TODO: Figure out if .Default() is what I need or something else
	r := gin.Default()

	r.GET("/friends", handler.GetFriendsHandler)
	r.GET("/friends/random", handler.GetRandomFriendHandler)
	r.GET("/friends/count", handler.GetFriendCountHandler)
	r.GET("/friends/id/:id", handler.GetFriendByIDHandler)
	r.GET("/friends/name/:name", handler.GetFriendByNameHandler)
	r.POST("/friends", handler.PostNewFriendHandler)

	return r
}

func main() {

	//var friendsFilePath = "config/friends.yaml"
	var dbFilePath = "sql/friends.db"

	// Sets up the logger
	SetupLogger()

	// Open the database connection
	db, err := createOrOpenSQLiteDB(dbFilePath)
	if err != nil {
		LogMessage(LogLevelFatal, "Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the table
	if err := createTable(db); err != nil {
		LogMessage(LogLevelFatal, "Failed to create table: %v", err)
		panic(err)
	}

	friendsList, err := buildFriendsList(db)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		panic(err)
	}

	// Create an instance of your handler struct with the friendsList
	friendsHandler := NewFriendsHandler(friendsList, db)

	router := setupRouter(friendsHandler)

	err = router.Run()
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		panic(err)
	}
}
