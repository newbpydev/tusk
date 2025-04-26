// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package domain

import (
	"context"
	"fmt"

	"github.com/yourusername/tusk-go/internal/core/errors"
)

// ServiceName defines the operations for managing domain objects.
type ServiceName interface {
	Create(ctx context.Context, req CreateRequest) (*DomainObject, error)
	Get(ctx context.Context, id string) (*DomainObject, error)
	List(ctx context.Context, filter Filter) ([]*DomainObject, error)
	Update(ctx context.Context, id string, req UpdateRequest) (*DomainObject, error)
	Delete(ctx context.Context, id string) error
}

// serviceImpl implements the ServiceName interface.
type serviceImpl struct {
	repository Repository
	// Add other dependencies here
}

// NewService creates a new instance of ServiceName.
func NewService(repository Repository) ServiceName {
	return &serviceImpl{
		repository: repository,
	}
}

// Create implements the ServiceName.Create method.
func (s *serviceImpl) Create(ctx context.Context, req CreateRequest) (*DomainObject, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validating request: %w", err)
	}

	// Convert request to domain model
	obj := &DomainObject{
		// Map fields from request to domain model
	}

	// Save to repository
	createdObj, err := s.repository.Create(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("creating domain object: %w", err)
	}

	return createdObj, nil
}

// Get implements the ServiceName.Get method.
func (s *serviceImpl) Get(ctx context.Context, id string) (*DomainObject, error) {
	if id == "" {
		return nil, errors.NewValidationError("id cannot be empty")
	}

	obj, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting domain object: %w", err)
	}

	if obj == nil {
		return nil, errors.NewNotFoundError("domain object not found")
	}

	return obj, nil
}

// List implements the ServiceName.List method.
func (s *serviceImpl) List(ctx context.Context, filter Filter) ([]*DomainObject, error) {
	objects, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing domain objects: %w", err)
	}

	return objects, nil
}

// Update implements the ServiceName.Update method.
func (s *serviceImpl) Update(ctx context.Context, id string, req UpdateRequest) (*DomainObject, error) {
	if id == "" {
		return nil, errors.NewValidationError("id cannot be empty")
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validating update request: %w", err)
	}

	// Get existing object
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting domain object for update: %w", err)
	}

	if existing == nil {
		return nil, errors.NewNotFoundError("domain object not found")
	}

	// Apply updates
	// existing.Field = req.Field

	// Save updated object
	updated, err := s.repository.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("updating domain object: %w", err)
	}

	return updated, nil
}

// Delete implements the ServiceName.Delete method.
func (s *serviceImpl) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.NewValidationError("id cannot be empty")
	}

	// Check if object exists
	existing, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting domain object for deletion: %w", err)
	}

	if existing == nil {
		return errors.NewNotFoundError("domain object not found")
	}

	// Delete the object
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting domain object: %w", err)
	}

	return nil
}
