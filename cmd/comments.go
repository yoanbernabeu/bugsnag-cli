package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage Bugsnag error comments",
}

var commentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List comments for an error",
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

		comments, hasMore, err := c.ListComments(projectID, errorID, getAllPages())
		if err != nil {
			return err
		}

		if getFormat() == "table" {
			return p.PrintList(output.ToTableRenderers(comments), len(comments), hasMore)
		}
		return p.PrintList(comments, len(comments), hasMore)
	},
}

var commentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a comment on an error",
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

		message, _ := cmd.Flags().GetString("message")
		if message == "" {
			return fmt.Errorf("--message is required")
		}

		c := client.New(getBaseURL(), token, getPerPage())
		p := output.NewPrinter(getFormat())

		comment, err := c.CreateComment(projectID, errorID, message)
		if err != nil {
			return err
		}

		return p.PrintSingle(comment)
	},
}

func init() {
	commentsListCmd.Flags().String("project-id", "", "Project ID (required)")
	commentsListCmd.Flags().String("error-id", "", "Error ID (required)")

	commentsCreateCmd.Flags().String("project-id", "", "Project ID (required)")
	commentsCreateCmd.Flags().String("error-id", "", "Error ID (required)")
	commentsCreateCmd.Flags().String("message", "", "Comment message (required)")

	commentsCmd.AddCommand(commentsListCmd)
	commentsCmd.AddCommand(commentsCreateCmd)
	rootCmd.AddCommand(commentsCmd)
}
