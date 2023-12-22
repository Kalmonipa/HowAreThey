package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Friend struct {
	Name          string `yaml:"name"`
	LastContacted string `yaml:"lastContacted"`
}

type FriendsList []Friend

func buildFriendsList(filePath string) (FriendsList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML into a map
	var friends FriendsList
	err = yaml.Unmarshal(data, &friends)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		os.Exit(1)
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

// SaveFriendsListToYAML serializes the FriendsList and saves it to a YAML file.
func SaveFriendsListToYAML(friends FriendsList, filePath string) error {
	data, err := yaml.Marshal(friends)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func main() {

	// Sets up the logger
	SetupLogger()

	var filePath = "config/friends.yaml"

	friendsList, err := buildFriendsList(filePath)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		os.Exit(1)
	}

	chosenFriend, err := pickRandomFriend(friendsList)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		os.Exit(1)
	}

	updatedChosenFriend := updateLastContacted(chosenFriend, time.Now())

	for ind, friend := range friendsList {
		if friend.Name == updatedChosenFriend.Name {
			friendsList[ind] = updatedChosenFriend
		}
	}

	err = SaveFriendsListToYAML(friendsList, filePath)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		os.Exit(1)
	}

	LogMessage(LogLevelInfo, "Suggested conversation is with %s. You last contacted them on %s", chosenFriend.Name, chosenFriend.LastContacted)
}
