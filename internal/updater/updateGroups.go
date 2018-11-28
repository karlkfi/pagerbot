package updater

import (
	"fmt"
	"github.com/karlkfi/pagerbot/internal/config"
	"github.com/karlkfi/pagerbot/internal/pagerduty"
	log "github.com/sirupsen/logrus"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Ensure all the slack groups are up to date
func (u *Updater) updateGroups() {
	for _, group := range config.Config.Groups {
		lf := log.Fields{
			"group": group.Name,
		}

		var currentUsers []*User
		var changeover time.Time
		for _, s := range group.Schedules {
			lf["scheduleId"] = s
			schd := u.Schedules.ById(s)
			if schd == nil {
				log.WithFields(lf).Warning("Could not find schedule with ID")
				continue
			}

			var activePeriod *pagerduty.CallPeriod

			if schd.NextPeriod != nil {
				if changeover.IsZero() || schd.NextPeriod.Start.Before(changeover) {
					changeover = schd.NextPeriod.Start
				}
			}

			if !changeover.IsZero() && time.Now().UTC().After(changeover) {
				activePeriod = schd.NextPeriod
			} else if schd.CurrentPeriod != nil {
				activePeriod = schd.CurrentPeriod
			}

			if activePeriod != nil {
				lf["pagerdutyId"] = activePeriod.User
				usr := u.Users.ByPagerdutyId(activePeriod.User)
				if usr == nil {
					log.WithFields(lf).Warning("Could not find user with ID")
					continue
				}
				currentUsers = append(currentUsers, usr)
			}
		}

		lf["scheduleId"] = nil
		lf["pagerdutyId"] = nil
		lf["changeover"] = changeover

		var pdUsers []string
		var slackUsers []string
		var userNames []string

		for _, u := range currentUsers {
			pdUsers = append(pdUsers, u.PagerdutyId)
			slackUsers = append(slackUsers, u.SlackId)
			userNames = append(userNames, fmt.Sprintf("@%s", u.SlackName))
		}

		lf["pdUsers"] = pdUsers
		lf["slackUsers"] = slackUsers

		currentMembers, err := u.Slack.GroupMembers(group.Name)
		if err != nil {
			lf["err"] = err
			log.WithFields(lf).Warning("Could not get Slack group members")
			continue
		}

		lf["currentMembers"] = currentMembers
		log.WithFields(lf).Debug("Group status")
		sort.Strings(currentMembers)
		sort.Strings(slackUsers)
		if !reflect.DeepEqual(currentMembers, slackUsers) {
			err := u.Slack.UpdateMembers(group.Name, slackUsers)
			if err != nil {
				lf["err"] = err
				log.WithFields(lf).Warning("Could not update Slack group members")
				continue
			}
			log.WithFields(lf).Info("Updating group members")

			userList := fmt.Sprintf("[ %s ]", strings.Join(userNames, ", "))
			msgText := fmt.Sprintf(group.UpdateMessage.Message, userList)
			for _, c := range group.UpdateMessage.Channels {
				u.Slack.PostMessage(c, msgText)
			}
		}
	}
}
