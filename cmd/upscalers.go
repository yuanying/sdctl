package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var upscalersCmd = &cobra.Command{
	Use:   "upscalers",
	Short: "List available upscalers",
	RunE:  runUpscalersList,
}

func init() {
	rootCmd.AddCommand(upscalersCmd)
}

func runUpscalersList(cmd *cobra.Command, args []string) error {
	upscalers, err := client.ListUpscalers()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, u := range upscalers {
		fmt.Println(u.Name)
	}
	return nil
}
