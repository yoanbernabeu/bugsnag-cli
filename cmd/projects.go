package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage Bugsnag projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects for an organization",
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

		projects, hasMore, err := c.ListProjects(orgID, getAllPages())
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(projects), len(projects), hasMore)
		}
		return p.PrintList(projects, len(projects), hasMore)
	},
}

var projectsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a project by ID",
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

		project, err := c.GetProject(projectID)
		if err != nil {
			return err
		}

		return p.PrintSingle(project)
	},
}

func init() {
	projectsListCmd.Flags().String("org-id", "", "Organization ID (required)")
	projectsGetCmd.Flags().String("project-id", "", "Project ID (required)")

	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
	rootCmd.AddCommand(projectsCmd)
}
