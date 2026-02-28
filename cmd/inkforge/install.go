package main

import (
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

// installCmd represents the install command for initializing environment
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Initialize InkForge environment including Playwright browsers",
	Long: `The install command sets up the required environment for InkForge,
including downloading and installing Playwright browser dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		installEnvironment()
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installEnvironment() {
	logs.Infof("Installing Playwright browsers and dependencies...")

	err := playwright.Install()
	if err != nil {
		logs.Errorf("Error installing Playwright: %v", err)
		panic(err)
	}

	logs.Infof("Successfully installed Playwright browsers and dependencies")
}
