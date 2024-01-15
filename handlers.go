package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FriendsHandler struct {
	FriendsList FriendsList
}

func NewFriendsHandler(friendsList FriendsList) *FriendsHandler {
	return &FriendsHandler{
		FriendsList: friendsList,
	}
}

// GET /friends/list
func (h *FriendsHandler) GetFriendsHandler(c *gin.Context) {
	//friendsNames := ListFriendsNames(h.FriendsList)
	c.JSON(http.StatusOK, h.FriendsList)
}

// GET /friends/random
func (h *FriendsHandler) GetRandomFriendHandler(c *gin.Context) {
	randomFriend, err := pickRandomFriend(h.FriendsList)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		c.JSON(http.StatusNotFound, "failed to pick a friend")
	}

	c.JSON(http.StatusOK, randomFriend)
}

// GET /friends/count
func (h *FriendsHandler) GetFriendCountHandler(c *gin.Context) {
	c.JSON(http.StatusOK, getFriendCount(h.FriendsList))
}
