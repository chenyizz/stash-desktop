package savedfilter

import (
	"context"

	"case/backend/pkg/models"
	"case/backend/pkg/models/jsonschema"
)

// ToJSON converts a SavedFilter object into its JSON equivalent.
func ToJSON(ctx context.Context, filter *models.SavedFilter) (*jsonschema.SavedFilter, error) {
	return &jsonschema.SavedFilter{
		Name:         filter.Name,
		Mode:         filter.Mode,
		FindFilter:   filter.FindFilter,
		ObjectFilter: filter.ObjectFilter,
		UIOptions:    filter.UIOptions,
	}, nil
}
