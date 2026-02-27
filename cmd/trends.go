package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var trendsCmd = &cobra.Command{
	Use:   "trends",
	Short: "View error trends",
}

var trendsProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Get error trends for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project-id")
		if projectID == "" {
			return fmt.Errorf("--project-id is required")
		}

		resolution, _ := cmd.Flags().GetString("resolution")
		bucketsCount, _ := cmd.Flags().GetInt("buckets-count")

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		buckets, err := c.GetProjectTrends(projectID, resolution, bucketsCount)
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(buckets), len(buckets), false)
		}
		return p.PrintList(buckets, len(buckets), false)
	},
}

var trendsErrorCmd = &cobra.Command{
	Use:   "error",
	Short: "Get error trends for a specific error",
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

		buckets, err := c.GetErrorTrends(projectID, errorID)
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(buckets), len(buckets), false)
		}
		return p.PrintList(buckets, len(buckets), false)
	},
}

func init() {
	trendsProjectCmd.Flags().String("project-id", "", "Project ID (required)")
	trendsProjectCmd.Flags().String("resolution", "", "Time resolution (1h, 1d, etc.)")
	trendsProjectCmd.Flags().Int("buckets-count", 0, "Number of trend buckets")

	trendsErrorCmd.Flags().String("project-id", "", "Project ID (required)")
	trendsErrorCmd.Flags().String("error-id", "", "Error ID (required)")

	trendsCmd.AddCommand(trendsProjectCmd)
	trendsCmd.AddCommand(trendsErrorCmd)
	rootCmd.AddCommand(trendsCmd)
}
