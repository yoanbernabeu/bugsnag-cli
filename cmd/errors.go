package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var errorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Manage Bugsnag errors",
}

var errorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List errors for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project-id")
		if projectID == "" {
			return fmt.Errorf("--project-id is required")
		}

		status, _ := cmd.Flags().GetString("status")
		severity, _ := cmd.Flags().GetString("severity")
		sort, _ := cmd.Flags().GetString("sort")
		direction, _ := cmd.Flags().GetString("direction")

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		opts := client.ListErrorsOptions{
			ProjectID: projectID,
			Status:    status,
			Severity:  severity,
			Sort:      sort,
			Direction: direction,
			AllPages:  getAllPages(),
		}

		errors, hasMore, err := c.ListErrors(opts)
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(errors), len(errors), hasMore)
		}
		return p.PrintList(errors, len(errors), hasMore)
	},
}

var errorsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an error by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project-id")
		if projectID == "" {
			return fmt.Errorf("--project-id is required")
		}

		errorID, _ := cmd.Flags().GetString("error-id")
		if errorID == "" {
			return fmt.Errorf("--error-id is required")
		}

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		bugsnagErr, err := c.GetError(projectID, errorID)
		if err != nil {
			return err
		}

		return p.PrintSingle(bugsnagErr)
	},
}

func init() {
	errorsListCmd.Flags().String("project-id", "", "Project ID (required)")
	errorsListCmd.Flags().String("status", "", "Filter by status (open, fixed, snoozed, ignored)")
	errorsListCmd.Flags().String("severity", "", "Filter by severity (info, warning, error)")
	errorsListCmd.Flags().String("sort", "", "Sort field (created_at, last_seen, events, users, unsorted)")
	errorsListCmd.Flags().String("direction", "", "Sort direction (asc, desc)")

	errorsGetCmd.Flags().String("project-id", "", "Project ID (required)")
	errorsGetCmd.Flags().String("error-id", "", "Error ID (required)")

	errorsCmd.AddCommand(errorsListCmd)
	errorsCmd.AddCommand(errorsGetCmd)
	rootCmd.AddCommand(errorsCmd)
}
