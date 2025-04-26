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

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourusername/tusk-go/internal/service/domain"
)

// NewCommandNameCommand creates a new command for handling domain objects.
func NewCommandNameCommand(service domain.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command-name",
		Short: "Brief description of command",
		Long:  `Detailed description of command and its usage.`,
	}

	// Add subcommands
	cmd.AddCommand(newCreateCommand(service))
	cmd.AddCommand(newListCommand(service))
	cmd.AddCommand(newGetCommand(service))
	cmd.AddCommand(newUpdateCommand(service))
	cmd.AddCommand(newDeleteCommand(service))

	return cmd
}

func newCreateCommand(service domain.Service) *cobra.Command {
	var (
		field1 string
		field2 string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new domain object",
		Long:  `Create a new domain object with the specified fields.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Create request from flags
			req := domain.CreateRequest{
				Field1: field1,
				Field2: field2,
			}

			// Call service
			result, err := service.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("creating domain object: %w", err)
			}

			// Output result
			fmt.Fprintf(cmd.OutOrStdout(), "Created domain object with ID: %s\n", result.ID)
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVar(&field1, "field1", "", "Description of field1")
	cmd.Flags().StringVar(&field2, "field2", "", "Description of field2")

	// Mark required flags
	_ = cmd.MarkFlagRequired("field1")

	return cmd
}

func newListCommand(service domain.Service) *cobra.Command {
	var (
		filterField string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List domain objects",
		Long:  `List all domain objects, optionally filtered by criteria.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Create filter from flags
			filter := domain.Filter{
				Field: filterField,
			}

			// Call service
			results, err := service.List(ctx, filter)
			if err != nil {
				return fmt.Errorf("listing domain objects: %w", err)
			}

			// Output results
			if len(results) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No domain objects found")
				return nil
			}

			for _, item := range results {
				fmt.Fprintf(cmd.OutOrStdout(), "ID: %s, Field1: %s\n", item.ID, item.Field1)
			}

			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVar(&filterField, "filter", "", "Filter results by this field")

	return cmd
}

func newGetCommand(service domain.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get a specific domain object",
		Long:  `Get a domain object by its unique identifier.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			id := args[0]

			// Call service
			result, err := service.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("getting domain object: %w", err)
			}

			// Output result
			fmt.Fprintf(cmd.OutOrStdout(), "ID: %s\nField1: %s\nField2: %s\n",
				result.ID, result.Field1, result.Field2)
			return nil
		},
	}
}

func newUpdateCommand(service domain.Service) *cobra.Command {
	var (
		field1 string
		field2 string
	)

	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Update a domain object",
		Long:  `Update a domain object with the specified fields.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			id := args[0]

			// Create request from flags
			req := domain.UpdateRequest{
				Field1: field1,
				Field2: field2,
			}

			// Call service
			result, err := service.Update(ctx, id, req)
			if err != nil {
				return fmt.Errorf("updating domain object: %w", err)
			}

			// Output result
			fmt.Fprintf(cmd.OutOrStdout(), "Updated domain object with ID: %s\n", result.ID)
			return nil
		},
	}

	// Define flags
	cmd.Flags().StringVar(&field1, "field1", "", "New value for field1")
	cmd.Flags().StringVar(&field2, "field2", "", "New value for field2")

	return cmd
}

func newDeleteCommand(service domain.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a domain object",
		Long:  `Delete a domain object by its unique identifier.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			id := args[0]

			// Call service
			if err := service.Delete(ctx, id); err != nil {
				return fmt.Errorf("deleting domain object: %w", err)
			}

			// Output result
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted domain object with ID: %s\n", id)
			return nil
		},
	}
}
