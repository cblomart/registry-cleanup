package hub

import "time"

//Tags is the tags response
type Tags struct {
	Count    int
	Next     string
	Previous string
	Results  []Tag
}

//Tag is a tag
type Tag struct {
	Name        string
	FullSize    int
	Images      []Image
	ID          int
	Repository  int
	Creator     int
	LastUpdater int       `json:"last_updater"`
	LastUpdated time.Time `json:"last_updated"`
	ImageID     int       `json:"image_id"`
	V2          bool
}

//Image contains an image information
type Image struct {
	Size         int
	Architecture string
	OS           string
}
