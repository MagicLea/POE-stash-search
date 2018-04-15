package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Crawler is a crawler
type Crawler struct {
	// TODO:
	// mysql db
	// mongo db
	// TODO:
	apiBaseURL string
}

// New returns a crawler
func New(baseURL string) *Crawler {
	return &Crawler{
		apiBaseURL: baseURL,
	}
}

// FollowStashStream follows public stash change stream, and stores changes into db
func (c *Crawler) FollowStashStream(lastID string) error {
	nextID := lastID
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

		fmt.Println("next id:", resMap["next_change_id"])
		nextID = resMap["next_change_id"].(string)
	}
	return nil
}
