package notionapi

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
)

// POST /api/v3/getUploadFileUrl request
type getUploadFileUrlRequest struct {
	Bucket      string `json:"bucket"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

// GetUploadFileUrlResponse is a response to POST /api/v3/getUploadFileUrl
type GetUploadFileUrlResponse struct {
	URL          string `json:"url"`
	SignedGetURL string `json:"signedGetUrl"`
	SignedPutURL string `json:"signedPutUrl"`

	FileID string `json:"-"`

	RawJSON map[string]interface{} `json:"-"`
}

func (r *GetUploadFileUrlResponse) Parse() {
	r.FileID = strings.Split(r.URL[len(s3URLPrefix):], "/")[0]
}

// getUploadFileURL executes a raw API call: POST /api/v3/getUploadFileUrl
func (c *Client) getUploadFileURL(name, contentType string) (*GetUploadFileUrlResponse, error) {
	const apiURL = "/api/v3/getUploadFileUrl"

	req := &getUploadFileUrlRequest{
		Bucket:      "secure",
		ContentType: contentType,
		Name:        name,
	}

	var rsp GetUploadFileUrlResponse
	var err error
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}

	rsp.Parse()

	return &rsp, nil
}

// GetFileContentType tries to figure out the content type of the file using http detection
func GetFileContentType(file *os.File) (contentType string, err error) {
	// Try using the extension to figure out the file's type
	ext := path.Ext(file.Name())
	contentType = mime.TypeByExtension(ext)
	if contentType != "" {
		return
	}

	// Seek the file to the start once done
	defer func() {
		_, err2 := file.Seek(0, 0)
		if err == nil && err2 != nil {
			err = fmt.Errorf("error seeking start of file: %s", err2)
		}
	}()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = file.Read(buffer)
	if err != nil {
		return
	}

	// Use the net/http package's handy DetectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType = http.DetectContentType(buffer)
	return
}

// TODO: Support adding new records to collections and other non-block parent tables
// SetNewRecordOp creates an operation to create a new record
func (c *Client) SetNewRecordOp(userID string, parent *Block, recordType string) (newBlock *Block, operation *Operation) {
	newID := uuid.New().String()
	now := Now()

	newBlock = &Block{
		ID:          newID,
		Version:     1,
		Alive:       true,
		Type:        recordType,
		CreatedBy:   userID,
		CreatedTime: now,
		ParentID:    parent.ID,
		ParentTable: "block",
	}

	operation = newBlock.buildOp(CommandSet, []string{}, map[string]interface{}{
		"id":           newBlock.ID,
		"version":      newBlock.Version,
		"alive":        newBlock.Alive,
		"type":         newBlock.Type,
		"created_by":   newBlock.CreatedBy,
		"created_time": newBlock.CreatedTime,
		"parent_id":    newBlock.ParentID,
		"parent_table": newBlock.ParentTable,
	})

	return
}

// UploadFile Uploads a file to notion's asset hosting(aws s3)
func (c *Client) UploadFile(file *os.File) (fileID, fileURL string, err error) {
	contentType, err := GetFileContentType(file)
	log(c, "contentType: %s", contentType)

	if err != nil {
		err = fmt.Errorf("couldn't figure out the content-type of the file: %s", err)
		return
	}

	fi, err := file.Stat()
	if err != nil {
		err = fmt.Errorf("error getting file's stats: %s", err)
		return
	}

	fileSize := fi.Size()

	// 1. getUploadFileURL
	uploadFileURLResp, err := c.getUploadFileURL(file.Name(), contentType)
	if err != nil {
		err = fmt.Errorf("get upload file URL error: %s", err)
		return
	}

	// 2. Upload file to amazon - PUT
	httpClient := c.getHTTPClient()

	req, err := http.NewRequest(http.MethodPut, uploadFileURLResp.SignedPutURL, file)
	if err != nil {
		return
	}
	req.ContentLength = fileSize
	req.TransferEncoding = []string{"identity"} // disable chunked (unsupported by aws)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var contents []byte
		contents, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			contents = []byte(fmt.Sprintf("Error from ReadAll: %s", err))
		}

		err = fmt.Errorf("http PUT '%s' failed with status %s: %s", req.URL, resp.Status, string(contents))
		return
	}

	return uploadFileURLResp.FileID, uploadFileURLResp.URL, nil
}

// EmbedFile creates a set of operations to embed a file into a block
func (b *Block) EmbedUploadedFileOps(client *Client, userID, fileID, fileURL string) (*Block, []*Operation) {
	newBlock, newBlockOp := client.SetNewRecordOp(userID, b, BlockEmbed)
	ops := []*Operation{
		newBlockOp,
		b.UpdateOp(&Block{LastEditedTime: Now(), LastEditedBy: userID}),
	}
	ops = append(ops, newBlock.embeddedFileOps(fileID, fileURL)...)

	/* TODO: Set size of image/video embeds
	newBlock.UpdateFormatOp(&FormatImage{
		BlockWidth: width,
		BlockHeight: height,
		BlockPreserveScale: true,
		BlockFullWidth: true,
		BlockPageWidth: false,
		BlockAspectRatio: float64(width) / float64(height),
	}),
	*/

	return newBlock, ops
}

// embeddedFileOps creates a set of operations to update the embedded file
func (b *Block) embeddedFileOps(fileID, fileURL string) []*Operation {
	if !b.IsEmbeddedType() {
		return nil
	}

	return []*Operation{
		b.UpdatePropertiesOp(fileURL),
		b.UpdateFormatOp(&FormatEmbed{DisplaySource: fileURL}),
		// TODO: Update block type based on upload
		//b.UpdateOp(&Block{Type: BlockImage}),
		b.ListAfterFileIDsOp(fileID),
	}
}

// UpdateEmbeddedFileOps creates a set of operations to update an existing embedded file
func (b *Block) UpdateEmbeddedFileOps(userID, fileID, fileURL string) []*Operation {
	if !b.IsEmbeddedType() {
		return nil
	}

	lastEditedData := &Block{
		LastEditedTime: Now(),
		LastEditedBy:   userID,
	}
	ops := b.embeddedFileOps(fileID, fileURL)
	ops = append(ops, b.UpdateOp(lastEditedData))
	if b.Parent != nil {
		ops = append(ops, b.Parent.UpdateOp(lastEditedData))
	}
	return ops
}
