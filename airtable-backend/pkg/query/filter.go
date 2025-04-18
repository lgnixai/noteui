package query

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"airtable-backend/pkg/models"
)

// FilterGroup represents a group of conditions combined by an operator (AND/OR).
type FilterGroup struct {
	Operator   string            `json:"operator"`   // "AND" or "OR"
	Conditions []json.RawMessage `json:"conditions"` // Can be Condition or nested FilterGroup
}

// Condition represents a single filter condition.
type Condition struct {
	FieldID  string          `json:"fieldId"`
	Operator string          `json:"operator"` // e.g., "=", "!=", ">", "<", ">=", "<=", "contains", "is_empty"
	Value    json.RawMessage `json:"value"`    // Raw value, interpretation depends on FieldType
}

// BuildGormFilter recursively builds GORM WHERE clauses from a FilterGroup.
// It requires the map of fields to get KeyName and Type.
func BuildGormFilter(db *gorm.DB, fields map[string]models.Field, filter *FilterGroup) (*gorm.DB, error) {
	if filter == nil || (filter.Operator == "" && len(filter.Conditions) == 0) {
		return db, nil // No filter applied
	}

	// Validate operator
	if filter.Operator != "AND" && filter.Operator != "OR" {
		return nil, fmt.Errorf("invalid filter operator: %s", filter.Operator)
	}

	// GORM needs separate OR clauses
	// We will build individual clause strings and combine them
	var clauses []string
	var args []interface{}

	for _, rawCondition := range filter.Conditions {
		// Try unmarshalling as a simple condition first
		var cond Condition
		if err := json.Unmarshal(rawCondition, &cond); err == nil {
			// Successfully unmarshalled as a condition
			field, ok := fields[cond.FieldID]
			if !ok {
				return nil, fmt.Errorf("field with ID %s not found", cond.FieldID)
			}

			clause, conditionArgs, err := buildConditionClause(field, cond)
			if err != nil {
				return nil, fmt.Errorf("failed to build condition clause for field %s: %v", field.Name, err)
			}
			clauses = append(clauses, clause)
			args = append(args, conditionArgs...)

		} else {
			// If not a condition, try unmarshalling as a nested group
			var nestedGroup FilterGroup
			if err := json.Unmarshal(rawCondition, &nestedGroup); err == nil {
				// Successfully unmarshalled as a nested group
				// Recursively build the nested clause
				nestedDB, err := BuildGormFilter(db.Session(&gorm.Session{}), fields, &nestedGroup)
				if err != nil {
					return nil, fmt.Errorf("failed to build nested filter group: %v", err)
				}
				// GORM returns a DB object with the WHERE clause applied. We need the SQL string and args.
				// This requires GORM internals or re-implementing part of it.
				// A simpler approach for this example: return the DB object directly and combine.
				// However, combining nested WHERE clauses with different operators (AND/OR) is tricky
				// when GORM applies them sequentially.
				// A more robust approach: build pure SQL clause strings and args here.

				// --- Re-thinking combining clauses ---
				// GORM's `Where` accepts strings and args, OR `func(*gorm.DB) *gorm.DB`.
				// For complex nested structures, passing functions is cleaner.
				// Let's return a function that applies the filter.

				// This requires changing the function signature.
				// New plan: BuildGormFilter returns a function `func(*gorm.DB) *gorm.DB`
				// Let's refactor this...

				// --- Refactored approach using GORM function ---
				// This function will return the modified *gorm.DB object directly.
				// The complexity of combining AND/OR needs careful handling.
				// GORM's `Or()` applies an OR condition *after* any preceding WHERE.
				// For `(A AND B) OR (C AND D)`, you might need `Where(A).Where(B).Or(Where(C).Where(D))`.
				// For `(A OR B) AND (C OR D)`, you need `Where(Where(A).Or(B)).Where(Where(C).Or(D))`.

				// Let's simplify the JSON structure mapping to GORM:
				// Top-level operator applies to all immediate conditions/groups.
				// `{"operator": "AND", "conditions": [Cond1, Cond2, Group1]}` -> `WHERE Cond1 AND Cond2 AND (Group1_Clause)`
				// `{"operator": "OR", "conditions": [Cond1, Cond2, Group1]}` -> `WHERE Cond1 OR Cond2 OR (Group1_Clause)`
				// This means we need to build the clause for each condition/group and then combine them with the group's operator.

				nestedDB, err = BuildGormFilter(db.Session(&gorm.Session{}), fields, &nestedGroup) // Use a session to avoid modifying the original db instance immediately
				if err != nil {
					return nil, fmt.Errorf("failed to build nested filter group: %v", err)
				}

				// GORM doesn't easily expose the generated SQL WHERE clause string and args from a DB object *before* execution.
				// We need to build the raw SQL clause parts.

				// Let's reconsider the return type. Returning a function seems most GORM-idiomatic for composition.
				// `func(*gorm.DB) *gorm.DB` represents a scope.
				// BuildGormFilter will return `func(*gorm.DB) *gorm.DB`.

				// This is getting complex quickly. For a practical example while keeping it reasonable,
				// let's implement parsing into *our* internal filter struct first, then build the GORM query.
				// And handle the AND/OR structure iteratively, applying GORM's `.Where()` and `.Or()` methods.
				// GORM's `.Where(clause, args...).Where(clause2, args2...)` creates `clause AND clause2`.
				// GORM's `.Where(clause, args...).Or(clause2, args2...)` creates `(clause) OR (clause2)`.
				// This is not quite what we need for nested `AND` within `OR`, etc.

				// A better approach for arbitrary nesting: build the entire SQL string and args for the WHERE clause.
				// This bypasses some of GORM's query builder but gives full control.

				// --- Re-attempt: Build raw SQL string and args ---
				fmt.Println(nestedDB)
				// This recursive function will build the SQL clause and collect args.
				clause, conditionArgs, err := buildGroupClause(fields, &nestedGroup)
				if err != nil {
					return nil, err
				}
				clauses = append(clauses, clause)
				args = append(args, conditionArgs...)
			} else {
				// Failed to unmarshal as either Condition or FilterGroup
				return nil, fmt.Errorf("invalid filter condition format: %s", string(rawCondition))
			}
		}
	}

	// Combine the individual clauses using the group's operator
	combinedClause := strings.Join(clauses, fmt.Sprintf(" %s ", filter.Operator))
	if len(clauses) > 1 {
		combinedClause = "(" + combinedClause + ")" // Wrap in parentheses for nesting correctness
	}

	// Apply the final combined clause to the DB object
	return db.Where(combinedClause, args...), nil
}

// buildGroupClause recursively builds the SQL string and arguments for a filter group.
func buildGroupClause(fields map[string]models.Field, group *FilterGroup) (string, []interface{}, error) {
	if group == nil || (group.Operator == "" && len(group.Conditions) == 0) {
		return "", nil, nil
	}

	if group.Operator != "AND" && group.Operator != "OR" {
		return "", nil, fmt.Errorf("invalid filter operator: %s", group.Operator)
	}

	var clauses []string
	var args []interface{}

	for _, rawCondition := range group.Conditions {
		var cond Condition
		if err := json.Unmarshal(rawCondition, &cond); err == nil {
			// It's a simple condition
			field, ok := fields[cond.FieldID]
			if !ok {
				return "", nil, fmt.Errorf("field with ID %s not found", cond.FieldID)
			}
			clause, conditionArgs, err := buildConditionClause(field, cond)
			if err != nil {
				return "", nil, fmt.Errorf("failed to build condition clause for field %s: %v", field.Name, err)
			}
			clauses = append(clauses, clause)
			args = append(args, conditionArgs...)
		} else {
			// Try as a nested group
			var nestedGroup FilterGroup
			if err := json.Unmarshal(rawCondition, &nestedGroup); err == nil {
				// It's a nested group
				nestedClause, nestedArgs, err := buildGroupClause(fields, &nestedGroup)
				if err != nil {
					return "", nil, err
				}
				if nestedClause != "" {
					clauses = append(clauses, nestedClause)
					args = append(args, nestedArgs...)
				}
			} else {
				// Neither a condition nor a group
				return "", nil, fmt.Errorf("invalid filter condition format: %s", string(rawCondition))
			}
		}
	}

	if len(clauses) == 0 {
		return "", nil, nil // No valid conditions in group
	}

	// Combine with the group's operator
	combinedClause := strings.Join(clauses, fmt.Sprintf(" %s ", group.Operator))
	if len(clauses) > 1 {
		combinedClause = "(" + combinedClause + ")" // Wrap if more than one clause
	}

	return combinedClause, args, nil
}

// buildConditionClause builds the SQL string and arguments for a single condition.
// Assumes value is passed as JSON raw message.
func buildConditionClause(field models.Field, cond Condition) (string, []interface{}, error) {
	// Base SQL template for accessing JSONB field value as text
	// We use ->> 'key' to get the value as TEXT. Casting is needed for non-text comparisons.
	keyAccessor := fmt.Sprintf("data ->> '%s'", field.KeyName)

	var clause string
	var args []interface{}

	// Determine comparison operator and value handling based on field type
	switch field.Type {
	case models.FieldTypeText:
		var value string
		// Unmarshal value assuming it's a string
		if err := json.Unmarshal(cond.Value, &value); err != nil {
			return "", nil, fmt.Errorf("invalid value format for text field %s: %v", field.Name, err)
		}

		switch cond.Operator {
		case "=":
			clause = fmt.Sprintf("%s = ?", keyAccessor)
			args = append(args, value)
		case "!=":
			clause = fmt.Sprintf("%s != ?", keyAccessor)
			args = append(args, value)
		case "contains":
			clause = fmt.Sprintf("%s LIKE ?", keyAccessor)
			args = append(args, "%"+value+"%")
		case "not_contains":
			clause = fmt.Sprintf("%s NOT LIKE ?", keyAccessor)
			args = append(args, "%"+value+"%")
		case "starts_with":
			clause = fmt.Sprintf("%s LIKE ?", keyAccessor)
			args = append(args, value+"%")
		case "ends_with":
			clause = fmt.Sprintf("%s LIKE ?", keyAccessor)
			args = append(args, "%"+value)
		case "is_empty":
			// Consider both NULL (missing key) and empty string
			clause = fmt.Sprintf("(%s IS NULL OR %s = '')", keyAccessor, keyAccessor)
		case "is_not_empty":
			clause = fmt.Sprintf("(%s IS NOT NULL AND %s != '')", keyAccessor, keyAccessor)
		default:
			return "", nil, fmt.Errorf("unsupported operator for text field %s: %s", field.Name, cond.Operator)
		}

	case models.FieldNumber:
		var value float64
		// Unmarshal value assuming it's a number
		if err := json.Unmarshal(cond.Value, &value); err != nil {
			return "", nil, fmt.Errorf("invalid value format for number field %s: %v", field.Name, err)
		}
		// Cast JSONB value to numeric for comparison
		numericAccessor := fmt.Sprintf("(%s)::numeric", keyAccessor)

		switch cond.Operator {
		case "=", "!=", ">", "<", ">=", "<=":
			clause = fmt.Sprintf("%s %s ?", numericAccessor, cond.Operator)
			args = append(args, value)
		case "is_empty":
			clause = fmt.Sprintf("%s IS NULL", keyAccessor) // Number fields are typically NULL when empty
		case "is_not_empty":
			clause = fmt.Sprintf("%s IS NOT NULL", keyAccessor)
		default:
			return "", nil, fmt.Errorf("unsupported operator for number field %s: %s", field.Name, cond.Operator)
		}

	case models.TypeBoolean:
		var value bool
		// Unmarshal value assuming it's a boolean
		if err := json.Unmarshal(cond.Value, &value); err != nil {
			return "", nil, fmt.Errorf("invalid value format for boolean field %s: %v", field.Name, err)
		}
		// Cast JSONB value to boolean for comparison
		booleanAccessor := fmt.Sprintf("(%s)::boolean", keyAccessor)

		switch cond.Operator {
		case "=", "!=":
			clause = fmt.Sprintf("%s %s ?", booleanAccessor, cond.Operator)
			args = append(args, value)
		case "is_empty":
			clause = fmt.Sprintf("%s IS NULL", keyAccessor) // Boolean fields are typically NULL when empty
		case "is_not_empty":
			clause = fmt.Sprintf("%s IS NOT NULL", keyAccessor)
		default:
			return "", nil, fmt.Errorf("unsupported operator for boolean field %s: %s", field.Name, cond.Operator)
		}

	case models.TypeDate:
		var value string // Expect date string in a format PostgreSQL can parse, e.g., 'YYYY-MM-DD' or RFC3339
		// Unmarshal value assuming it's a string
		if err := json.Unmarshal(cond.Value, &value); err != nil {
			return "", nil, fmt.Errorf("invalid value format for date field %s: %v", field.Name, err)
		}
		// Cast JSONB value to timestamp or date for comparison
		dateAccessor := fmt.Sprintf("(%s)::timestamp", keyAccessor) // Use timestamp for flexibility

		switch cond.Operator {
		case "=", "!=", ">", "<", ">=", "<=":
			clause = fmt.Sprintf("%s %s ?", dateAccessor, cond.Operator)
			args = append(args, value) // Pass date string, PG will parse
		case "is_empty":
			clause = fmt.Sprintf("%s IS NULL", keyAccessor) // Date fields are typically NULL when empty
		case "is_not_empty":
			clause = fmt.Sprintf("%s IS NOT NULL", keyAccessor)
		default:
			return "", nil, fmt.Errorf("unsupported operator for date field %s: %s", field.Name, cond.Operator)
		}

	default:
		// Handle unsupported or generic types
		return "", nil, fmt.Errorf("unsupported field type for filtering: %s (field: %s)", field.Type, field.Name)
	}

	return clause, args, nil
}

// ParseFilterJSON parses a JSON byte slice into a FilterGroup.
func ParseFilterJSON(filterJSON []byte) (*FilterGroup, error) {
	if len(filterJSON) == 0 {
		return nil, nil
	}
	var filter FilterGroup
	if err := json.Unmarshal(filterJSON, &filter); err != nil {
		return nil, fmt.Errorf("invalid filter JSON format: %v", err)
	}
	return &filter, nil
}
