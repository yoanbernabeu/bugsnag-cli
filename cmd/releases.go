package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var releasesCmd = &cobra.Command{
	Use:   "releases",
	Short: "Manage Bugsnag releases",
}

var releasesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List releases for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project-id")
		if projectID == "" {
			return fmt.Errorf("--project-id is required")
		}

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		releases, hasMore, err := c.ListReleases(projectID, getAllPages())
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(releases), len(releases), hasMore)
		}
		return p.PrintList(releases, len(releases), hasMore)
	},
}

func init() {
	releasesListCmd.Flags().String("project-id", "", "Project ID (required)")

	releasesCmd.AddCommand(releasesListCmd)
	rootCmd.AddCommand(releasesCmd)
}
