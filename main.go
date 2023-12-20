package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Friend struct {
	Name          string `yaml:"name"`
	LastContacted string `yaml:"lastContacted"`
}

type FriendsMap map[string]Friend

func buildFriendsList(filePath string) (FriendsMap, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the YAML into a map
	var friends FriendsMap
	err = yaml.Unmarshal(data, &friends)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return friends, nil
}

func main() {

	_, err := buildFriendsList("config/friends.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
