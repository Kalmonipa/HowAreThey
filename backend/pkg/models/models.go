package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"howarethey/pkg/logger"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// DiscordWebhookPayload defines the JSON structure for the webhook payload
type DiscordWebhookPayload struct {
	Username *string `json:"username,omitempty"`
	Content  *string `json:"content"`
}

type Friend struct {
	ID            string
	Name          string
	LastContacted string
	Birthday      string
	Notes         string
}

type FriendsList []Friend

// Builds the list from the yaml file specified
func BuildFriendsList(db *sql.DB) (FriendsList, error) {
	// Query the database
	rows, err := db.Query("SELECT id, name, lastContacted, birthday, notes FROM friends")
	if err != nil {
		logger.LogMessage(logger.LogLevelFatal, "Failed to select from db: %v", err)
		return nil, err
	}
	defer rows.Close()

	var friends FriendsList
	for rows.Next() {
		var f Friend
		if err := rows.Scan(&f.ID, &f.Name, &f.LastContacted, &f.Birthday, &f.Notes); err != nil {
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
func DeleteFriend(db *sql.DB, friend Friend) error {
	stmt, err := db.Prepare("DELETE FROM friends WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(friend.ID)
	if err != nil {
		return err
	}

	logger.LogMessage(logger.LogLevelInfo, friend.Name+" (ID:"+friend.ID+") deleted successfully")
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

// Notifications

func SendDiscordNotification(friend Friend, url string) {
	var username = "HowAreThey"

	var content = "You should get in touch with " + friend.Name + ". You haven't spoken to them since " +
		friend.LastContacted + ". "

	if friend.Notes != "" {
		content = content + "Here's what you've got written down for them: " + friend.Notes
	}

	// Create the payload
	payload := DiscordWebhookPayload{
		Content:  &content,
		Username: &username,
	}

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.LogMessage(logger.LogLevelWarn, "Failed to marshal the payload: %s", err)
	}

	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.LogMessage(logger.LogLevelWarn, "Failed to send the request: %s", err)
	}
	defer resp.Body.Close()
}

func SendNtfyNotification(friend Friend, url string) {
	var message = "You should get in touch with " + friend.Name + ". You haven't spoken to them since " +
		friend.LastContacted + ". "

	if friend.Notes != "" {
		message = message + "Here's what you've got written down for them: " + friend.Notes
	}

	// Send the POST request
	resp, err := http.Post(url, "text/plain", bytes.NewBufferString(message))
	if err != nil {
		logger.LogMessage(logger.LogLevelWarn, "Failed to send the request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.LogMessage(logger.LogLevelFatal, "Failed to send message to nfty. Status: %s", resp.Status)
	}
}

func UpdateFriend(friendList FriendsList, newFriend *Friend) (FriendsList, error) {
	for i, friend := range friendList {
		logger.LogMessage(logger.LogLevelDebug, "Checking %s", friend.Name)
		if friend.ID == newFriend.ID {
			friendList[i] = *newFriend
		}
	}
	return friendList, nil
}

func UpdateLastContacted(friend Friend, todaysDate time.Time) *Friend {
	friend.LastContacted = todaysDate.Format("02/01/2006")

	return &friend
}

// SQL Functions

// addFriend inserts a new friend into the database
func AddFriend(db *sql.DB, newFriend Friend) error {
	stmt, err := db.Prepare("INSERT INTO friends(name, lastContacted, birthday, notes) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newFriend.Name, newFriend.LastContacted, newFriend.Birthday, newFriend.Notes)
	if err != nil {
		return err
	}

	successMsg := newFriend.Name + " added successfully"

	logger.LogMessage(logger.LogLevelInfo, successMsg)
	return nil
}

// Updates a friend with new details
func SqlUpdateFriend(db *sql.DB, id string, updatedFriend *Friend) error {
	stmt, err := db.Prepare("UPDATE friends SET name = ?, lastContacted = ? , notes = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(updatedFriend.Name, updatedFriend.LastContacted, updatedFriend.Birthday, updatedFriend.Notes, id)
	if err != nil {
		return err
	}

	logger.LogMessage(logger.LogLevelInfo, "Friend with ID %s updated successfully", updatedFriend.ID)
	return nil
}
