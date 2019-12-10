package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"./database"
	"./routes"
	"./types"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/xid"
)

func createRandom(count int) {
	ev := events.APIGatewayProxyRequest{}
	ev.Headers = make(map[string]string)

	for index := 0; index < count; index++ {
		rand.Seed(time.Now().UTC().UnixNano())

		name := xid.New().String()
		pass := xid.New().String()

		namelen := rand.Intn(12-2) + 2
		name = name[len(name)-namelen:]

		ev.Body = fmt.Sprintf(`{"name": "%v", "password": "%v"}`, name, pass)
		_, _ = routes.CreateAccount(nil, ev)

		ev.Headers["Authorization"] = fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", name, pass))))

		r1 := rand.Intn(50-0) + 0
		r2 := rand.Intn(50-0) + 0
		r3 := rand.Intn(50-0) + 0
		ev.Body = fmt.Sprintf(`{"name": "%v", "wins": %v, "draws": %v, "losses": %v}`, name, r1, r2, r3)
		r, _ := routes.Update(nil, ev)
		fmt.Println(r)
	}
}

func router(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	switch request.HTTPMethod {
	case "GET":
		return routes.ValidateAuth(ctx, request)
	case "PUT":
		return routes.CreateAccount(ctx, request)
	case "POST":
		return routes.GetLeaderboard(ctx, request)
	case "PATCH":
		return routes.Update(ctx, request)
	}

	r.StatusCode = 500
	r.Message = "Internal server error (could not route the request)"

	return r, nil
}

func main() {
	database.Init()

	// ev.Headers["Authorization"] = "Basic NmE6YW5pbWFsMQ=="

	// ev.Body = `{"name": "6a", "wins": 55, "draws": 0, "losses": 0}`

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println(res)

	// createRandom(4)

	lambda.Start(router)
}
