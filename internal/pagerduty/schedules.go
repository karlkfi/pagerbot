package pagerduty

import (
	"bytes"
	"fmt"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	log "github.com/Sirupsen/logrus"
)

type Schedules []Schedule

type Schedule struct {
	Id            string
	Name          string
	Timezone      string `json:"time_zone"`
	CurrentPeriod *CallPeriod
	NextPeriod    *CallPeriod
}

func (s Schedules) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[\n")
	for i, schd := range s {
		if i > 0 {
			buffer.WriteString(",\n")
		}
		buffer.WriteString(fmt.Sprintf("  &Schedule{ Id:\"%s\", Name:\"%s\", Timezone:\"%s\", CurrentPeriod:\"%s\", NextPeriod:\"%s\" }", schd.Id, schd.Name, schd.Timezone, schd.CurrentPeriod, schd.NextPeriod))
	}
	buffer.WriteString("\n]")
	return buffer.String()
}

type CallPeriod struct {
	Start time.Time
	User  string
}

func (p CallPeriod) String() string {
	return fmt.Sprintf(
		"&CallPeriod{ Start:\"%s\", User:\"%s\" }",
		p.Start.Format(time.RFC3339),
		p.User,
	)
}

// pdUsers returns a list of PagerDuty schedules.
// Pagination is required by the PagerDuty API, so this may make multiple calls.
func (a *Api) pdSchedules() ([]pagerduty.Schedule, error) {
	var schedules []pagerduty.Schedule

	pageLimit := uint(25)
	listOpts := pagerduty.APIListObject{
		Offset: 0,
		Limit:  pageLimit,
	}
	opts := pagerduty.ListSchedulesOptions{APIListObject: listOpts}

	for {
		res, err := a.client.ListSchedules(opts)
		if err != nil {
			return schedules, fmt.Errorf("Failed to list PagerDuty schedules: %s", err)
		}

		schedules = append(schedules, res.Schedules...)
		listOpts.Offset += pageLimit
		opts = pagerduty.ListSchedulesOptions{APIListObject: listOpts}

		if res.APIListObject.More != true {
			return schedules, nil
		}
	}
	return schedules, nil
}

// Fetch the main schedule list then the details about specific schedules
func (a *Api) Schedules() (Schedules, error) {
	var schdList Schedules

	pdSchdList, err := a.pdSchedules()
	if err != nil {
		return schdList, err
	}

	var now = time.Now().UTC()
	var today string = now.Format("2006-01-02")
	var nextWeek string = now.Add(time.Hour * 24 * 7).Format("2006-01-02")

	for _, bareSchedule := range pdSchdList {
		schd := Schedule{
			Id:       bareSchedule.ID,
			Name:     bareSchedule.Name,
			Timezone: bareSchedule.TimeZone,
		}

		res, err := a.client.GetSchedule(bareSchedule.ID, pagerduty.GetScheduleOptions{
			TimeZone: a.timezone,
			Since:    today,
			Until:    nextWeek,
		})
		if err != nil {
			return schdList, err
		}

		var activeEntries int
		for _, se := range res.FinalSchedule.RenderedScheduleEntries {
			start, err := time.Parse(time.RFC3339Nano, se.Start)
			if err != nil {
				return schdList, err
			}

			end, err := time.Parse(time.RFC3339Nano, se.End)
			if err != nil {
				return schdList, err
			}

			log.WithFields(log.Fields{
				"scheduleId":   bareSchedule.ID,
				"scheduleName": bareSchedule.Name,
				"user":         se.User.ID,
				"start":        start,
				"end":          end,
			}).Debug("PagerDuty Schedule Entry")

			if start.Before(now) && end.After(now) {
				if activeEntries == 0 {
					schd.CurrentPeriod = &CallPeriod{
						Start: start,
						User:  se.User.ID,
					}
				}
				activeEntries += 1
			}

			if start.After(now) && (schd.NextPeriod == nil || start.Before(schd.NextPeriod.Start)) {
				schd.NextPeriod = &CallPeriod{
					Start: start,
					User:  se.User.ID,
				}
			}
		}

		lf := log.Fields{
			"scheduleId":   bareSchedule.ID,
			"scheduleName": bareSchedule.Name,
		}

		if schd.CurrentPeriod == nil {
			log.WithFields(lf).Warning("No active current period for schedule")
		} else {
			lf["currentCall"] = schd.CurrentPeriod.User
		}

		if schd.NextPeriod == nil {
			log.WithFields(lf).Warning("No active next period for schedule")
		} else {
			lf["nextCall"] = schd.NextPeriod.User
			lf["changeover"] = schd.NextPeriod.Start
		}

		if activeEntries > 1 {
			log.WithFields(lf).Warning("Multiple active schedules")
		}
		log.WithFields(lf).Debug("Got schedule entries")

		schdList = append(schdList, schd)
	}

	log.WithFields(log.Fields{
		"schedules": schdList,
	}).Debug("Known PagerDuty Schedules")

	return schdList, nil
}
