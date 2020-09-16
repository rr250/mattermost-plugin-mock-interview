package main

import (
	"fmt"
	"strings"
	"time"

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
						DisplayName: "Specify a Date",
						Name:        "date",
						Type:        "text",
						SubType:     "text",
						Placeholder: "dd/mm/yy",
					},
					{
						DisplayName: "Specify a time ((24hrs format), (Timezone: IST +5:30))",
						Name:        "time",
						Type:        "text",
						SubType:     "text",
						Placeholder: "HH:mm",
					},
				},
			},
		}
		if pErr := p.API.OpenInteractiveDialog(dialogRequest); pErr != nil {
			p.API.LogError("Failed opening interactive dialog " + pErr.Error())
			p.SendEphermeral(args.UserId, args.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", pErr))
		}
	} else if strings.Trim(command, " ") == "/"+trigger+" list" {
		mockInterviewPerUserList, err := p.GetMockInterviewPerUserList(args.UserId)
		if err == nil {
			postModel := &model.Post{
				UserId:    args.UserId,
				ChannelId: args.ChannelId,
				Message:   "Mock Interviews :-",
				Props: model.StringInterface{
					"attachments": []*model.SlackAttachment{},
				},
			}
			for _, mockInterview := range mockInterviewPerUserList {
				mockInterview1, err1 := p.GetMockInterview(mockInterview.MockInterviewID)
				if err1 != nil {
					p.API.LogError("", err1.(string))
					p.SendEphermeral(args.UserId, args.ChannelId, fmt.Sprintf("%s", err1.(string)))
					continue
				}
				attachment := &model.SlackAttachment{
					Text: "Mock Interview: " + mockInterview1.InterviewType + "\nCreatedAt: " + mockInterview1.CreatedAt.Format(time.ANSIC) + "\nIs Cancelled: " + fmt.Sprintf("%t", mockInterview1.IsCancelled) + "\nIs Expired: " + fmt.Sprintf("%t", mockInterview1.IsExpired),
					Actions: []*model.PostAction{
						{
							Integration: &model.PostActionIntegration{
								URL: fmt.Sprintf("/plugins/%s/cancelmockinterviewbyid", manifest.ID),
								Context: model.StringInterface{
									"action":          "cancelmockinterviewbyid",
									"mockinterviewid": mockInterview.MockInterviewID,
								},
							},
							Type: model.POST_ACTION_TYPE_BUTTON,
							Name: "Cancel/Uncancel Request",
						},
						{
							Integration: &model.PostActionIntegration{
								URL: fmt.Sprintf("/plugins/%s/editmockinterviewbyid", manifest.ID),
								Context: model.StringInterface{
									"action":          "editmockinterviewbyid",
									"mockinterviewid": mockInterview.MockInterviewID,
								},
							},
							Type: model.POST_ACTION_TYPE_BUTTON,
							Name: "Edit",
						},
					},
				}
				postModel.Props["attachments"] = append(postModel.Props["attachments"].([]*model.SlackAttachment), attachment)
			}
			p.API.SendEphemeralPost(args.UserId, postModel)

		} else {
			postModel := &model.Post{
				UserId:    args.UserId,
				ChannelId: args.ChannelId,
				Message:   err.(string),
			}
			p.API.SendEphemeralPost(args.UserId, postModel)
		}
	} else if strings.Trim(command, " ") == "/"+trigger+" help" {
		p.SendEphermeral(args.UserId, args.ChannelId, "* `/mockinterview` - opens up an [interactive dialog] to post a mock interview request\n* `/mockinterview list` - get all your mock interview requests which you can edit or cancel")
	}

	return &model.CommandResponse{}, nil
}
