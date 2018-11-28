package updater

import (
	"github.com/karlkfi/pagerbot/internal/pagerduty"
)

type ScheduleList struct {
	schedules []*pagerduty.Schedule
}

func (s *ScheduleList) ById(id string) *pagerduty.Schedule {
	for _, schedule := range s.schedules {
		if schedule.Id == id {
			return schedule
		}
	}
	return nil
}
