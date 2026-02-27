package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Save configuration to ~/.bugsnag-cli.yaml",
	Long:  "Write API token and other settings to the config file so you don't have to pass them every time.",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("api-token")
		if token == "" {
			return fmt.Errorf("--api-token is required")
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}

		cfgPath := filepath.Join(home, ".bugsnag-cli.yaml")

		format, _ := cmd.Flags().GetString("default-format")
		baseURL, _ := cmd.Flags().GetString("default-base-url")
		perPage, _ := cmd.Flags().GetInt("default-per-page")

		var content string
		content += fmt.Sprintf("api_token: %s\n", token)
		if format != "" {
			content += fmt.Sprintf("format: %s\n", format)
		}
		if baseURL != "" {
			content += fmt.Sprintf("base_url: %s\n", baseURL)
		}
		if perPage > 0 {
			content += fmt.Sprintf("per_page: %d\n", perPage)
		}

		if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
			return fmt.Errorf("writing config file: %w", err)
		}

		p := output.NewPrinter(getFormat())
		result := map[string]string{
			"status": "ok",
			"path":   cfgPath,
		}
		return p.PrintSingle(result)
	},
}

func init() {
	configureCmd.Flags().StringP("api-token", "t", "", "Bugsnag API token (required)")
	configureCmd.Flags().String("default-format", "", "Default output format (json or table)")
	configureCmd.Flags().String("default-base-url", "", "Default API base URL")
	configureCmd.Flags().Int("default-per-page", 0, "Default results per page")

	rootCmd.AddCommand(configureCmd)
}
