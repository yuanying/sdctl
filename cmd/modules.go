package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modulesCmd = &cobra.Command{
	Use:   "modules",
	Short: "List available VAE and text encoder modules",
	RunE:  runModulesList,
}

func init() {
	rootCmd.AddCommand(modulesCmd)
}

func runModulesList(cmd *cobra.Command, args []string) error {
	modules, err := client.ListSDModules()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, m := range modules {
		fmt.Printf("%-50s %s\n", m.ModelName, m.Filename)
	}
	return nil
}
