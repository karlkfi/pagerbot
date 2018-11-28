package pagerduty

import (
	"fmt"
	"github.com/PagerDuty/go-pagerduty"
	log "github.com/sirupsen/logrus"
)

type User struct {
	Id    string
	Name  string
	Email string
}

// pdUsers returns a list of PagerDuty users.
// Pagination is required by the PagerDuty API, so this may make multiple calls.
func (a *Api) pdUsers() ([]pagerduty.User, error) {
	var users []pagerduty.User

	pageLimit := uint(25)
	listOpts := pagerduty.APIListObject{
		Offset: 0,
		Limit:  pageLimit,
	}
	opts := pagerduty.ListUsersOptions{APIListObject: listOpts}

	for {
		res, err := a.client.ListUsers(opts)
		if err != nil {
			return users, fmt.Errorf("Failed to list PagerDuty users: %s", err)
		}

		users = append(users, res.Users...)
		log.WithFields(log.Fields{
			"pageSize": len(res.Users),
			"total":    len(users),
		}).Debug("Recieved page of PagerDuty users")

		listOpts.Offset += pageLimit
		opts = pagerduty.ListUsersOptions{APIListObject: listOpts}

		if res.APIListObject.More != true {
			return users, nil
		}
	}
	return users, nil
}

func (a *Api) Users() ([]User, error) {
	pdUserList, err := a.pdUsers()
	if err != nil {
		return nil, err
	}

	userList := make([]User, 0, len(pdUserList))
	for _, user := range pdUserList {
		userList = append(userList, User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}
	return userList, nil
}
