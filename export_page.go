package notionapi

import (
	"fmt"
	"time"
)

const (
	eventExportBlock      = "exportBlock"
	defaultExportTimeZone = "America/Los_Angeles"
	statusComplete        = "complete"
	ExportTypeMarkdown    = "markdown"
	ExportTypeHTML        = "html"
)

type exportPageTaskRequest struct {
	Task *exportPageTask `json:"task"`
}

type exportPageTask struct {
	EventName string             `json:"eventName"`
	Request   *exportPageRequest `json:"request"`
}

type exportPageBlock struct {
	ID string `json:"id"`
}

type exportPageRequest struct {
	Block                exportPageBlock    `json:"block"`
	Recursive            bool               `json:"recursive"`
	ShouldExportComments bool               `json:"shouldExportComments"`
	ExportOptions        *exportPageOptions `json:"exportOptions"`
}

type exportPageOptions struct {
	ExportType string `json:"exportType"`
	TimeZone   string `json:"timeZone"`
	Locale     string `json:"locale"`
}

type enqueueTaskResponse struct {
	TaskID  string                 `json:"taskId"`
	RawJSON map[string]interface{} `json:"-"`
}

type getTasksExportPageResponse struct {
	Results []*exportPageResult `json:"results"`
}

type exportPageResult struct {
	ID        string             `json:"id"`
	EventName string             `json:"eventName"`
	Request   *exportPageRequest `json:"request"`
	UserID    string             `json:"userId"`
	State     string             `json:"state"`
	Status    *exportPageStatus  `json:"status"`
}

type exportPageStatus struct {
	Type          string `json:"type"`
	ExportURL     string `json:"exportURL"`
	PagesExported int64  `json:"pagesExported"`
}

const (
	taskStateSuccess = "success"
	taskStateFailure = "failure"
	// how long to wait for export task to complete
	exportTaskTimeout = 5 * time.Minute
)

type getTasksRequest struct {
	TaskIDS []string `json:"taskIds"`
}

// RequestPageExportURL executes a raw API call to enqueue an export of pages
// and returns the URL to the exported data once the task is complete
func (c *Client) RequestPageExportURL(id string, exportType string, recursive bool) (string, error) {
	id = ToDashID(id)
	if !IsValidDashID(id) {
		return "", fmt.Errorf("'%s' is not a valid notion id", id)
	}

	req := &exportPageTaskRequest{
		Task: &exportPageTask{
			EventName: eventExportBlock,
			Request: &exportPageRequest{
				Block:     exportPageBlock{ID: id},
				Recursive: recursive,
				ExportOptions: &exportPageOptions{
					ExportType: exportType,
					TimeZone:   defaultExportTimeZone,
					Locale:     "en",
				},
			},
		},
	}

	var rsp enqueueTaskResponse
	var err error
	apiURL := "/api/v3/enqueueTask"
	err = c.doNotionAPI(apiURL, req, &rsp, &rsp.RawJSON)
	if err != nil {
		return "", err
	}

	var exportURL string
	taskID := rsp.TaskID
	timeStart := time.Now()
	for {
		if time.Since(timeStart) > exportTaskTimeout {
			return "", fmt.Errorf("export task '%s' didn't complete in %s", taskID, exportTaskTimeout)
		}
		time.Sleep(250 * time.Millisecond)
		req := getTasksRequest{
			TaskIDS: []string{taskID},
		}
		var err error
		var rsp getTasksExportPageResponse
		apiURL = "/api/v3/getTasks"
		err = c.doNotionAPI(apiURL, req, &rsp, nil)
		if err != nil {
			return "", err
		}
		if len(rsp.Results) == 0 {
			return "", fmt.Errorf("getTasks returned no results for task '%s'", taskID)
		}
		res := rsp.Results[0]
		if res.State == taskStateFailure {
			return "", fmt.Errorf("export task '%s' failed", taskID)
		}
		status := res.Status
		if status != nil && status.ExportURL != "" && (status.Type == statusComplete || res.State == taskStateSuccess) {
			exportURL = status.ExportURL
			break
		}
		time.Sleep(750 * time.Millisecond)
	}

	return exportURL, nil
}

// ExportPages exports a page as html or markdown, potentially recursively.
// Requires Client.AuthToken.
func (c *Client) ExportPages(id string, exportType string, recursive bool) ([]byte, error) {
	exportURL, err := c.RequestPageExportURL(id, exportType, recursive)
	if err != nil {
		return nil, err
	}

	dlRsp, err := c.DownloadFile(exportURL, nil)
	if err != nil {
		return nil, err
	}
	return dlRsp.Data, nil
}
