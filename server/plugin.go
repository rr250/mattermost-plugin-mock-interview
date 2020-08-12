package main

import (
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/robfig/cron/v3"
)

type Plugin struct {
	plugin.MattermostPlugin
	router            *mux.Router
	botUserID         string
	configurationLock sync.RWMutex
	configuration     *configuration
	cron              *cron.Cron
}
