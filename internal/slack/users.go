package slack

import (
	"context"
	"fmt"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
	"time"
)

type User struct {
	Id    string
	Name  string
	Email string
}

// slackUsers queries the full user list and blocks/retries if rate limited
func (a *Api) slackUsers() ([]slack.User, error) {
	var err error
	var p slack.UserPagination

	userList := make([]slack.User, 0)
	ctx := context.Background()
	opts := []slack.GetUsersOption{
		slack.GetUsersOptionPresence(true),
		slack.GetUsersOptionLimit(500),
	}
	for p = a.api.GetUsersPaginated(opts...); !p.Done(err); p, err = p.Next(ctx) {
		if err != nil {
			if rle, ok := err.(*slack.RateLimitedError); ok {
				log.WithFields(log.Fields{
					"error": rle,
				}).Warning("Being rate limited by Slack")
				time.Sleep(rle.RetryAfter)
				continue
			}
		}
		userList = append(userList, p.Users...)
		log.WithFields(log.Fields{
			"pageSize": len(p.Users),
			"total":    len(userList),
		}).Debug("Recieved page of Slack users")
	}
	if err != nil && !p.Done(err) {
		return userList, fmt.Errorf("Failed polling Slack for users: %s", err)
	}

	return userList, nil
}

// UserMap - returns a map of slack users indexed by email
func (a *Api) UserMap() (map[string]User, error) {
	users, err := a.slackUsers()
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]User, len(users))
	for _, user := range users {
		userMap[user.Profile.Email] = User{
			Name:  user.Name,
			Id:    user.ID,
			Email: user.Profile.Email,
		}
	}
	return userMap, nil
}
