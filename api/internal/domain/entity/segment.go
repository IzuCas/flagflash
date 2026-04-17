package entity

import (
	"time"

	"github.com/google/uuid"
)

// Segment represents a reusable user segment/cohort
type Segment struct {
	ID            uuid.UUID   `json:"id"`
	TenantID      uuid.UUID   `json:"tenant_id"`
	Name          string      `json:"name"`
	Description   string      `json:"description,omitempty"`
	Conditions    []Condition `json:"conditions"`
	IsDynamic     bool        `json:"is_dynamic"`
	IncludedUsers []string    `json:"included_users,omitempty"`
	ExcludedUsers []string    `json:"excluded_users,omitempty"`
	CreatedBy     *uuid.UUID  `json:"created_by,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// NewSegment creates a new segment
func NewSegment(tenantID uuid.UUID, name, description string, conditions []Condition, createdBy *uuid.UUID) *Segment {
	now := time.Now()
	return &Segment{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Name:          name,
		Description:   description,
		Conditions:    conditions,
		IsDynamic:     true,
		IncludedUsers: []string{},
		ExcludedUsers: []string{},
		CreatedBy:     createdBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// MatchesContext checks if the given context matches the segment conditions
func (s *Segment) MatchesContext(ctx *EvaluationContext) bool {
	// Check excluded users first
	for _, userID := range s.ExcludedUsers {
		if ctx.UserID == userID {
			return false
		}
	}

	// Check included users (static segment)
	if !s.IsDynamic {
		for _, userID := range s.IncludedUsers {
			if ctx.UserID == userID {
				return true
			}
		}
		return false
	}

	// Dynamic segment - evaluate conditions
	if len(s.Conditions) == 0 {
		return false
	}

	// All conditions must match (AND logic)
	for _, condition := range s.Conditions {
		if !condition.Matches(ctx) {
			return false
		}
	}
	return true
}

// Update updates segment details
func (s *Segment) Update(name, description string, conditions []Condition) {
	if name != "" {
		s.Name = name
	}
	if description != "" {
		s.Description = description
	}
	if conditions != nil {
		s.Conditions = conditions
	}
	s.UpdatedAt = time.Now()
}

// AddIncludedUser adds a user to the included list
func (s *Segment) AddIncludedUser(userID string) {
	for _, u := range s.IncludedUsers {
		if u == userID {
			return
		}
	}
	s.IncludedUsers = append(s.IncludedUsers, userID)
	s.UpdatedAt = time.Now()
}

// RemoveIncludedUser removes a user from the included list
func (s *Segment) RemoveIncludedUser(userID string) {
	for i, u := range s.IncludedUsers {
		if u == userID {
			s.IncludedUsers = append(s.IncludedUsers[:i], s.IncludedUsers[i+1:]...)
			s.UpdatedAt = time.Now()
			return
		}
	}
}

// AddExcludedUser adds a user to the excluded list
func (s *Segment) AddExcludedUser(userID string) {
	for _, u := range s.ExcludedUsers {
		if u == userID {
			return
		}
	}
	s.ExcludedUsers = append(s.ExcludedUsers, userID)
	s.UpdatedAt = time.Now()
}
