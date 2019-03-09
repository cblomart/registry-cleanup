package registry

import "time"

//Image is the image description
type Image struct {
	ID           string
	Parent       string
	Created      time.Time
	Author       string
	Architecture string
	OS           string
	CheckSum     string
}
