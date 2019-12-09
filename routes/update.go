package routes

import (
	"context"
	"encoding/json"

	"../database"
	"../types"
	"github.com/aws/aws-lambda-go/events"
)

// Update updates the w/d/l for the specified user. auth must be valid and match user field in body
func Update(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {

	var username string
	r.StatusCode, r.Message, username = database.ProcessAuth(request.Headers)

	if r.StatusCode == 204 {
		uur := types.UserUpdateRequest{}
		err = json.Unmarshal([]byte(request.Body), &uur)
		if err != nil {
			r.StatusCode = 400
			r.Message = err.Error()
		} else {
			userexists, _ := database.UserExists(uur.Name)
			if !userexists {
				r.StatusCode = 404
				r.Message = "The specified user does not exist"
			} else if username != uur.Name {
				r.StatusCode = 403
				r.Message = "Authenticated user does not match update target user"
			} else {
				err = database.Update(uur)
				if err != nil {
					r.StatusCode = 400
					r.Message = err.Error()
				}
			}
		}
	}

	return r, nil
}
