package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	expectedResult := FriendsMap{
		"johnwick":    Friend{Name: "John Wick", LastContacted: "06/06/2023"},
		"peterparker": Friend{Name: "Peter Parker", LastContacted: "12/12/2023"},
	}

	readFileResult, err := buildFriendsList("test/friends_test.yaml")
	if err != nil {
		t.Fatalf("Failed to read or parse YAML: %v", err)
	}

	assert.Equal(t, expectedResult, readFileResult)
}
