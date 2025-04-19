package models

// FilterOperator 定义过滤操作符
type FilterOperator string

const (
	FilterEqual        FilterOperator = "eq"
	FilterNotEqual     FilterOperator = "neq"
	FilterGreaterThan  FilterOperator = "gt"
	FilterLessThan     FilterOperator = "lt"
	FilterGreaterEqual FilterOperator = "gte"
	FilterLessEqual    FilterOperator = "lte"
	FilterContains     FilterOperator = "contains"
	FilterIn           FilterOperator = "in"
	FilterNotIn        FilterOperator = "notIn"
	FilterIsNull       FilterOperator = "isNull"
	FilterIsNotNull    FilterOperator = "isNotNull"
)

// SortDirection 定义排序方向
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// FilterCondition 定义过滤条件
type FilterCondition struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
}

// SortCondition 定义排序条件
type SortCondition struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"direction"`
}

// QueryParams 定义查询参数
type QueryParams struct {
	Filters    []FilterCondition `json:"filters,omitempty"`
	Sort       []SortCondition   `json:"sort,omitempty"`
	Page       int               `json:"page,omitempty"`
	PageSize   int               `json:"pageSize,omitempty"`
	GroupBy    []string          `json:"groupBy,omitempty"`
	Aggregates []string          `json:"aggregates,omitempty"`
	Joins      []string          `json:"joins,omitempty"`
}

// QueryResult 定义查询结果
type QueryResult struct {
	Records    []Record               `json:"records"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
	Aggregates map[string]interface{} `json:"aggregates,omitempty"`
}

// AggregateFunction 定义聚合函数
type AggregateFunction string

const (
	AggregateCount AggregateFunction = "count"
	AggregateSum   AggregateFunction = "sum"
	AggregateAvg   AggregateFunction = "avg"
	AggregateMin   AggregateFunction = "min"
	AggregateMax   AggregateFunction = "max"
)

// AggregateResult 定义聚合结果
type AggregateResult struct {
	Function AggregateFunction `json:"function"`
	Field    string            `json:"field"`
	Value    interface{}       `json:"value"`
}
