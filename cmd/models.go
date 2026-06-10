package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage Stable Diffusion models",
}

var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models",
	RunE:  runModelsList,
}

var modelsSetCmd = &cobra.Command{
	Use:   "set [model-name]",
	Short: "Set active model",
	Args:  cobra.ExactArgs(1),
	RunE:  runModelsSet,
}

func init() {
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsSetCmd)
	rootCmd.AddCommand(modelsCmd)
}

func runModelsList(cmd *cobra.Command, args []string) error {
	models, err := client.ListModels()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}
	for _, m := range models {
		fmt.Printf("%-50s %s\n", m.ModelName, m.Hash)
	}
	return nil
}

func runModelsSet(cmd *cobra.Command, args []string) error {
	if err := client.SetModel(args[0]); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	fmt.Printf("Model set to: %s\n", args[0])
	return nil
}
