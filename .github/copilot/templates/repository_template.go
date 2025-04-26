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

package output

import (
	"context"

	"github.com/yourusername/tusk-go/internal/core/domain"
)

// Repository defines the interface for persisting and retrieving domain objects.
type Repository interface {
	// Create persists a new domain object and returns the created entity.
	Create(ctx context.Context, entity *domain.Entity) (*domain.Entity, error)

	// GetByID retrieves a domain object by its unique identifier.
	GetByID(ctx context.Context, id string) (*domain.Entity, error)

	// List retrieves all domain objects matching the provided filter.
	List(ctx context.Context, filter domain.Filter) ([]*domain.Entity, error)

	// Update modifies an existing domain object.
	Update(ctx context.Context, entity *domain.Entity) (*domain.Entity, error)

	// Delete removes a domain object by its unique identifier.
	Delete(ctx context.Context, id string) error
}
