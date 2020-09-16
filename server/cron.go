package main

import (
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/robfig/cron/v3"
)

func (p *Plugin) InitCRON() *cron.Cron {
	c := cron.New()
	c.AddFunc("@every 15m", p.DeleteApplyButton)
	return c
}

func (p *Plugin) DeleteApplyButton() {
	p.API.LogInfo("Scheduler started")
	now := time.Now()
	mockInterviewSummaryList, err := p.GetMockInterviewSummaryList()
	if err != nil {
		p.API.LogError("", err)
		return
	}
	for _, mockInterviewSummary := range mockInterviewSummaryList {
		mockInterview, err := p.GetMockInterview(mockInterviewSummary.MockInterviewID)
		if err != nil {
			p.API.LogError("", err)
			continue
		}
		if mockInterview.ScheduledAt.Before(now) && !mockInterview.IsExpired {
			mockInterview.IsExpired = true
			err1 := p.UpdateMockInterview(mockInterview)
			if err1 != nil {
				p.API.LogError("", err1)
			}
			post, err := p.API.GetPost(mockInterview.PostID)
			if err != nil {
				p.API.LogError("Unable to get post", err)
				continue
			}
			post.Props = model.StringInterface{
				"attachments": []*model.SlackAttachment{
					{
						Text: "\nMock Interview Posted By: " + mockInterview.CreatedBy + "\nInterview Type: " + mockInterview.InterviewType + "\nRequest Expired",
					},
				},
			}
			_, err = p.API.UpdatePost(post)
			if err != nil {
				p.API.LogError("Unable to update post", err)
			}
		}
	}
}
