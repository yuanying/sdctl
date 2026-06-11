package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var schedulersCmd = &cobra.Command{
	Use:   "schedulers",
	Short: "Manage schedulers",
}

var schedulersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available schedulers",
	RunE:  runSchedulersList,
}

func init() {
	schedulersCmd.AddCommand(schedulersListCmd)
	rootCmd.AddCommand(schedulersCmd)
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
