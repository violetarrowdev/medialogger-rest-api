package main

import (
	"math/rand"
	"medialogger/server/datastructs"
	"medialogger/server/privacy"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

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
}

// handles logins and passes back a session token, needs more security + collision avoidance
func postLogin(c *gin.Context) {
	username := c.GetString("username")
	var user, userExists = users[username]
	if !userExists {
		c.String(http.StatusUnauthorized, "Bad username or password.")
		return
	}
	pwHash := privacy.HashPassword(c.GetString("password"), user.UUID)
	if pwHash != user.PasswordHash {
		c.String(http.StatusUnauthorized, "Bad username or password.")
		return
	}
	var token uint8 = uint8(rand.Uint32())
	user.SessionToken = token

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
		newEmail := c.GetString("newEmail")
		user.Email = newEmail
		c.String(http.StatusOK, newEmail)
	}
}

func postPassword(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		oldPassword := c.GetString("oldPassword")
		if oldPassword == "" {
			c.String(http.StatusBadRequest, "No password given.")
			return
		}
		if privacy.HashPassword(oldPassword, user.UUID) != user.PasswordHash {
			c.String(http.StatusUnauthorized, "Bad password.")
			return
		}
		newPassword := c.GetString("newPassword")
		if newPassword == "" {
			c.String(http.StatusBadRequest, "No new password given.")
			return
		}
		user.PasswordHash = privacy.HashPassword(newPassword, user.UUID)
	}
}

func putMedia(c *gin.Context) {
	if validateSessionToken(c) {
		user := users[c.Param("name")]
		var newMedia datastructs.MediaItem

		if err := c.BindJSON(&newMedia); err != nil {
			c.String(http.StatusBadRequest, "Malformed JSON.")
			return
		}

		user.SavedMedia = append(user.SavedMedia, newMedia)
		c.JSON(http.StatusCreated, newMedia)
	}
}

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
			c.String(http.StatusMethodNotAllowed, "Use PUT to add new media.")
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

// Finds a media item under the given user that matches the provided UID; the second return value is false if no such item is found.
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

// func putList() {

// }

// func postList() {

// }

// func getLists() {

// }

// func getSpecificList() {

// }

func postLogout(c *gin.Context) {
	if validateSessionToken(c) {
		username := c.Param("user")
		user := users[username]
		user.SessionToken = 0
		c.String(http.StatusOK, "Logged out user %s.", username)
	}
}

func validateSessionToken(c *gin.Context) bool {
	authHeader := c.GetHeader("authorization")
	auth := strings.Split(c.GetHeader("authorization"), " ")
	if authHeader == "" || len(auth) != 2 || auth[0] != "Basic" {
		c.String(http.StatusBadRequest, "No session token or badly formed session token.")
		return false
	}
	username := c.Param("name")
	var user, userExists = users[username]
	var token64, err = strconv.ParseInt(auth[1], 10, 16)
	if err == nil {
		c.String(http.StatusBadRequest, "No session token or badly formed session token.")
		return false
	}
	token := uint8(token64)
	if !userExists || user.SessionToken != token || token <= 0 {
		c.String(http.StatusUnauthorized, "User does not exist, or you are not logged into this account.")
		return false
	} else {
		return true
	}
}
