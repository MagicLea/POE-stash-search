package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Common Constans
const (
	// POE API Constant
	StashAPIURL = "public-stash-tabs"
	APIParamID  = "id"

	// DB Columns Constans
	DBOfPOEStashSearch = "poe-stash-search"
	COfStashChange     = "StashChange"
	KeyOfChangeID      = "change_id"
	KeyOfNextChangeID  = "next_change_id"
	COfItem            = "Item"
	StringOfStash      = "stashes"
	StringOfID         = "id"
	StringOfItem       = "item"
)

// Crawler is a crawler
type Crawler struct {
	// TODO:
	// mysql db
	// mongo db
	// TODO:
	apiBaseURL string
	mgo        *mgo.Session
}

// New returns a crawler
func New(baseURL string, mgo *mgo.Session) *Crawler {
	return &Crawler{
		apiBaseURL: baseURL,
		mgo:        mgo,
	}
}

// FollowStashStream follows public stash change stream, and stores changes into db
func (c *Crawler) FollowStashStream(lastID string) error {
	nextID := lastID
	conn := c.mgo.Copy()
	defer conn.Close()
	// FIXME: it is a demo, and just do 10 times
	for i := 0; i < 10; i++ {
		url := fmt.Sprintf("%s/%s/?%s=%s", c.apiBaseURL, StashAPIURL, APIParamID, nextID)
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		resMap := map[string]interface{}{}
		if err := json.Unmarshal(body, &resMap); err != nil {
			return err
		}

		conn.DB(DBOfPOEStashSearch).C(COfStashChange).Insert(
			map[string]interface{}{
				KeyOfChangeID:     nextID,
				KeyOfNextChangeID: resMap[KeyOfNextChangeID],
			})
		bulk := conn.DB(DBOfPOEStashSearch).C(COfItem).Bulk()
		bulk.Unordered()
		if resMap[StringOfStash] != nil {
			for _, stash := range resMap[StringOfStash].([]interface{}) {
				stashMap := stash.(map[string]interface{})
				if stashMap[StringOfItem] != nil {
					for _, item := range stashMap[StringOfItem].([]interface{}) {
						itemMap := item.(map[string]interface{})
						selector := bson.M{StringOfID: itemMap[StringOfID]}
						update := itemMap
						update["stash_id"] = stashMap[StringOfID]
						update["stash_type"] = stashMap["stashType"]
						bulk.Upsert(selector, update)
					}
				}
			}
		}
		bulk.Run()

		fmt.Println("next id:", resMap[KeyOfNextChangeID])
		nextID = resMap[KeyOfNextChangeID].(string)
	}
	return nil
}
