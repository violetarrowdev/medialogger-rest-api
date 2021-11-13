package main

// Represents a book, movie, game, or show
type mediaItem struct {
	UID         string `json:"uid"`
	Title       string `json:"title"`
	ReleaseDate string `json:"releaseDate"` // must be in mm/dd/yyyy format
	Medium      string `json:"medium"`      // the type of media this item is

}

func main() {

}
