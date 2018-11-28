package updater

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Start the updater process
func (u *Updater) Start() {
	u.Wg.Add(1)
	go u.run()
}

// Loop for updater
// Will call for new data then call the update function
// Runs on each `updateEvery` interval
const updateEvery time.Duration = time.Minute * 5

func (u *Updater) run() {
	defer u.Wg.Done()

	for {
		start := time.Now().UTC()
		log.Debug("Update starting...")

		// fetch users and schedules in parallel
		w := sync.WaitGroup{}
		w.Add(2)
		go func() {
			defer w.Done()
			u.updateUsers()
		}()
		go func() {
			defer w.Done()
			u.updateSchedules()
		}()
		w.Wait()
		u.LastFetch = time.Now().UTC()

		// slack groups depend on users & schedules
		u.updateGroups()
		end := time.Now().UTC()

		log.WithFields(log.Fields{
			"duration":  end.Sub(start),
			"users":     len(u.Users.users),
			"schedules": len(u.Schedules.schedules),
		}).Info("Update complete")

		time.Sleep(updateEvery)
	}
}
