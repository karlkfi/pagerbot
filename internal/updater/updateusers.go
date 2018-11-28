package updater

import (
	log "github.com/sirupsen/logrus"
)

// Fetch the users from Pagerduty and slack, and make sure we can match them
// all up. We match Pagerduty users to Slack users based on their email address
func (u *Updater) updateUsers() {

	var err error
	pdUsers, err := u.Pagerduty.Users()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warning("Failed fetching PagerDuty users")
		return
	}
	log.WithFields(log.Fields{
		"users": pdUsers,
	}).Debug("Fetched PagerDuty users")

	slackUserMap, err := u.Slack.UserMap()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warning("Failed fetching Slack users")
		return
	}
	log.WithFields(log.Fields{
		"users": slackUserMap,
	}).Debug("Fetched Slack users")

	var users UserList
	for _, pdUser := range pdUsers {
		slackUser, found := slackUserMap[pdUser.Email]
		if !found {
			log.WithFields(log.Fields{
				"email":       pdUser.Email,
				"pagerdutyId": pdUser.Id,
			}).Debug("Could not find Slack account for Pagerduty user")
			continue
		}

		usr := User{}
		usr.Name = pdUser.Name
		usr.SlackId = slackUser.Id
		usr.PagerdutyId = pdUser.Id
		usr.SlackName = slackUser.Name
		usr.Email = pdUser.Email
		users.users = append(users.users, &usr)
	}

	u.Users = &users
}
