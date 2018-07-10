package notion

import (
	"encoding/json"
)

// /api/v3/loadPageChunk request
type loadPageChunkRequest struct {
	PageID          string `json:"pageId"`
	Limit           int    `json:"limit"`
	Cursor          cursor `json:"cursor"`
	VerticalColumns bool   `json:"verticalColumns"`
}

type cursor struct {
	Stack [][]stack `json:"stack"`
}

type stack struct {
	Table string `json:"table"`
	ID    string `json:"id"`
	Index int    `json:"index"`
}

// /api/v3/loadPageChunk response
type loadPageChunkResponse struct {
	RecordMap recordMap `json:"recordMap"`
	Cursor    cursor    `json:"cursor"`
}

type recordMap struct {
	Blocks map[string]*BlockWithRole  `json:"block"`
	Space  map[string]interface{}     `json:"space"` // TODO: figure out the type
	Users  map[string]*notionUserInfo `json:"notion_user"`
	// TDOO: there might be more records types
}

type notionUserInfo struct {
	Role  string      `json:"role"`
	Value *notionUser `json:"value"`
}

type notionUser struct {
	Email                     string `json:"email"`
	FamilyName                string `json:"family_name"`
	GivenName                 string `json:"given_name"`
	ID                        string `json:"id"`
	Locale                    string `json:"locale"`
	MobileOnboardingCompleted bool   `json:"mobile_onboarding_completed"`
	OnboardingCompleted       bool   `json:"onboarding_completed"`
	ProfilePhoto              string `json:"profile_photo"`
	TimeZone                  string `json:"time_zone"`
	Version                   int64  `json:"version"`
}

func parseLoadPageChunk(d []byte) (*loadPageChunkResponse, error) {
	var rsp loadPageChunkResponse
	err := json.Unmarshal(d, &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}
