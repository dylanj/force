package force

import (
	"encoding/json"
	"fmt"
	"strings"
)

type QueryRecord map[string]any

type QueryResponse struct {
	TotalSize int  `json:"totalSize"`
	Done      bool `json:"done"`
	Records   []QueryRecord
}

func (c *Client) QueryBasic(q string) ([]QueryRecord, error) {
	query := strings.ReplaceAll(q, " ", "+")
	path := "/services/data/v" + c.version + "/query/?q=" + query

	m := QueryResponse{}
	b, err := c.Get(path)
	if err != nil {
		return []QueryRecord{}, err
	}

	err = json.Unmarshal(b, &m)

	if err != nil {
		return []QueryRecord{}, err
	}

	return m.Records, nil
}

func (c *Client) Explain(q string) ([]QueryPlan, error) {
	query := strings.ReplaceAll(q, " ", "+")
	path := "/services/data/v" + c.version + "/query/?explain=" + query

	m := QueryPlanResponse{}
	b, err := c.Get(path)
	if err != nil {
		return m.Plans, err
	}
	fmt.Println(string(b))
	err = json.Unmarshal(b, &m)

	if err != nil {
		return m.Plans, err
	}

	return m.Plans, nil
}
