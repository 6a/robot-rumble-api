package main

import (
	"context"

	"./database"
	"./routes"
	"./types"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func router(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	switch request.HTTPMethod {
	case "GET":
		return routes.ValidateAuth(ctx, request)
	case "PUT":
		return routes.CreateAccount(ctx, request)
	case "POST":
		return routes.GetLeaderboard(ctx, request)
	}

	r.StatusCode = 500
	r.Message = "Internal server error (could not route the request)"

	return r, nil
}

func main() {
	database.Init()
	lambda.Start(router)

	// ev := events.APIGatewayProxyRequest{}
	// ev.Headers = make(map[string]string)

	// ev.Headers["Authorization"] = "Basic NmE6YW5pbWFsMQ=="
	// res, err := routes.ValidateAuth(nil, ev)

	// ev.Body = `{"name": "6a"}`
	// res, err := routes.GetLeaderboard(nil, ev)

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println(res)
}
