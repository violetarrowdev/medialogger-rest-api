package privacy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"medialogger/server/datastructs"
)

func HashPassword(password string, uuid uint32) string {
	input := password + fmt.Sprintf("%x", uuid)
	output := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))
	return output
}

// strips user PII and returns their data as JSON
func StripPII(user *datastructs.User) string {
	strippedUser := &datastructs.User{
		UUID:         0,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: "",
		SavedMedia:   user.SavedMedia,
		SavedLists:   user.SavedLists,
	}
	userData, err := json.Marshal(strippedUser)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return ""
	} else {
		return string(userData)
	}
}
