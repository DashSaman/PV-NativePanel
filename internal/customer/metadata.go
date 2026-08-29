package customer

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var ErrInvalidCustomerMetadata = errors.New("customer: invalid customer metadata")

type CustomerMetadataInput struct {
	Note         *string  `json:"note,omitempty"`
	GroupID      *string  `json:"group_id,omitempty"`
	OnHold       *bool    `json:"on_hold,omitempty"`
	AddTagIDs    []string `json:"add_tag_ids,omitempty"`
	RemoveTagIDs []string `json:"remove_tag_ids,omitempty"`
}

type CustomerMetadata struct {
	Note            string         `json:"note"`
	Group           *CustomerGroup `json:"group,omitempty"`
	Tags            []CustomerTag  `json:"tags"`
	OnHold          bool           `json:"on_hold"`
	AssignedActorID string         `json:"assigned_actor_id,omitempty"`
}

type customerMetadataStore interface {
	UpdateCustomerMetadataTx(context.Context, *sql.Tx, string, string, CustomerMetadataInput) (CustomerMetadata, error)
}

func (s *Service) UpdateCustomerMetadata(
	ctx context.Context,
	tx *sql.Tx,
	actorID, userID string,
	input CustomerMetadataInput,
) (CustomerMetadata, error) {
	if s == nil || s.store == nil || tx == nil || strings.TrimSpace(actorID) == "" || strings.TrimSpace(userID) == "" {
		return CustomerMetadata{}, ErrInvalidCustomerMetadata
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		if len(note) > 4000 {
			return CustomerMetadata{}, ErrInvalidCustomerMetadata
		}
		input.Note = &note
	}
	if input.GroupID != nil {
		groupID := strings.TrimSpace(*input.GroupID)
		input.GroupID = &groupID
	}
	input.AddTagIDs = uniqueStrings(input.AddTagIDs)
	input.RemoveTagIDs = uniqueStrings(input.RemoveTagIDs)
	store, ok := s.store.(customerMetadataStore)
	if !ok {
		return CustomerMetadata{}, errors.New("customer: metadata capability is unavailable")
	}
	return store.UpdateCustomerMetadataTx(ctx, tx, actorID, userID, input)
}
