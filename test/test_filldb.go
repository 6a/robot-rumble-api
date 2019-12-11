package test

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"../routes"
	"github.com/aws/aws-lambda-go/events"
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
