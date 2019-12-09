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
	lr := types.LeaderboardRequest{}
	err = json.Unmarshal([]byte(request.Body), &lr)
	if err != nil {
		r.StatusCode = 400
		r.Message = err.Error()
	} else {
		userexists, _ := database.UserExists(lr.Name)
		if !userexists {
			r.StatusCode = 404
			r.Message = "The specified user does not exist"
		} else {
			leaderboard, err := database.GetLeaderboard(lr.Name, maxRows)
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
