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

func pickRandomFriend(friends FriendsList) (Friend, error) {
	if len(friends) == 0 {
		return Friend{}, errors.New("The friends list is empty")
	}
	seed := rand.NewSource(time.Now().Unix())
	r := rand.New(seed)
	randomIndex := r.Intn(len(friends))

	return friends[randomIndex], nil
}

func main() {

	friendsList, err := buildFriendsList("config/friends.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	chosenFriend, err := pickRandomFriend(friendsList)

	fmt.Printf("You should talk to %s who you last spoke to on %s", chosenFriend.Name, chosenFriend.LastContacted)
}
