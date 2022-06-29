package notionapi

import (
	"encoding/json"
	"fmt"
)

// Record represents a polymorphic record
type Record struct {
	// fields returned by the server
	Role string `json:"role"`
	// polymorphic value of the record, which we decode into Block, Space etc.
	Value json.RawMessage `json:"value"`

	// fields calculated from Value based on type
	ID             string          `json:"-"`
	Table          string          `json:"-"`
	Activity       *Activity       `json:"-"`
	Block          *Block          `json:"-"`
	Space          *Space          `json:"-"`
	NotionUser     *NotionUser     `json:"-"`
	UserRoot       *UserRoot       `json:"-"`
	UserSettings   *UserSettings   `json:"-"`
	Collection     *Collection     `json:"-"`
	CollectionView *CollectionView `json:"-"`
	Comment        *Comment        `json:"-"`
	Discussion     *Discussion     `json:"-"`
	// TODO: add more types
}

// table is not always present in Record returned by the server
// so must be provided based on what was asked
func parseRecord(table string, r *Record) error {
	// it's ok if some records don't return a value
	if len(r.Value) == 0 {
		return nil
	}
	if r.Table == "" {
		r.Table = table
	} else {
		// TODO: probably never happens
		panicIf(r.Table != table)
	}

	// set Block/Space etc. based on TableView type
	var pRawJSON *map[string]interface{}
	var obj interface{}
	switch table {
	case TableActivity:
		r.Activity = &Activity{}
		obj = r.Activity
		pRawJSON = &r.Activity.RawJSON
	case TableBlock:
		r.Block = &Block{}
		obj = r.Block
		pRawJSON = &r.Block.RawJSON
	case TableNotionUser:
		r.NotionUser = &NotionUser{}
		obj = r.NotionUser
		pRawJSON = &r.NotionUser.RawJSON
	case TableUserRoot:
		r.UserRoot = &UserRoot{}
		obj = r.UserRoot
		pRawJSON = &r.UserRoot.RawJSON
	case TableUserSettings:
		r.UserSettings = &UserSettings{}
		obj = r.UserSettings
		pRawJSON = &r.UserSettings.RawJSON
	case TableSpace:
		r.Space = &Space{}
		obj = r.Space
		pRawJSON = &r.Space.RawJSON
	case TableCollection:
		r.Collection = &Collection{}
		obj = r.Collection
		pRawJSON = &r.Collection.RawJSON
	case TableCollectionView:
		r.CollectionView = &CollectionView{}
		obj = r.CollectionView
		pRawJSON = &r.CollectionView.RawJSON
	case TableDiscussion:
		r.Discussion = &Discussion{}
		obj = r.Discussion
		pRawJSON = &r.Discussion.RawJSON
	case TableComment:
		r.Comment = &Comment{}
		obj = r.Comment
		pRawJSON = &r.Comment.RawJSON
	}
	if obj == nil {
		return fmt.Errorf("unsupported table '%s'", r.Table)
	}
	if err := jsonit.Unmarshal(r.Value, pRawJSON); err != nil {
		return err
	}
	id := (*pRawJSON)["id"]
	if id != nil {
		r.ID = id.(string)
	}
	if err := jsonit.Unmarshal(r.Value, &obj); err != nil {
		return err
	}
	return nil
}
