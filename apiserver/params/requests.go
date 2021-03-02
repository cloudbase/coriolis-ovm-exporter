package params

// LoginRequest represents username/password request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
