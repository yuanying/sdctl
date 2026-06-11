package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var samplersCmd = &cobra.Command{
	Use:   "samplers",
	Short: "Manage samplers",
}

var samplersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available samplers",
	RunE:  runSamplersList,
}

func init() {
	samplersCmd.AddCommand(samplersListCmd)
	rootCmd.AddCommand(samplersCmd)
}

func runSamplersList(cmd *cobra.Command, args []string) error {
	samplers, err := client.ListSamplers()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, s := range samplers {
		fmt.Println(s.Name)
	}
	return nil
}
