package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var stabilityCmd = &cobra.Command{
	Use:   "stability",
	Short: "View stability metrics",
}

var stabilityTrendCmd = &cobra.Command{
	Use:   "trend",
	Short: "Get stability trend for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := getAPIToken()
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project-id")
		if projectID == "" {
			return fmt.Errorf("--project-id is required")
		}

		releaseStage, _ := cmd.Flags().GetString("release-stage")

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		trend, err := c.GetStabilityTrend(projectID, releaseStage)
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(trend.TimelinePoints), len(trend.TimelinePoints), false)
		}
		return p.PrintSingle(trend)
	},
}

func init() {
	stabilityTrendCmd.Flags().String("project-id", "", "Project ID (required)")
	stabilityTrendCmd.Flags().String("release-stage", "", "Release stage (optional)")

	stabilityCmd.AddCommand(stabilityTrendCmd)
	rootCmd.AddCommand(stabilityCmd)
}
