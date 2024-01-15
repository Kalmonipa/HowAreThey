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

// GET /friends/id/{ID}
func (h *FriendsHandler) GetFriendByIDHandler(c *gin.Context) {
	friendID := c.Param("id")
	friend, err := getFriendByID(friendID, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friend)
}

// GET /friends/name/{NAME}
func (h *FriendsHandler) GetFriendByNameHandler(c *gin.Context) {
	friendName := c.Param("name")
	friend, err := getFriendByName(friendName, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friend)
}
