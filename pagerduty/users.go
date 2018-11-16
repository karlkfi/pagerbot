package pagerduty

import (
	"bytes"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	log "github.com/Sirupsen/logrus"
)

type Users []User

type User struct {
	Id    string
	Name  string
	Email string
}

func (u Users) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[\n")
	for i, user := range u {
		if i > 0 {
			buffer.WriteString(",\n")
		}
		buffer.WriteString(fmt.Sprintf("  User{ Id:\"%s\", Name:\"%s\", Email:\"%s\" }", user.Id, user.Name, user.Email))
	}
	buffer.WriteString("\n]")
	return buffer.String()
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
		listOpts.Offset += pageLimit
		opts = pagerduty.ListUsersOptions{APIListObject: listOpts}

		if res.APIListObject.More != true {
			return users, nil
		}
	}
	return users, nil
}

func (a *Api) Users() (Users, error) {
	var userList Users

	pdUserList, err := a.pdUsers()
	if err != nil {
		return userList, err
	}

	for _, user := range pdUserList {
		userList = append(userList, User{
			Id:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	log.WithFields(log.Fields{
		"users": userList,
	}).Debug("Known PagerDuty Users")

	return userList, nil
}
