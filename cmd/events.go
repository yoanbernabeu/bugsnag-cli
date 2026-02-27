package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage Bugsnag events",
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events for a project or error",
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

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		events, hasMore, err := c.ListEvents(projectID, errorID, getAllPages())
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(events), len(events), hasMore)
		}
		return p.PrintList(events, len(events), hasMore)
	},
}

var eventsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an event by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project-id")
		if projectID == "" {
			return fmt.Errorf("--project-id is required")
		}

		eventID, _ := cmd.Flags().GetString("event-id")
		if eventID == "" {
			return fmt.Errorf("--event-id is required")
		}

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		event, err := c.GetEvent(projectID, eventID)
		if err != nil {
			return err
		}

		return p.PrintSingle(event)
	},
}

func init() {
	eventsListCmd.Flags().String("project-id", "", "Project ID (required)")
	eventsListCmd.Flags().String("error-id", "", "Error ID (optional, scope events to an error)")

	eventsGetCmd.Flags().String("project-id", "", "Project ID (required)")
	eventsGetCmd.Flags().String("event-id", "", "Event ID (required)")

	eventsCmd.AddCommand(eventsListCmd)
	eventsCmd.AddCommand(eventsGetCmd)
	rootCmd.AddCommand(eventsCmd)
}
