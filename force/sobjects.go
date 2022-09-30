package force

import "encoding/json"

type SObjectDescription struct {
	Activateable        bool   `json:"activateable"`
	Custom              bool   `json:"custom"`
	CustomSetting       bool   `json:"customSetting"`
	Createable          bool   `json:"createable"`
	Deletable           bool   `json:"deletable"`
	DeprecatedAndHidden bool   `json:"deprecatedAndHidden"`
	FeedEnabled         bool   `json:"feedEnabled"`
	KeyPrefix           string `json:"keyPrefix"`
	Label               string `json:"label"`
	LabelPlural         string `json:"labelPlural"`
	Layoutable          bool   `json:"layoutable"`
	Mergeable           bool   `json:"mergeable"`
	MruEnabled          bool   `json:"mruEnabled"`
	Name                string `json:"name"`
	Queryable           bool   `json:"queryable"`
	Replicateable       bool   `json:"replicateable"`
	Retrieveable        bool   `json:"retrieveable"`
	Searchable          bool   `json:"searchable"`
	Triggerable         bool   `json:"triggerable"`
	Undeletable         bool   `json:"undeletable"`
	Updateable          bool   `json:"updateable"`
	Urls                struct {
		Sobject     string `json:"sobject"`
		Describe    string `json:"describe"`
		RowTemplate string `json:"rowTemplate"`
	}
}

type SObjectsResponse struct {
	Encoding     string               `json:"encoding"`
	MaxBatchSize int                  `json:"maxBatchSize"`
	SObjects     []SObjectDescription `json:"sobjects"`
}

func (c Client) DescribeGlobal() ([]SObjectDescription, error) {
	path := "/services/data/" + c.version + "/sobjects/"

	b, err := c.Get(path)
	if err != nil {
		return []SObjectDescription{}, err
	}

	r := SObjectsResponse{}
	err = json.Unmarshal(b, &r)

	if err != nil {
		return []SObjectDescription{}, err
	}

	return r.SObjects, nil
}
