package query

import (
	`encoding/json`
	"fmt"
	"strings"

	"gorm.io/gorm"

	"airtable-backend/pkg/models"
)

// Sort represents a single sort definition.
type Sort struct {
	FieldID   string `json:"fieldId"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// BuildGormSort builds GORM ORDER BY clauses from a slice of Sort structs.
// It requires the map of fields to get KeyName and Type for proper casting.
func BuildGormSort(db *gorm.DB, fields map[string]models.Field, sorts []Sort) (*gorm.DB, error) {
	if len(sorts) == 0 {
		return db, nil // No sorting applied
	}

	var orderClauses []string

	for _, sort := range sorts {
		field, ok := fields[sort.FieldID]
		if !ok {
			return nil, fmt.Errorf("field with ID %s not found for sorting", sort.FieldID)
		}

		// Validate direction
		direction := strings.ToUpper(sort.Direction)
		if direction != "ASC" && direction != "DESC" {
			return nil, fmt.Errorf("invalid sort direction for field %s: %s", field.Name, sort.Direction)
		}

		// Build the order clause with type casting for the JSONB field
		keyAccessor := fmt.Sprintf("data ->> '%s'", field.KeyName)
		var typedAccessor string

		// Determine casting based on field type
		switch field.Type {
		case models.FieldTypeText:
			typedAccessor = keyAccessor // No specific cast needed for text comparison/ordering
		case models.FieldNumber:
			typedAccessor = fmt.Sprintf("(%s)::numeric", keyAccessor)
		case models.TypeBoolean:
			typedAccessor = fmt.Sprintf("(%s)::boolean", keyAccessor)
		case models.TypeDate:
			typedAccessor = fmt.Sprintf("(%s)::timestamp", keyAccessor)
		default:
			// For unsupported types, sort as text or skip? Sorting as text is safer.
			typedAccessor = keyAccessor
			// Or error: return nil, fmt.Errorf("unsupported field type for sorting: %s (field: %s)", field.Type, field.Name)
		}

		// Add NULLs LAST or FIRST? Default is NULLS LAST for ASC, NULLS FIRST for DESC.
		// Let's explicitly use NULLS LAST for ASC and NULLS FIRST for DESC for consistency with many UIs.
		nullOrder := "NULLS LAST"
		if direction == "DESC" {
			nullOrder = "NULLS FIRST"
		}

		orderClause := fmt.Sprintf("%s %s %s", typedAccessor, direction, nullOrder)
		orderClauses = append(orderClauses, orderClause)
	}

	// Combine all order clauses
	finalOrderClause := strings.Join(orderClauses, ", ")

	return db.Order(finalOrderClause), nil
}

// ParseSortJSON parses a JSON byte slice into a slice of Sort.
func ParseSortJSON(sortJSON []byte) ([]Sort, error) {
	if len(sortJSON) == 0 {
		return nil, nil
	}
	var sorts []Sort
	if err := json.Unmarshal(sortJSON, &sorts); err != nil {
		return nil, fmt.Errorf("invalid sort JSON format: %v", err)
	}
	return sorts, nil
}
