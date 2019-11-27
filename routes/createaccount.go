package routes

import (
	"context"
	"encoding/json"

	"../database"
	"../types"
	"github.com/aws/aws-lambda-go/events"
)

// CreateAccount creates an account
func CreateAccount(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	ucr := types.UserCreationRequest{}
	err = json.Unmarshal([]byte(request.Body), &ucr)
	if err != nil {
		r.StatusCode = 400
		r.Message = err.Error()
	} else if len(ucr.Name) < 2 || len(ucr.Password) < 6 {
		r.StatusCode = 400
		if len(ucr.Name) < 2 {
			r.Message = "Username too short"
		} else {
			r.Message = "Password too short"
		}
	} else {
		err = database.CreateUser(ucr.Name, ucr.Password)
		if err != nil {
			r.StatusCode = 400
			r.Message = err.Error()
		} else {
			r.StatusCode = 200
			r.Message = "Account created"
		}
	}

	return r, nil
}
