package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var collaboratorsCmd = &cobra.Command{
	Use:   "collaborators",
	Short: "Manage Bugsnag collaborators",
}

var collaboratorsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List collaborators for an organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		orgID, _ := cmd.Flags().GetString("org-id")
		if orgID == "" {
			return fmt.Errorf("--org-id is required")
		}

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		collaborators, hasMore, err := c.ListCollaborators(orgID, getAllPages())
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(collaborators), len(collaborators), hasMore)
		}
		return p.PrintList(collaborators, len(collaborators), hasMore)
	},
}

func init() {
	collaboratorsListCmd.Flags().String("org-id", "", "Organization ID (required)")

	collaboratorsCmd.AddCommand(collaboratorsListCmd)
	rootCmd.AddCommand(collaboratorsCmd)
}
