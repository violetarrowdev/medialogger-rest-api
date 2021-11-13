package datastructs

type ListOrder struct {
	UID   int `json:"uid"`
	Order int `json:"order"`
}

type MediaItem struct {
	UID            uint32 `json:"uid"` // has unique ID only within scope of user
	Title          string `json:"title"`
	ReleaseDate    string `json:"releaseDate"` // mm/dd/yyyy or mm/yyyy or yyyy
	Medium         string `json:"medium"`
	Description    string `json:"description"`
	Thumbnail      string `json:"thumbnail"` // link to thumbnail image, empty if no thumbnail
	Rating         int    `json:"rating"`
	LinkedPlatform string `json:"linkedPlatform"` // empty when not linked
	Notes          string `json:"notes"`
}

type MediaList struct {
	Name        string      `json:"name"` // lists must have unique names for each user
	Description string      `json:"description"`
	MediaTypes  []string    `json:"mediaTypes"` // array is empty if all mediums allowed
	Contents    []ListOrder `json:"contents"`
}

type User struct {
	UUID         uint32      `json:"uuid,omitempty"` // unique across all users
	Username     string      `json:"username"`       // must be unique
	Email        string      `json:"email"`
	PasswordHash string      `json:"passwordHash,omitempty"` // SHA-256
	SessionToken uint8       `json:"-"`                      // TODO: temporary session token
	SavedMedia   []MediaItem `json:"savedMedia"`
	SavedLists   []MediaList `json:"savedLists"`
}
