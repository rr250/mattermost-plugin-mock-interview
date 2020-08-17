package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type MockInterview struct {
	ID            string
	CreatedBy     string
	CreatedByID   string
	CreatedAt     time.Time
	InterviewType string
	Rating        string
	Language      string
	ScheduledAt   time.Time
	AcceptedBy    string
	AcceptedByID  string
	PostID        string
	IsAccepted    bool
	IsExpired     bool
}

type MockInterviewSummary struct {
	MockInterviewID string
}

func (p *Plugin) AddMockInterview(mockInterview MockInterview) interface{} {
	mockInterviewJSON, err := json.Marshal(mockInterview)
	if err != nil {
		p.API.LogError("failed to marshal mockInterview %s", mockInterview.ID)
		return fmt.Sprintf("failed to marshal mockInterview %s", mockInterview.ID)
	}
	err1 := p.API.KVSet("mock-"+mockInterview.ID, mockInterviewJSON)
	if err1 != nil {
		p.API.LogError("failed KVSet %s", err1, mockInterview)
		return fmt.Sprintf("failed KVSet %s", err1)
	}

	bytes, err2 := p.API.KVGet("mockinterviews")
	if err2 != nil {
		p.API.LogError("failed KVGet %s", err)
		return fmt.Sprintf("failed KVGet %s", err)
	}
	mockInterviewSummary := MockInterviewSummary{
		MockInterviewID: mockInterview.ID,
	}
	var mockInterviews []MockInterviewSummary
	if bytes != nil {
		if err = json.Unmarshal(bytes, &mockInterviews); err != nil {
			return fmt.Sprintf("failed to unmarshal  %s", err)
		}
		mockInterviews = append(mockInterviews, mockInterviewSummary)
	} else {
		mockInterviews = []MockInterviewSummary{mockInterviewSummary}
	}
	mockInterviewsJSON, err := json.Marshal(mockInterviews)
	if err != nil {
		p.API.LogError("failed to marshal mockInterviews  %s", mockInterviews)
		return fmt.Sprintf("failed to marshal mockInterviews  %s", mockInterviews)
	}
	err3 := p.API.KVSet("mockinterviews", mockInterviewsJSON)
	if err3 != nil {
		p.API.LogError("failed KVSet", err3, mockInterviewsJSON)
		return fmt.Sprintf("failed KVSet %s", err3)
	}
	return nil
}

func (p *Plugin) GetMockInterview(mockInterviewID string) (MockInterview, interface{}) {
	var mockInterview MockInterview
	bytes, err := p.API.KVGet("mock-" + mockInterviewID)
	if err != nil {
		p.API.LogError("failed KVGet %s", err)
		return mockInterview, fmt.Sprintf("failed to unmarshal %s", err)
	}
	if bytes != nil {
		if err3 := json.Unmarshal(bytes, &mockInterview); err3 != nil {
			return mockInterview, fmt.Sprintf("failed to unmarshal %s", err3)
		}
	} else {
		return mockInterview, "No MockInterview found"
	}
	return mockInterview, nil
}

func (p *Plugin) GetMockInterviewSummaryList() ([]MockInterviewSummary, interface{}) {
	var mockInterviewSummaryList []MockInterviewSummary
	bytes, err := p.API.KVGet("mockinterviews")
	if err != nil {
		p.API.LogError("failed KVGet %s", err)
		return mockInterviewSummaryList, fmt.Sprintf("failed to unmarshal %s", err)
	}
	if bytes != nil {
		if err3 := json.Unmarshal(bytes, &mockInterviewSummaryList); err3 != nil {
			return mockInterviewSummaryList, fmt.Sprintf("failed to unmarshal %s", err3)
		}
	} else {
		return mockInterviewSummaryList, "No MockInterview found"
	}
	return mockInterviewSummaryList, nil
}

func (p *Plugin) UpdateMockInterview(mockInterview MockInterview) interface{} {
	mockInterviewJSON, err := json.Marshal(mockInterview)
	if err != nil {
		p.API.LogError("failed to marshal mockInterview %s", mockInterview.ID)
		return fmt.Sprintf("failed to marshal mockInterview %s", mockInterview.ID)
	}
	err1 := p.API.KVSet("mock-"+mockInterview.ID, mockInterviewJSON)
	if err1 != nil {
		p.API.LogError("failed KVSet %s", err1, mockInterview)
		return fmt.Sprintf("failed KVSet %s", err1)
	}
	return nil
}
