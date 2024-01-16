package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FriendsHandler struct {
	FriendsList FriendsList
	DB          *sql.DB
}

func NewFriendsHandler(friendsList FriendsList, db *sql.DB) *FriendsHandler {
	return &FriendsHandler{
		FriendsList: friendsList,
		DB:          db,
	}
}

// DELETE /friends/{ID}
func (h *FriendsHandler) DeleteFriend(c *gin.Context) {
	friendID := c.Param("id")

	err := deleteFriend(h.DB, friendID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	friendsList, err := buildFriendsList(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FriendsList = friendsList

	c.JSON(http.StatusCreated, gin.H{"message": "Friend removed successfully"})
}

// GET /friends
func (h *FriendsHandler) GetFriends(c *gin.Context) {
	c.JSON(http.StatusOK, h.FriendsList)
}

// GET /friends/random
func (h *FriendsHandler) GetRandomFriend(c *gin.Context) {
	randomFriend, err := pickRandomFriend(h.FriendsList)
	if err != nil {
		LogMessage(LogLevelFatal, "error: %v", err)
		c.JSON(http.StatusNotFound, "failed to pick a friend")
	}

	c.JSON(http.StatusOK, randomFriend)
}

// GET /friends/count
func (h *FriendsHandler) GetFriendCount(c *gin.Context) {
	c.JSON(http.StatusOK, getFriendCount(h.FriendsList))
}

// GET /friends/id/{ID}
func (h *FriendsHandler) GetFriendByID(c *gin.Context) {
	friendID := c.Param("id")
	friend, err := getFriendByID(friendID, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friend)
}

// GET /friends/name/{NAME}
func (h *FriendsHandler) GetFriendByName(c *gin.Context) {
	friendName := c.Param("name")
	friend, err := getFriendByName(friendName, h.FriendsList)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, friend)
}

// POST /friends
func (h *FriendsHandler) PostNewFriend(c *gin.Context) {
	var newFriend Friend
	if err := c.ShouldBindJSON(&newFriend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := addFriend(h.DB, newFriend)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	friendsList, err := buildFriendsList(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FriendsList = friendsList

	successMsg := newFriend.ID + " (" + newFriend.Name + ") added successfully"

	c.JSON(http.StatusCreated, gin.H{"message": successMsg})
}

// PUT /friends/{ID}
func (h *FriendsHandler) PutFriend(c *gin.Context) {
	id := c.Param("id")

	var updatedFriend Friend
	if err := c.ShouldBindJSON(&updatedFriend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := updateFriend(h.DB, id, updatedFriend); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	friendsList, err := buildFriendsList(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.FriendsList = friendsList

	successMsg := updatedFriend.ID + " updated successfully"

	c.JSON(http.StatusOK, gin.H{"message": successMsg})
}
