package main

import (
	"encoding/base64"
	"math/rand"
	"medialogger/server/datastructs"
	"medialogger/server/privacy"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TODO: Store this in a database instead!
var users = map[string]*datastructs.User{
	"default": {
		UUID:         1,
		Username:     "default",
		Email:        "default@email.com",
		PasswordHash: privacy.HashPassword("password", 1),
		SessionToken: 0,
		SavedMedia:   []datastructs.MediaItem{},
		SavedLists:   []datastructs.MediaList{},
	},
}

func main() {

	// TODO: import config settings here

	router := gin.Default()

	router.POST("/login", postLogin)
	router.GET("/users/:name", getUser)
	router.GET("/users/:name/email", getEmail)
	router.POST("/users/:name/email", postEmail)
	router.POST("/users/:name/password", postPassword)
	router.PUT("/users/:name/media", putMedia)
	router.GET("/users/:name/media", getMedia)
	router.POST("/users/:name/media/:uid", postMedia)
	router.DELETE("/users/:name/media/:uid", deleteMedia)
	router.POST("/users/:name/logout", postLogout)

	router.Run("localhost:8080")
}

// handles logins and passes back a session token, needs more security + collision avoidance
func postLogin(c *gin.Context) {
	username := c.PostForm("username")
	var user, userExists = users[username]
	if !userExists {
		c.String(http.StatusUnauthorized, "Bad username or password.")
		return
	}
	password, pwFound := c.GetPostForm("password")
	if !pwFound {
		c.String(http.StatusBadRequest, "No password given.")
	}
	pwHash := privacy.HashPassword(password, user.UUID)
	if pwHash != user.PasswordHash {
		c.String(http.StatusUnauthorized, "Bad username or password.")
		return
	}

	// Token generation should be seeded with time and should use the crypto/rand package instead
	var token = rand.Uint32()
	user.SessionToken = token

	// TODO: Start timeout goroutine here

	c.String(http.StatusOK, "%d", token)
}

func getUser(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		strippedUser := privacy.StripPII(user)
		c.String(http.StatusOK, strippedUser)
	}
}

func getEmail(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		c.String(http.StatusOK, user.Email)
	}
}

func postEmail(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		newEmail := c.PostForm("newEmail")
		if newEmail == "" {
			c.String(http.StatusBadRequest, "No replacement email given.")
			return
		}
		user.Email = newEmail
		c.String(http.StatusOK, newEmail)
	}
}

func postPassword(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		oldPassword := c.PostForm("oldPassword")
		if oldPassword == "" {
			c.String(http.StatusBadRequest, "No password given.")
			return
		}
		if privacy.HashPassword(oldPassword, user.UUID) != user.PasswordHash {
			c.String(http.StatusUnauthorized, "Bad password.")
			return
		}
		newPassword := c.PostForm("newPassword")
		if newPassword == "" {
			c.String(http.StatusBadRequest, "No new password given.")
			return
		}
		user.PasswordHash = privacy.HashPassword(newPassword, user.UUID)
	}
}

// Adds a new piece of media to the given user.
func putMedia(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		var newMedia datastructs.MediaItem

		if err := c.BindJSON(&newMedia); err != nil {
			c.String(http.StatusBadRequest, "Malformed JSON.")
			return
		}

		var _, uidIndex = findMediaUID(user, newMedia.UID)

		if uidIndex >= 0 {
			c.String(http.StatusUnprocessableEntity, "UID already exists. Use POST at the proper endpoint to overwrite existing media.")
		}

		user.SavedMedia = append(user.SavedMedia, newMedia)
		c.JSON(http.StatusCreated, newMedia)
	}
}

// Updates a piece of media matching the UID parameter for the given user with the new data provided.
func postMedia(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		uidStr := c.Param("uid")
		uid64, parseErr := strconv.ParseInt(uidStr, 10, 32)
		if parseErr != nil {
			c.String(http.StatusBadRequest, "Malformed media UID parameter.")
		}
		uid := uint32(uid64)

		media, index := findMediaUID(user, uid)

		if index >= 0 {
			if bindErr := c.BindJSON(media); bindErr != nil {
				c.String(http.StatusBadRequest, "Malformed media JSON.")
				return
			} else {
				c.JSON(http.StatusCreated, *media)
			}
		} else {
			c.String(http.StatusUnprocessableEntity, "Use PUT to add new media.")
		}
	}
}

// Gets all media for the user provided as a parameter
func getMedia(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]

		c.JSON(http.StatusOK, user.SavedMedia)
	}
}

// Deletes the media item specified by the parameter
func deleteMedia(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		uidStr := c.Param("uid")
		uid64, parseErr := strconv.ParseInt(uidStr, 10, 32)
		if parseErr != nil {
			c.String(http.StatusBadRequest, "Malformed media UID parameter.")
			return
		}
		uid := uint32(uid64)
		_, index := findMediaUID(user, uid)
		savedMedia := &user.SavedMedia
		*savedMedia = append((*savedMedia)[:index], (*savedMedia)[index+1:]...)
		c.String(http.StatusOK, "Media with UID %d deleted.", uid)
	}
}

// Finds a media item under the given user that matches the provided UID; the second return value is -1 if no such
// media item is found, and otherwise returns the index of the media item.
// Simple inefficient search, will replace this with sorting + sensible efficient search in time
func findMediaUID(user *datastructs.User, uid uint32) (*datastructs.MediaItem, int) {
	var mediaMatch *datastructs.MediaItem
	var index = -1
	for x := 0; x < len(user.SavedMedia); x++ {
		var media *datastructs.MediaItem = &user.SavedMedia[x]
		if media.UID == uid {
			mediaMatch = media
			index = x
			break
		}
	}
	return mediaMatch, index
}

// These functions will be added at a later point.
// func putList() {

// }

// func postList() {

// }

// func getLists() {

// }

// func getSpecificList() {

// }

// Clears the user's session token, logging them out.
func postLogout(c *gin.Context) {
	if validateSessionToken(c) {
		username := c.Param("user")
		user := users[username]
		user.SessionToken = 0
		c.String(http.StatusOK, "Logged out user %s.", username)
	}
}

// Checks whether the session token of the request matches that of the user in the request. Returns false and sends an HTTP error if it's not a match.
func validateSessionToken(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	auth := strings.Split(authHeader, " ")
	if authHeader == "" || len(auth) != 2 || auth[0] != "Basic" {
		c.String(http.StatusBadRequest, "No session token or badly formed session token.")
		return false
	}
	username := c.Param("name")
	var user, userExists = users[username]
	// Check if token with decode from base 64 into a byte string
	var tokenArr, err = base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		c.String(http.StatusBadRequest, "No session token or badly formed session token.")
		return false
	}
	// Trim out leading colon from not having a username in authentication header
	tokenStr := strings.Trim(string(tokenArr), ":")

	token, tokenErr := strconv.Atoi(tokenStr)
	if tokenErr != nil {
		c.String(http.StatusBadRequest, "No session token or badly formed session token.")
		return false
	}
	// Check if session token is correct
	if !userExists || user.SessionToken != uint32(token) || token <= 0 {
		c.String(http.StatusUnauthorized, "User does not exist, or you are not logged into this account.")
		return false
	} else {
		return true
	}
}
