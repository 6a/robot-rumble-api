package routes

import (
	"context"
	"encoding/json"

	"../database"
	"../types"
	"github.com/aws/aws-lambda-go/events"
)

const maxRows = 100

// GetLeaderboard gets the top N players in the leaderboard as well as an extra blob containing the stats for a particular user
func GetLeaderboard(ctx context.Context, request events.APIGatewayProxyRequest) (r types.Response, err error) {
	ucr := types.LeaderboardRequest{}
	err = json.Unmarshal([]byte(request.Body), &ucr)
	if err != nil {
		r.StatusCode = 400
		r.Message = err.Error()
	} else {
		userexists, _ := database.UserExists(ucr.Name)
		if !userexists {
			r.StatusCode = 404
			r.Message = "The specified user does not exist"
		} else {
			leaderboard, err := database.GetLeaderboard(ucr.Name, maxRows)
			if err != nil {
				r.StatusCode = 400
				r.Message = err.Error()
			} else {
				jsonLeaderboards, err := json.Marshal(leaderboard)
				if err != nil {
					r.StatusCode = 400
					r.Message = err.Error()
				} else {
					r.StatusCode = 200
					r.Message = string(jsonLeaderboards)
				}
			}
		}
	}

	return r, nil
}
