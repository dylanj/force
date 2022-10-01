package force

import (
	"encoding/json"
)

type SObjectFilteredLookupInfo struct {
	ControllingFields []string `json:"controllingFields"`
	Dependent         bool     `json:"dependent"`
	OptionalFilter    bool     `json:"optionalFilter"`
}

type SObjectPicklistEntry struct {
	Active       bool   `json:"active"`
	ValidFor     []byte `json:"validFor"`
	DefaultValue bool   `json:"defaultValue"`
	Label        string `json:"label"`
	Value        string `json:"value"`
}

type SObjectField struct {
	Autonumber               bool   `json:"autonumber"`
	ByteLength               int    `json:"byteLength"`
	Calculated               bool   `json:"calculated"`
	CaseSensitive            bool   `json:"caseSensitive"`
	ControllerName           string `json:"controllerName"`
	Createable               bool   `json:"createable"`
	Custom                   bool   `json:"custom"`
	DataTranslationEnabled   bool   `json:"dataTranslationEnabled"`
	DefaultedOnCreate        bool   `json:"defaultedOnCreate"`
	DefaultValueFormula      string `json:"defaultValueFormula"`
	DependentPicklist        bool   `json:"dependentPicklist"`
	DeprecatedAndHidden      bool   `json:"deprecatedAndHidden"`
	Digits                   int    `json:"digits"`
	DisplayLocationInDecimal bool   `json:"displayLocationInDecimal"`
	Encrypted                bool   `json:"encrypted"`
	ExtraTypeInfo            string `json:"extraTypeInfo"`
	Filterable               bool   `json:"filterable"`

	FilteredLookupInfo SObjectFilteredLookupInfo `json:"filteredLookupInfo"`

	Formula         string `json:"formula"`
	Groupable       bool   `json:"groupable"`
	HighScaleNumber bool   `json:"highScaleNumber"`
	HtmlFormatted   bool   `json:"htmlFormatted"`
	IdLookup        bool   `json:"idLookup"`
	InlineHelpText  string `json:"inlineHelpText"`
	label           string `json:"label"`
	Length          int    `json:"length"`
	Mask            string `json:"mask"`
	MaskType        string `json:"maskType"`
	Name            string `json:"name"`
	NameField       bool   `json:"nameField"`
	namePointing    bool   `json:"namePointing"`
	Nillable        bool   `json:"nillable"`
	Permissionable  bool   `json:"permissionable"`

	PicklistValues []SObjectPicklistEntry `json:"picklistValues"`

	PolymorphicForeignKey bool     `json:"polymorphicForeignKey"`
	Precision             int      `json:"precision"`
	RelationshipName      string   `json:"relationshipName"`
	RelationshipOrder     int      `json:"relationshipOrder"`
	ReferenceTargetField  string   `json:"referenceTargetField"`
	ReferenceTo           []string `json:"referenceTo"`
	RestrictedPicklist    bool     `json:"restrictedPicklist"`
	Scale                 int      `json:"scale"`
	SearchPrefilterable   bool     `json:"searchPrefilterable"`
	SoapType              string   `json:"soapType"`
	Sortable              bool     `json:"sortable"`
	Type                  string   `json:"type"`
	Unique                bool     `json:"unique"`
	Updateable            bool     `json:"updateable"`
}

type SObjectDescription struct {
	Activateable        bool           `json:"activateable"`
	Custom              bool           `json:"custom"`
	CustomSetting       bool           `json:"customSetting"`
	Createable          bool           `json:"createable"`
	Deletable           bool           `json:"deletable"`
	DeprecatedAndHidden bool           `json:"deprecatedAndHidden"`
	FeedEnabled         bool           `json:"feedEnabled"`
	Fields              []SObjectField `json:"fields"`
	KeyPrefix           string         `json:"keyPrefix"`
	Label               string         `json:"label"`
	LabelPlural         string         `json:"labelPlural"`
	Layoutable          bool           `json:"layoutable"`
	Mergeable           bool           `json:"mergeable"`
	MruEnabled          bool           `json:"mruEnabled"`
	Name                string         `json:"name"`
	Queryable           bool           `json:"queryable"`
	Replicateable       bool           `json:"replicateable"`
	Retrieveable        bool           `json:"retrieveable"`
	Searchable          bool           `json:"searchable"`
	Triggerable         bool           `json:"triggerable"`
	Undeletable         bool           `json:"undeletable"`
	Updateable          bool           `json:"updateable"`
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

func (c Client) DescribeSObject(obj string) (SObjectDescription, error) {
	path := "/services/data/" + c.version + "/sobjects/" + obj + "/describe"

	b, err := c.Get(path)
	if err != nil {
		return SObjectDescription{}, err
	}

	r := SObjectDescription{}
	err = json.Unmarshal(b, &r)

	if err != nil {
		return SObjectDescription{}, err
	}

	return r, nil
}

//curl https://MyDomainName.my.salesforce.com/services/data/v56.0/sobjects/Account/describe/ -H "Authorization: Bearer token"
