package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yoanbernabeu/bugsnag-cli/internal/client"
	"github.com/yoanbernabeu/bugsnag-cli/internal/output"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "bugsnag",
	Short: "CLI for the Bugsnag Data Access API",
	Long:  "A command-line tool for interacting with the Bugsnag Data Access API. Designed for code agents (JSON output by default) and humans (--format table).",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		p := output.NewPrinter(getFormat())
		exitCode := classifyError(err)
		p.PrintError(err.Error(), exitCode)
	}
}

func classifyError(err error) int {
	if err == nil {
		return output.ExitOK
	}
	msg := err.Error()
	if strings.Contains(msg, "API token is required") || strings.Contains(msg, "is required") {
		return output.ExitConfig
	}
	if apiErr, ok := err.(*client.APIError); ok {
		_ = apiErr
		return output.ExitAPI
	}
	if strings.Contains(msg, "network error") {
		return output.ExitNetwork
	}
	if strings.Contains(msg, "API error") {
		return output.ExitAPI
	}
	return output.ExitGeneral
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.bugsnag-cli.yaml)")
	rootCmd.PersistentFlags().StringP("api-token", "t", "", "Bugsnag API token")
	rootCmd.PersistentFlags().StringP("format", "f", "json", "Output format: json or table")
	rootCmd.PersistentFlags().Int("per-page", 30, "Number of results per page")
	rootCmd.PersistentFlags().BoolP("all-pages", "a", false, "Fetch all pages of results")
	rootCmd.PersistentFlags().String("base-url", "https://api.bugsnag.com", "Bugsnag API base URL")

	_ = viper.BindPFlag("api_token", rootCmd.PersistentFlags().Lookup("api-token"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("per_page", rootCmd.PersistentFlags().Lookup("per-page"))
	_ = viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigName(".bugsnag-cli")
			viper.SetConfigType("yaml")
		}
	}

	viper.SetEnvPrefix("BUGSNAG")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
}

func getAPIToken() (string, error) {
	token := viper.GetString("api_token")
	if token == "" {
		return "", fmt.Errorf("API token is required. Set via --api-token, BUGSNAG_API_TOKEN env var, or config file")
	}
	return token, nil
}

func getFormat() string {
	f := viper.GetString("format")
	if f == "" {
		return "json"
	}
	return f
}

func getPerPage() int {
	pp := viper.GetInt("per_page")
	if pp <= 0 || pp > 100 {
		return 30
	}
	return pp
}

func getAllPages() bool {
	ap, _ := rootCmd.PersistentFlags().GetBool("all-pages")
	return ap
}

func getBaseURL() string {
	u := viper.GetString("base_url")
	if u == "" {
		return "https://api.bugsnag.com"
	}
	return u
}
