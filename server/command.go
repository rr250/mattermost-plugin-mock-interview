package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	command := strings.Trim(args.Command, " ")

	if strings.Trim(command, " ") == "/"+trigger {

		dialogRequest := model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s/createmockinterview", manifest.ID),
			Dialog: model.Dialog{
				Title:       "Request a Mock Interview",
				CallbackId:  model.NewId(),
				SubmitLabel: "Share",
				Elements: []model.DialogElement{
					{
						DisplayName: "Select Mock Interview Type",
						Name:        "interviewType",
						Type:        "select",
						SubType:     "select",
						Options: []*model.PostActionOptions{
							{
								Text:  "Programming DS/Algo",
								Value: "Programming DS/Algo",
							},
							{
								Text:  "Frontend",
								Value: "Frontend",
							},
							{
								Text:  "HLD",
								Value: "HLD",
							},
							{
								Text:  "LLD",
								Value: "LLD",
							},
							{
								Text:  "Tech Stack (Please specify the tech stack in the language field)",
								Value: "Tech Stack",
							},
						},
					},
					{
						DisplayName: "How will you rate yourself?",
						Name:        "rating",
						Type:        "select",
						SubType:     "select",
						Options: []*model.PostActionOptions{
							{
								Text:  "Beginner",
								Value: "Beginner",
							},
							{
								Text:  "Intermediate",
								Value: "Intermediate",
							},
							{
								Text:  "Expert",
								Value: "Expert",
							},
						},
					},
					{
						DisplayName: "Language",
						Name:        "language",
						Type:        "text",
						SubType:     "text",
						Optional:    true,
					},
					{
						DisplayName: "Specify a Date (format dd/mm/yy)",
						Name:        "date",
						Type:        "text",
						SubType:     "text",
					},
					{
						DisplayName: "Specify a time (format hh:mm, Timezone: IST (+5:30))",
						Name:        "time",
						Type:        "text",
						SubType:     "text",
					},
				},
			},
		}
		if pErr := p.API.OpenInteractiveDialog(dialogRequest); pErr != nil {
			p.API.LogError("Failed opening interactive dialog " + pErr.Error())
			p.SendEphermeral(args.UserId, args.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", pErr))
		}
	} else if strings.Trim(command, " ") == "/"+trigger+" help" {
		p.SendEphermeral(args.UserId, args.ChannelId, "* `/mockinterview` - opens up an [interactive dialog] to post a mock interview request")
	}

	return &model.CommandResponse{}, nil
}
