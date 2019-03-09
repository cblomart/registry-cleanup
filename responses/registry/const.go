package registry

const (
	//ManifestMimeV2 mime type of manifests v2 format
	ManifestMimeV2 = "application/vnd.docker.distribution.manifest.v2+json"
	//ManifestMimeV1 mime type of manifests v1 format
	ManifestMimeV1 = "application/vnd.docker.distribution.manifest.v1+prettyjws"
	//AuthHeader registry authentication header
	AuthHeader = "Www-Authenticate"
	//DigestHeader registery digest header
	DigestHeader = "Docker-Content-Digest"
	//ValidAuthHeader regex to  validate auth header
	ValidAuthHeader = "^[Bb]earer *((realm|service|scope|error)=\"[A-Za-z0-9-_./:]+\",?){2,4}$"
	//Scope to delete tags
	Scope = "pull,push,delete"
)
