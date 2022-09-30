package force

type QueryPlanResponse struct {
	Plans []QueryPlan `json:"plans"`
}

type QueryPlan struct {
	Cardinality          int             `json:"cardinality"`
	Fields               []string        `json:"fields"`
	LeadingOperationType string          `json:"leadingOperationType"`
	Notes                []QueryPlanNote `json:"notes"`
	RelativeCost         float64         `json:"relativeCost"`
	SObjectCardinality   int             `json:"sobjectCardinality"`
	SObjectType          string          `json:"sobjectType"`
}

type QueryPlanNote struct {
	Description   string   `json:"description"`
	Fields        []string `json:"fields"`
	TableEnumOrId string   `json:"table_enum_or_id"`
}
