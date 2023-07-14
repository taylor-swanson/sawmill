package logs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/taylor-swanson/sawmill/internal/collections"
)

type FilterOp uint8

const (
	FilterOpEquals FilterOp = iota
	FilterOpNotEquals
	FilterOpGreaterThan
	FilterOpLessThan
	FilterOpIncludes
	FilterOpExcludes
	FilterOpBetween
	FilterOpNotBetween
)

func (o *FilterOp) String() string {
	switch *o {
	case FilterOpEquals:
		return "Equals"
	case FilterOpNotEquals:
		return "Not Equals"
	case FilterOpGreaterThan:
		return "Greater Than"
	case FilterOpLessThan:
		return "Less Than"
	case FilterOpIncludes:
		return "Includes"
	case FilterOpExcludes:
		return "Excludes"
	case FilterOpBetween:
		return "Between"
	case FilterOpNotBetween:
		return "Not Between"
	}

	return ""
}

func (o *FilterOp) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("unable to unmarshal FilterOp: %w", err)
	}
	switch strings.ToUpper(s) {
	case "EQUALS":
		*o = FilterOpEquals
	case "NOT_EQUALS":
		*o = FilterOpEquals
	case "GREATER_THAN":
		*o = FilterOpGreaterThan
	case "LESS_THAN":
		*o = FilterOpLessThan
	case "INCLUDES":
		*o = FilterOpIncludes
	case "EXCLUDES":
		*o = FilterOpExcludes
	case "BETWEEN":
		*o = FilterOpBetween
	case "NOT_BETWEEN":
		*o = FilterOpNotBetween
	default:
		return fmt.Errorf("unable to unmarshal FilterOp, unknown operator: %q", s)
	}

	return nil
}

func (o *FilterOp) MarshalJSON() ([]byte, error) {
	var out string

	switch *o {
	case FilterOpEquals:
		out = "EQUALS"
	case FilterOpNotEquals:
		out = "NOT_EQUALS"
	case FilterOpGreaterThan:
		out = "GREATER_THAN"
	case FilterOpLessThan:
		out = "LESS_THAN"
	case FilterOpIncludes:
		out = "INCLUDES"
	case FilterOpExcludes:
		out = "EXCLUDES"
	case FilterOpBetween:
		out = "BETWEEN"
	case FilterOpNotBetween:
		out = "NOT_BETWEEN"
	default:
		return nil, fmt.Errorf("unable to marshal FilterOp, unknown operator: %d", *o)
	}

	return []byte(out), nil
}

type Filter interface {
	Filter(line collections.Fields) bool
	ValidOps() FilterOp
}

type TextFilter struct {
	Operator FilterOp `json:"operator"`
	Field    string   `json:"field"`
	Value    string   `json:"value"`
}

func (f *TextFilter) Filter(line collections.Fields) bool {
	rawValue, ok := line.Get(f.Field)
	if !ok {
		return false
	}
	value, ok := rawValue.(string)
	if !ok {
		return false
	}

	switch f.Operator {
	case FilterOpEquals:
		return strings.EqualFold(value, f.Value)
	case FilterOpNotEquals:
		return !strings.EqualFold(value, f.Value)
	case FilterOpIncludes:
		return strings.Contains(strings.ToLower(value), strings.ToLower(f.Value))
	case FilterOpExcludes:
		return !strings.Contains(strings.ToLower(value), strings.ToLower(f.Value))
	}

	return false
}

func (f *TextFilter) ValidOps() []FilterOp {
	return []FilterOp{FilterOpEquals, FilterOpNotEquals, FilterOpIncludes, FilterOpExcludes}
}

type NumberFilter struct {
	Operator FilterOp `json:"operator"`
	Field    string   `json:"field"`
	Value    float64  `json:"value"`
	Value2   float64  `json:"value2,omitempty"`
}

func (f *NumberFilter) Filter(line collections.Fields) bool {
	value, ok := line.GetNumber(f.Field)
	if !ok {
		return false
	}

	switch f.Operator {
	case FilterOpEquals:
		return value == f.Value
	case FilterOpNotEquals:
		return value != f.Value
	case FilterOpGreaterThan:
		return f.Value > value
	case FilterOpLessThan:
		return f.Value < value
	case FilterOpBetween:
		return f.Value <= value && value <= f.Value2
	case FilterOpNotBetween:
		return f.Value < value || value > f.Value2
	}

	return false
}

func (f *NumberFilter) ValidOps() []FilterOp {
	return []FilterOp{FilterOpEquals, FilterOpNotEquals, FilterOpGreaterThan, FilterOpLessThan, FilterOpBetween, FilterOpNotBetween}
}

type TimeFilter struct {
	Operator FilterOp  `json:"operator"`
	Field    string    `json:"field"`
	Value    time.Time `json:"value"`
	Value2   time.Time `json:"value2,omitempty"`
	Format   string    `json:"format"`
}

func (f *TimeFilter) Filter(line collections.Fields) bool {
	if f.Format == "" {
		f.Format = time.RFC3339Nano
	}
	value, ok := line.GetTime(f.Field, f.Format)
	if !ok {
		return false
	}

	switch f.Operator {
	case FilterOpEquals:
		return value == f.Value
	case FilterOpNotEquals:
		return value != f.Value
	case FilterOpGreaterThan:
		return f.Value.After(value)
	case FilterOpLessThan:
		return f.Value.Before(value)
	case FilterOpBetween:
		if value == f.Value || value == f.Value2 {
			return true
		}
		return f.Value.After(value) && f.Value2.Before(value)
	case FilterOpNotBetween:
		return f.Value.Before(value) || f.Value2.After(value)
	}

	return false
}

func (f *TimeFilter) ValidOps() []FilterOp {
	return []FilterOp{FilterOpEquals, FilterOpNotEquals, FilterOpGreaterThan, FilterOpLessThan, FilterOpBetween, FilterOpNotBetween}
}

type BoolFilter struct {
	Operator FilterOp `json:"operator"`
	Field    string   `json:"field"`
	Value    bool     `json:"value"`
}

func (f *BoolFilter) Filter(line collections.Fields) bool {
	value, ok := line.GetBool(f.Field)
	if !ok {
		return false
	}

	switch f.Operator {
	case FilterOpEquals:
		return value == f.Value
	case FilterOpNotEquals:
		return value != f.Value
	}

	return false
}

func (f *BoolFilter) ValidOps() []FilterOp {
	return []FilterOp{FilterOpEquals, FilterOpNotEquals}
}
