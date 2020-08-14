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
	}

	mockInterview := MockInterview{
		ID:            model.NewId(),
		CreatedBy:     user.GetFullName(),
		CreatedByID:   user.Id,
		CreatedAt:     time.Now(),
		InterviewType: request.Submission["interviewType"].(string),
		Language:      request.Submission["language"].(string),
		AcceptedBy:    "na",
		AcceptedByID:  "na",
		IsAccepted:    false,
		IsExpired:     false,
	}

	date := request.Submission["date"].(string)
	t1 := request.Submission["time"].(string)
	value := strings.Trim(date, " ") + ", " + strings.Trim(t1, " ") + ", +0530"
	layout := "02/01/06, 15:04, -0700"
	t, _ := time.Parse(layout, value)
	mockInterview.ScheduledAt = t
	postModel := &model.Post{
		UserId:    p.botUserID,
		ChannelId: request.ChannelId,
		Props: model.StringInterface{
			"attachments": []*model.SlackAttachment{
				{
					Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nLanguage: " + mockInterview.Language,
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
	}
	mockInterview.PostID = post.Id
	err1 := p.AddMockInterview(mockInterview)
	if err1 != nil {
		p.API.LogError("", err1.(string))
		p.API.DeletePost(mockInterview.PostID)
	}
}

func (p *Plugin) AcceptRequest(w http.ResponseWriter, req *http.Request) {
	request := model.PostActionIntegrationRequestFromJson(req.Body)
	writePostActionIntegrationResponseOk(w, &model.PostActionIntegrationResponse{})
	user, err := p.API.GetUser(request.UserId)
	if err != nil {
		p.API.LogError("Unable to get User", err)
	}
	mockInterviewID := request.Context["mockinterviewid"].(string)
	mockInterview, err1 := p.GetMockInterview(mockInterviewID)
	if err1 != nil {
		p.API.LogError("", err1.(string))
	}
	mockInterview.AcceptedBy = user.GetFullName()
	mockInterview.AcceptedByID = user.Id
	mockInterview.IsAccepted = true
	channel, err := p.API.GetDirectChannel(mockInterview.AcceptedByID, mockInterview.CreatedByID)
	if err != nil {
		p.API.LogError("Unable to get Channel", err)
	}
	postModel := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channel.Id,
		Message:   mockInterview.AcceptedBy + " accepted your request of mock interview. Here are the details:-",
		Props: model.StringInterface{
			"attachments": []*model.SlackAttachment{
				{
					Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nLanguage: " + mockInterview.Language + "\nAccepted By: " + mockInterview.AcceptedBy,
				},
			},
		},
	}

	p.API.CreatePost(postModel)
	post, err := p.API.GetPost(request.PostId)
	if err != nil {
		p.API.LogError("Unable to get Post", err)
	}
	post.Props = model.StringInterface{
		"attachments": []*model.SlackAttachment{
			{
				Text: "Mock Interview Request At : " + mockInterview.ScheduledAt.Format(time.RFC822) + "\nPosted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nLanguage: " + mockInterview.Language + "\nAccepted By: " + mockInterview.AcceptedBy,
			},
		},
	}
	_, err = p.API.UpdatePost(post)
	if err != nil {
		p.API.LogError("Unable to update Post", err)
	}
}

func writePostActionIntegrationResponseOk(w http.ResponseWriter, response *model.PostActionIntegrationResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response.ToJson())
}
