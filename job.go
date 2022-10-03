package force

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
)

type ErrorResponse struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

type QueryJobBody struct {
	// Operation can be either query (non deleted) or queryAll (includes deleted
	// records)
	Operation       string `json:"operation"`
	Query           string `json:"query"`
	ContentType     string `json:"contentType,omitempty"`
	ColumnDelimeter string `json:"columnDelimeter,omitempty"`
	LineEnding      string `json:"lineEnding,omitempty"`
}

type QueryJobResponse struct {
	Id              string  `json:"id"`
	Operation       string  `json:"operation"`
	Object          string  `json:"object"`
	CreatedById     string  `json:"createdById"`
	CreatedDate     string  `json:"createdDate"`
	SystemModstamp  string  `json:"systemModstamp"`
	State           string  `json:"state"`
	ConcurrencyMode string  `json:"concurrencyMode"`
	ContentType     string  `json:"content_type"`
	APIVersion      float64 `json:"apiVersion"`
	LineEnding      string  `json:"lineEnding"`
	ColumnDelimiter string  `json:"columnDelimiter"`
}

/*
"Id","Name"
"0011r00001zLVT1AAO","Mp6C3GwvHE"
"0011r00001zLvkCAAS","rAkO8O9krh"
*/
func parseResults(b []byte) ([]map[string]string, error) {
	r := csv.NewReader(bytes.NewReader(b))
	m := make([]map[string]string, 0)

	// parse headers from first line
	h, err := r.Read()
	if err != nil {
		return m, err
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return m, err
		}

		v := map[string]string{}
		for i := 0; i < len(h); i++ {
			v[h[i]] = record[i]
		}

		m = append(m, v)
	}

	return m, nil
}

func (c *Client) QueryJobResults(jobId string, callback func([]map[string]string)) error {
	q := url.Values{}
	q.Add("maxRecords", "1000")

	for {
		path := "/services/data/" + c.version + "/jobs/query/" + jobId + "/results?" + q.Encode()
		b, h, err := c.GetWithHeaders(path)

		if err != nil {
			return err
		}

		m, err := parseResults(b)
		if err != nil {
			return err
		}
		callback(m)

		locator := h.Get("Sforce-Locator")
		q.Set("locator", locator)

		if locator == "null" {
			break
		}
	}

	return nil
}

func (c *Client) QueryJobStatus(jobId string) (*QueryJobResponse, error) {
	b, err := c.Get("/services/data/" + c.version + "/jobs/query/" + jobId)
	if err != nil {
		return nil, err
	}

	resp := QueryJobResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		err = parseError(b)
		return nil, err
	}

	return &resp, nil
}

func (c *Client) QueryJob(query string) (*QueryJobResponse, error) {
	q := QueryJobBody{
		Operation: "query",
		Query:     query,
	}

	b, err := c.Post("/services/data/"+c.version+"/jobs/query", q)
	if err != nil {
		return nil, err
	}

	resp := QueryJobResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		err = parseError(b)
		fmt.Println(string(b))
		return nil, err
	}

	return &resp, nil
}

type QueryAllJobsResponse struct {
	Done           bool               `json:"done"`
	Records        []QueryJobResponse `json:"records"`
	NextRecordsURL string             `json:"nextRecordsUrl"`
}

func (c Client) GetAllJobs() (*QueryAllJobsResponse, error) {
	b, err := c.Get("/services/data/" + c.version + "/jobs/query")
	if err != nil {
		return nil, err
	}

	resp := QueryAllJobsResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		err = parseError(b)
		return nil, err
	}

	return &resp, nil
}

func parseError(b []byte) error {
	e := []ErrorResponse{}
	json.Unmarshal(b, &e)

	// todo: better error handling
	if e[0].ErrorCode == "INVALID_SESSION_ID" {
		return errors.New("invalid session" + string(b))
	}

	return errors.New("generic: " + string(b))
}

type queryJobState struct {
	State string `json:"state"`
}

func (c Client) AbortQueryJob(jobId string) (*QueryJobResponse, error) {
	b, err := c.Patch("/services/data/"+c.version+"/jobs/query/"+jobId, queryJobState{State: "Aborted"})
	if err != nil {
		return nil, err
	}

	resp := QueryJobResponse{}
	err = json.Unmarshal(b, &resp)
	if err != nil {
		err = parseError(b)
		return nil, err
	}

	return &resp, nil
}
