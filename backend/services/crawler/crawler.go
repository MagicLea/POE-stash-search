package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
		url := fmt.Sprintf("%s/public-stash-tabs/?id=%s", c.apiBaseURL, nextID)
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

		conn.DB("poe-stash-search").C("StashChange").Insert(
			map[string]interface{}{
				"change_id":      nextID,
				"next_change_id": resMap["next_change_id"],
			})
		bulk := conn.DB("poe-stash-search").C("Item").Bulk()
		bulk.Unordered()
		if resMap["stashes"] != nil {
			for _, stash := range resMap["stashes"].([]interface{}) {
				stashMap := stash.(map[string]interface{})
				if stashMap["items"] != nil {
					for _, item := range stashMap["items"].([]interface{}) {
						itemMap := item.(map[string]interface{})
						selector := bson.M{"id": itemMap["id"]}
						update := itemMap
						update["stash_id"] = stashMap["id"]
						update["stash_type"] = stashMap["stashType"]
						bulk.Upsert(selector, update)
					}
				}
			}
		}
		bulk.Run()

		fmt.Println("next id:", resMap["next_change_id"])
		nextID = resMap["next_change_id"].(string)
	}
	return nil
}
