package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const (
	trigger = "mockinterview"
)

const (
	botUserName    = "mockinterviewbot"
	botDisplayName = "Mock Interview Bot"
)

// OnActivate register the plugin command
func (p *Plugin) OnActivate() error {
	p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Command for Mock Interview Plugin",
		DisplayName:      "Command for Mock Interview Plugin",
		AutoComplete:     true,
		AutoCompleteDesc: "Type /mock and press enter. For more commands type /mock help",
		AutoCompleteHint: "Command for Mock Interview Plugin",
	})
	botUserID, err := p.ensureBotExists()
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot user")
	}
	p.botUserID = botUserID
	p.router = p.InitAPI()
	p.cron = p.InitCRON()
	p.cron.Start()
	return nil
}

func (p *Plugin) ensureBotExists() (string, error) {
	bot := &model.Bot{
		Username:    botUserName,
		DisplayName: botDisplayName,
	}

	return p.Helpers.EnsureBot(bot)
}
