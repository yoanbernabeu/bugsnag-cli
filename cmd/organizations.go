package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var organizationsCmd = &cobra.Command{
	Use:   "organizations",
	Short: "Manage Bugsnag organizations",
}

var organizationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations for the authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		orgs, hasMore, err := c.ListOrganizations(getAllPages())
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(orgs), len(orgs), hasMore)
		}
		return p.PrintList(orgs, len(orgs), hasMore)
	},
}

func init() {
	organizationsCmd.AddCommand(organizationsListCmd)
	rootCmd.AddCommand(organizationsCmd)
}
