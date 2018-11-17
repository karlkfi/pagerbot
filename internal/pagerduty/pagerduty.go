package pagerduty

import (
	"fmt"
	"github.com/PagerDuty/go-pagerduty"
	log "github.com/Sirupsen/logrus"
)

type Api struct {
	key      string
	org      string
	client   *pagerduty.Client
	timezone string
}

// Pagerduty API doesn't provide a sane way of checking for auth
// so we just get the schedules at setup time
func New(key string, org string) (*Api, error) {
	a := Api{}
	a.key = key
	a.org = org
	a.timezone = "UTC"

	a.client = pagerduty.NewClient(key)

	_, err := a.client.ListSchedules(pagerduty.ListSchedulesOptions{})
	if err != nil {
		return &a, fmt.Errorf("Failed to list on-call schedules: %v", err)
	}

	log.Info("Authenticated with PagerDuty")

	return &a, nil
}
