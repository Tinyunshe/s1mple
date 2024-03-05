package auth

type ConfluenceUser struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}
