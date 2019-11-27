package types

// UserCreationRequest describes the data needed to create a new user
type UserCreationRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Response is the standard response struct
type Response struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"body"`
}
