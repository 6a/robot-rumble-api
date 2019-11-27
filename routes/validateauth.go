package routes

import (
	"context"
	"database/sql"
	"encoding/base64"
	"strings"

	"../database"
	"../types"
	"github.com/aws/aws-lambda-go/events"
)

// ValidateAuth validates credentials
func ValidateAuth(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {

	if authHeader, ok := request.Headers["Authorization"]; ok {
		decodedHeader, err := base64.StdEncoding.DecodeString(strings.Replace(authHeader, "Basic ", "", 1))
		if err != nil {
			r.StatusCode = 401
			r.Message = "Auth header invalid format"
		} else {
			var credentials = strings.Split(string(decodedHeader), ":")
			if len(credentials) != 2 {
				r.StatusCode = 401
				r.Message = "Auth header invalid format"
			} else {
				valid, err := database.CheckPassword(credentials[0], credentials[1])
				if err != nil {
					if err == sql.ErrNoRows {
						r.StatusCode = 403
						r.Message = "Username or password incorrect"
					} else {
						r.StatusCode = 400
						r.Message = err.Error()
					}
				} else {
					if valid {
						r.StatusCode = 204
					} else {
						r.StatusCode = 403
						r.Message = "Username or password incorrect"
					}
				}
			}
		}
	} else {
		r.StatusCode = 401
		r.Message = "Missing 'Authorization' header"
	}

	return r, nil
}
