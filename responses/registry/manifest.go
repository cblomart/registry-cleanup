package registry

type versioned struct {
	SchemaVersion int
	MediaType     string
}

//ManifestRespV2 is a manifest v2 request response
type ManifestRespV2 struct {
	versioned
	Config BlobInfo
	Layers []BlobInfo
}

//BlobInfo contains the informations about a blob
type BlobInfo struct {
	MediaType string
	Size      int
	Digest    string
}

//ManifestRespV1 is a manifest v1 request response
type ManifestRespV1 struct {
	versioned
	Name         string
	Tag          string
	Architecture string
	History      []History
}

//History is the raw v1 image configuration
type History struct {
	V1Compatibility string
}
