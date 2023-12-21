package main

import (
	"errors"
	"fmt"
	"log"
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
		log.Fatalf("error: %v", err)
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
	lastContactedNormalised := time.Date(lastContactedDate.Year(), lastContactedDate.Month(), lastContactedDate.Day(), 0, 0, 0, 0, lastContactedDate.Location())
	currDateNormalised := time.Date(currDate.Year(), currDate.Month(), currDate.Day(), 0, 0, 0, 0, currDate.Location())

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

func main() {

	friendsList, err := buildFriendsList("config/friends.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	chosenFriend, err := pickRandomFriend(friendsList)

	fmt.Printf("You should talk to %s who you last spoke to on %s", chosenFriend.Name, chosenFriend.LastContacted)
}
