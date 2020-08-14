package main

import (
	"github.com/robfig/cron/v3"
)

func (p *Plugin) InitCRON() *cron.Cron {
	c := cron.New()
	c.AddFunc("@every 1m", func() { p.API.LogInfo("CRONNNNNNNNNN Every hour on the half hour") })
	return c
}
