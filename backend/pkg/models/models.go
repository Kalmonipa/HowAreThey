package models

import (
	"database/sql"
	"errors"
	"fmt"
	"howarethey/pkg/logger"
	"math/rand"
	"strings"
	"time"
)

type Friend struct {
	ID            string
	Name          string
	LastContacted string
	Notes         string
}

type FriendsList []Friend

// addFriend inserts a new friend into the database
func AddFriend(db *sql.DB, newFriend Friend) error {
	stmt, err := db.Prepare("INSERT INTO friends(name, lastContacted, notes) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newFriend.Name, newFriend.LastContacted, newFriend.Notes)
	if err != nil {
		return err
	}

	successMsg := newFriend.Name + " added successfully"

	logger.LogMessage(logger.LogLevelInfo, successMsg)
	return nil
}

// Builds the list from the yaml file specified
func BuildFriendsList(db *sql.DB) (FriendsList, error) {
	// Query the database
	rows, err := db.Query("SELECT id, name, lastContacted, notes FROM friends")
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to select from db: %v", err)
		return nil, err
	}
	defer rows.Close()

	var friends FriendsList
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.Name, &f.LastContacted, &f.Notes); err != nil {
			logger.LogMessage(logger.LogLevelFatal, "Failed to scan: %v", err)
			return nil, err
		}
		friends = append(friends, f)
	}

	if err := rows.Err(); err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to close: %v", err)
		return nil, err
	}

	return friends, nil
}

// Calculates the weight of each name based on how many days since last contacted
// The longer the time since last contact, the higher the chance of them coming up in the selection
func CalculateWeight(lastContacted string, currDate time.Time) (int, error) {
	layout := "02/01/2006"

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

// Delete a friend from the db based on the ID provided
func DeleteFriend(db *sql.DB, id string) error {
	stmt, err := db.Prepare("DELETE FROM friends WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	logger.LogMessage(logger.LogLevelInfo, "Friend deleted successfully")
	return nil
}

// Returns the friend based on the ID provided
func GetFriendByID(id string, friends FriendsList) (*Friend, error) {
	for _, friend := range friends {
		if friend.ID == id {
			return &friend, nil
		}
	}
	return nil, errors.New("friend not found")
}

// Returns the friend based on the name provided
// This function replaces whitespace with hyphens and makes it lower case
func GetFriendByName(name string, friends FriendsList) (*Friend, error) {
	for _, friend := range friends {
		lowercased := strings.ToLower(friend.Name)
		hyphenated := strings.ReplaceAll(lowercased, " ", "-")
		if hyphenated == name {
			return &friend, nil
		}
	}
	return nil, errors.New("friend not found")
}

// Lists all the friends names in the friendsList
func ListFriendsNames(friends FriendsList) []string {
	var friendsNames []string
	for _, friend := range friends {
		friendsNames = append(friendsNames, friend.Name)
	}
	return friendsNames
}

// Returns how many elements are in the list
func GetFriendCount(friends FriendsList) int {
	return len(friends)
}

func PickRandomFriend(friends FriendsList) (Friend, error) {
	totalWeight := 0
	weights := make([]int, len(friends))

	for i, friend := range friends {

		weight, err := CalculateWeight(friend.LastContacted, time.Now())
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

// Updates a friend with new details
func UpdateFriend(db *sql.DB, id string, updatedFriend Friend) error {
	stmt, err := db.Prepare("UPDATE friends SET name = ?, lastContacted = ? , notes = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(updatedFriend.Name, updatedFriend.LastContacted, updatedFriend.Notes, id)
	if err != nil {
		return err
	}

	logger.LogMessage(logger.LogLevelInfo, "Friend updated successfully")
	return nil
}

func UpdateLastContacted(friend Friend, todaysDate time.Time) Friend {
	friend.LastContacted = todaysDate.Format("02/01/2006")

	return friend
}
