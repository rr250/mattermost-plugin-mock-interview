package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func (p *Plugin) InitAPI() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/createmockinterview", p.CreateMockInterview).Methods("POST")
	r.HandleFunc("/acceptrequest", p.AcceptRequest).Methods("POST")
	r.HandleFunc("/cancelmockinterviewbyid", p.CancelMockInterview).Methods("POST")
	r.HandleFunc("/editmockinterviewbyid", p.EditMockInterview).Methods("POST")
	r.HandleFunc("/editmockinterviewsubmit", p.EditMockInterviewSubmit).Methods("POST")
	return r
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) CreateMockInterview(w http.ResponseWriter, req *http.Request) {

	request := model.SubmitDialogRequestFromJson(req.Body)

	user, err := p.API.GetUser(request.UserId)
	if err != nil {
		p.API.LogError("Unable to get User", err)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err))
		return
	}

	mockInterview := MockInterview{
		ID:            model.NewId(),
		CreatedBy:     user.GetFullName(),
		CreatedByID:   user.Id,
		CreatedAt:     time.Now(),
		InterviewType: fmt.Sprintf("%v", request.Submission["interviewType"]),
		Rating:        fmt.Sprintf("%v", request.Submission["rating"]),
		Language:      fmt.Sprintf("%v", request.Submission["language"]),
		IsAccepted:    false,
		IsExpired:     false,
	}

	date := fmt.Sprintf("%v", request.Submission["date"])
	t1 := fmt.Sprintf("%v", request.Submission["time"])
	value := strings.Trim(date, " ") + ", " + strings.Trim(t1, " ") + ", +0530"
	p.API.LogInfo(value)
	layout := "02/01/06, 15:04, -0700"
	t, err2 := time.Parse(layout, value)
	if err2 != nil || t.Before(time.Now()) {
		p.API.LogError("Not a valid time", err2)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Not a valid time %s", err2))
		return
	}
	configuration := p.getConfiguration()
	mockInterview.ScheduledAt = t
	postModel := &model.Post{
		UserId:    p.botUserID,
		ChannelId: configuration.ChannelID,
		Props: model.StringInterface{
			"attachments": []*model.SlackAttachment{
				{
					Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nRating: " + mockInterview.Rating + "\nLanguage: " + mockInterview.Language,
					Actions: []*model.PostAction{
						{
							Integration: &model.PostActionIntegration{
								URL: fmt.Sprintf("/plugins/%s/acceptrequest", manifest.ID),
								Context: model.StringInterface{
									"action":          "acceptrequest",
									"mockinterviewid": mockInterview.ID,
								},
							},
							Type: model.POST_ACTION_TYPE_BUTTON,
							Name: "Accept Request",
						},
					},
				},
			},
		},
	}

	post, err := p.API.CreatePost(postModel)
	if err != nil {
		p.API.LogError("Unable to create post", err)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err))
		return
	}
	mockInterview.PostID = post.Id
	err1 := p.AddMockInterview(mockInterview)
	if err1 != nil {
		p.API.LogError("", err1.(string))
		p.API.DeletePost(mockInterview.PostID)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("%s", err1.(string)))
		return
	}
}

func (p *Plugin) AcceptRequest(w http.ResponseWriter, req *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(req.Body)
	writePostActionIntegrationResponseOk(w, &model.PostActionIntegrationResponse{})
	user, err := p.API.GetUser(request.UserId)
	if err != nil {
		p.API.LogError("Unable to get User", err)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err))
		return
	}
	mockInterviewID := request.Context["mockinterviewid"].(string)
	mockInterview, err1 := p.GetMockInterview(mockInterviewID)
	if err1 != nil {
		p.API.LogError("", err1.(string))
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("%s", err1.(string)))
		return
	}
	if mockInterview.ScheduledAt.Before(time.Now()) || mockInterview.IsAccepted {
		p.SendEphermeral(request.UserId, request.ChannelId, "Request Expired")
	} else if mockInterview.CreatedByID == user.Id {
		p.SendEphermeral(request.UserId, request.ChannelId, "Can't accept your own request")
	} else {
		mockInterview.AcceptedBy = user.GetFullName()
		mockInterview.AcceptedByID = user.Id
		mockInterview.IsAccepted = true
		err1 = p.UpdateMockInterview(mockInterview)
		if err1 != nil {
			p.API.LogError("", err1)
			p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("%s", err1))
			return
		}
		channel, err := p.API.GetDirectChannel(mockInterview.AcceptedByID, mockInterview.CreatedByID)
		if err != nil {
			p.API.LogError("Unable to get Channel", err)
			p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err))
			return
		}
		postModel := &model.Post{
			UserId:    p.botUserID,
			ChannelId: channel.Id,
			Message:   "Mock Interview has been scheduled between " + mockInterview.CreatedBy + " and " + mockInterview.AcceptedBy + ".\nPlease follow the guidelines: https://docs.google.com/spreadsheets/d/1HAYsoH-wJoDZ3Sihdlb5HTBxngB7zNuDTci7-qfZbX0/edit?usp=sharing",
			Props: model.StringInterface{
				"attachments": []*model.SlackAttachment{
					{
						Text: "Mock Interview Scheduled At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nRating: " + mockInterview.Rating + "\nLanguage: " + mockInterview.Language + "\nAccepted By: " + mockInterview.AcceptedBy,
					},
				},
			},
		}

		p.API.CreatePost(postModel)
		post, err := p.API.GetPost(request.PostId)
		if err != nil {
			p.API.LogError("Unable to get Post", err)
			p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err))
			return
		}
		post.Props = model.StringInterface{
			"attachments": []*model.SlackAttachment{
				{
					Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nRating: " + mockInterview.Rating + "\nLanguage: " + mockInterview.Language + "\nAccepted By: " + mockInterview.AcceptedBy,
				},
			},
		}
		_, err = p.API.UpdatePost(post)
		if err != nil {
			p.API.LogError("Unable to update Post", err)
			p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err))
			return
		}
	}
}

func (p *Plugin) EditMockInterview(w http.ResponseWriter, req *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(req.Body)
	mockInterviewID := request.Context["mockinterviewid"].(string)
	mockInterview, err1 := p.GetMockInterview(mockInterviewID)
	if err1 != nil {
		p.API.LogError("failed to get mock interview", err1)
		postModel := &model.Post{
			UserId:    request.UserId,
			ChannelId: request.ChannelId,
			Message:   "Failed to get mockinterview",
		}
		p.API.SendEphemeralPost(request.UserId, postModel)
		return
	}
	dialogRequest := model.OpenDialogRequest{
		TriggerId: request.TriggerId,
		URL:       fmt.Sprintf("/plugins/%s/editmockinterviewsubmit", manifest.ID),
		Dialog: model.Dialog{
			Title:       "Edit Mock Interview",
			CallbackId:  model.NewId(),
			SubmitLabel: "Edit",
			State:       mockInterviewID,
			Elements: []model.DialogElement{
				{
					DisplayName: "Select Mock Interview Type",
					Name:        "interviewType",
					Type:        "select",
					SubType:     "select",
					Default:     mockInterview.InterviewType,
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
					Default:     mockInterview.Rating,
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
					Default:     mockInterview.Language,
					Optional:    true,
				},
				{
					DisplayName: "Specify a Date",
					Name:        "date",
					Type:        "text",
					SubType:     "text",
					Placeholder: "dd/mm/yy",
					Default:     mockInterview.ScheduledAt.Format("02/01/06"),
				},
				{
					DisplayName: "Specify a time ((24hrs format), (Timezone: IST +5:30))",
					Name:        "time",
					Type:        "text",
					SubType:     "text",
					Placeholder: "HH:mm",
					Default:     mockInterview.ScheduledAt.Format("15:04"),
				},
			},
		},
	}
	if pErr := p.API.OpenInteractiveDialog(dialogRequest); pErr != nil {
		p.API.LogError("Failed opening interactive dialog " + pErr.Error())
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", pErr))
	}
	writePostActionIntegrationResponseOk(w, &model.PostActionIntegrationResponse{})
}

func (p *Plugin) EditMockInterviewSubmit(w http.ResponseWriter, req *http.Request) {

	request := model.SubmitDialogRequestFromJson(req.Body)
	date := fmt.Sprintf("%v", request.Submission["date"])
	t1 := fmt.Sprintf("%v", request.Submission["time"])
	value := strings.Trim(date, " ") + ", " + strings.Trim(t1, " ") + ", +0530"
	p.API.LogInfo(value)
	layout := "02/01/06, 15:04, -0700"
	t, err2 := time.Parse(layout, value)
	if err2 != nil || t.Before(time.Now()) {
		p.API.LogError("Not a valid time", err2)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Not a valid time %s", err2))
		return
	}
	mockInterview, err1 := p.GetMockInterview(request.State)
	if err1 != nil {
		p.API.LogError("", err1.(string))
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("%s", err1.(string)))
		return
	}
	if mockInterview.IsExpired || mockInterview.IsCancelled || mockInterview.IsAccepted {
		p.SendEphermeral(request.UserId, request.ChannelId, "Mock Interview is either expired, cancelled or accepted")
		return
	}
	mockInterview.InterviewType = fmt.Sprintf("%v", request.Submission["interviewType"])
	mockInterview.Rating = fmt.Sprintf("%v", request.Submission["rating"])
	mockInterview.Language = fmt.Sprintf("%v", request.Submission["language"])
	mockInterview.ScheduledAt = t
	err3 := p.UpdateMockInterview(mockInterview)
	if err3 != nil {
		p.API.LogError("", err3)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("%s", err3))
		return
	}
	post, err4 := p.API.GetPost(mockInterview.PostID)
	if err4 != nil {
		p.API.LogError("Unable to get Post", err4)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err4))
		return
	}
	post.Props = model.StringInterface{
		"attachments": []*model.SlackAttachment{
			{
				Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nRating: " + mockInterview.Rating + "\nLanguage: " + mockInterview.Language,
				Actions: []*model.PostAction{
					{
						Integration: &model.PostActionIntegration{
							URL: fmt.Sprintf("/plugins/%s/acceptrequest", manifest.ID),
							Context: model.StringInterface{
								"action":          "acceptrequest",
								"mockinterviewid": mockInterview.ID,
							},
						},
						Type: model.POST_ACTION_TYPE_BUTTON,
						Name: "Accept Request",
					},
				},
			},
		},
	}
	_, err5 := p.API.UpdatePost(post)
	if err5 != nil {
		p.API.LogError("Unable to update Post", err5)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err5))
		return
	}
	p.SendEphermeral(request.UserId, request.ChannelId, "Mock interview Request is edited")
}

func (p *Plugin) CancelMockInterview(w http.ResponseWriter, req *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(req.Body)
	writePostActionIntegrationResponseOk(w, &model.PostActionIntegrationResponse{})
	mockInterviewID := request.Context["mockinterviewid"].(string)
	mockInterview, err1 := p.GetMockInterview(mockInterviewID)
	if err1 != nil {
		p.API.LogError("", err1.(string))
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("%s", err1.(string)))
		return
	}
	if mockInterview.IsAccepted {
		p.SendEphermeral(request.UserId, request.ChannelId, "Request already accepted")
		return
	}
	if mockInterview.IsExpired {
		p.SendEphermeral(request.UserId, request.ChannelId, "Request already expired")
		return
	}
	mockInterview.IsCancelled = !mockInterview.IsCancelled
	err2 := p.UpdateMockInterview(mockInterview)
	if err2 != nil {
		p.API.LogError("", err2)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("%s", err2))
		return
	}
	post, err3 := p.API.GetPost(mockInterview.PostID)
	if err3 != nil {
		p.API.LogError("Unable to get Post", err3)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err3))
		return
	}
	if mockInterview.IsCancelled {
		post.Props = model.StringInterface{
			"attachments": []*model.SlackAttachment{
				{
					Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nRequest Cancelled",
				},
			},
		}
	} else {
		post.Props = model.StringInterface{
			"attachments": []*model.SlackAttachment{
				{
					Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nRating: " + mockInterview.Rating + "\nLanguage: " + mockInterview.Language,
					Actions: []*model.PostAction{
						{
							Integration: &model.PostActionIntegration{
								URL: fmt.Sprintf("/plugins/%s/acceptrequest", manifest.ID),
								Context: model.StringInterface{
									"action":          "acceptrequest",
									"mockinterviewid": mockInterview.ID,
								},
							},
							Type: model.POST_ACTION_TYPE_BUTTON,
							Name: "Accept Request",
						},
					},
				},
			},
		}
	}
	_, err4 := p.API.UpdatePost(post)
	if err4 != nil {
		p.API.LogError("Unable to update Post", err4)
		p.SendEphermeral(request.UserId, request.ChannelId, fmt.Sprintf("Some Error happened. Try Again %s", err4))
		return
	}
}

func writePostActionIntegrationResponseOk(w http.ResponseWriter, response *model.PostActionIntegrationResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.ToJson())
}

func (p *Plugin) SendEphermeral(userID string, channelID string, message string) {
	postModel := &model.Post{
		UserId:    userID,
		ChannelId: channelID,
		Message:   message,
	}
	p.API.SendEphemeralPost(userID, postModel)
}
