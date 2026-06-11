package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var samplersCmd = &cobra.Command{
	Use:   "samplers",
	Short: "Manage samplers and schedulers",
}

var samplersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available samplers",
	RunE:  runSamplersList,
}

var schedulersListCmd = &cobra.Command{
	Use:   "schedulers",
	Short: "List available schedulers",
	RunE:  runSchedulersList,
}

func init() {
	samplersCmd.AddCommand(samplersListCmd)
	samplersCmd.AddCommand(schedulersListCmd)
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

func runSchedulersList(cmd *cobra.Command, args []string) error {
	schedulers, err := client.ListSchedulers()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, s := range schedulers {
		fmt.Printf("%-20s %s\n", s.Name, s.Label)
	}
	return nil
}
