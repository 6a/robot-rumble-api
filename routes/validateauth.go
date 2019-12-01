package routes

import (
	"context"

	"../database"
	"../types"
	"github.com/aws/aws-lambda-go/events"
)

// ValidateAuth validates credentials
func ValidateAuth(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	r.StatusCode, r.Message = database.ProcessAuth(request.Headers)

	return r, nil
}
