package registry

//TokenResp is a token request response
type TokenResp struct {
	Token        string `json:"token"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	IssuedAt     string `json:"issued_at"`
	RefreshToken string `json:"refresh_token"`
}
