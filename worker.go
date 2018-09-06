package inasnap

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ryanjyoder/couchdb"
)

const (
	statusPending    = ""
	statusSettingUp  = "setting_up"
	statusInstalling = "installing"
	statusRunning    = "running"
	statusFailed     = "failed"
)

type Worker interface {
	Run() error
}

type worker struct {
	couchdb couchdb.DatabaseService
}

func NewWorker(configs Configs) (Worker, error) {
	s := worker{}

	u, err := url.Parse(configs.CouchURL)
	if err != nil {
		return nil, err
	}
	client, err := couchdb.NewAuthClient(configs.CouchUser, configs.CouchPassword, u)
	if err != nil {
		return nil, err
	}

	s.couchdb = client.Use(configs.CouchDBName)

	return &s, nil

}

func (s *worker) Run() error {
	for {
		app, err := s.getNextPendingApp()
		if app == nil || err != nil {
			if err != nil {
				log.Println("failed to get pending apps:", err)
			}
			// replace with call to _changes
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("setting up new lxd continaer")
		app.Status = statusSettingUp
		resp, err := s.couchdb.Put(app)
		if err != nil {
			log.Println("failed to update job status")
			continue
		}
		app.Rev = resp.Rev
		id, err := newLXDContainer()
		if err != nil {
			app.Status = statusFailed
			app.Error = err.Error()
			s.couchdb.Put(app)
			log.Println("setting up lxd failed")
			continue
		}
		app.LxdContainerID = id
		app.Status = statusInstalling
		resp, err = s.couchdb.Put(app)
		if err != nil {
			log.Println("failed to update job status")
			continue
		}
		app.Rev = resp.Rev

	}
	return nil
}

func newLXDContainer() (string, error) {
	return "", fmt.Errorf("LXD not yet implemented")
}

func (s *worker) getNextPendingApp() (app *appJob, err error) {
	key := `""`
	reduce := false
	limit := 1
	resp, err := s.couchdb.View("apps").Get("status-id", couchdb.QueryParameters{
		Key:    &key,
		Reduce: &reduce,
		Limit:  &limit,
	})
	if err != nil {
		return
	}
	for i := range resp.Rows {
		if appID, ok := resp.Rows[i].Value.(string); ok {
			app, err = s.getJob(appID)
			return
		}
	}
	return
}

func (w *worker) getJob(appID string) (*appJob, error) {
	job := &appJob{}
	return job, w.couchdb.Get(job, appID)
}
