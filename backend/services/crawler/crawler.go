package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// POE API Constant
	stashAPIURL = "public-stash-tabs"
	apiParamID  = "id"

	// DB Fields Constants
	dbPOEStashSearch = "poe-stash-search"
	cStashChange     = "StashChange"
	keyChangeID      = "change_id"
	keyNextChangeID  = "next_change_id"
	cItem            = "Item"
	keyStash         = "stashes"
	keyID            = "id"
	keyItem          = "item"
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
		url := fmt.Sprintf("%s/%s/?%s=%s", c.apiBaseURL, stashAPIURL, apiParamID, nextID)
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

		conn.DB(dbPOEStashSearch).C(cStashChange).Insert(
			map[string]interface{}{
				keyChangeID:     nextID,
				keyNextChangeID: resMap[keyNextChangeID],
			})
		bulk := conn.DB(dbPOEStashSearch).C(cItem).Bulk()
		bulk.Unordered()
		if resMap[keyStash] != nil {
			for _, stash := range resMap[keyStash].([]interface{}) {
				stashMap := stash.(map[string]interface{})
				if stashMap[keyItem] != nil {
					for _, item := range stashMap[keyItem].([]interface{}) {
						itemMap := item.(map[string]interface{})
						selector := bson.M{keyID: itemMap[keyID]}
						update := itemMap
						update["stash_id"] = stashMap[keyID]
						update["stash_type"] = stashMap["stashType"]
						bulk.Upsert(selector, update)
					}
				}
			}
		}
		bulk.Run()

		fmt.Println("next id:", resMap[keyNextChangeID])
		nextID = resMap[keyNextChangeID].(string)
	}
	return nil
}
