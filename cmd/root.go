package cmd

import (
	"fmt"
	"os"

	"github.com/putzeys/putzeys-cli/internal/config"
	"github.com/putzeys/putzeys-cli/internal/ui"
	"github.com/spf13/cobra"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "putzeys-cli",
	Short: "Check Ollama model compatibility with your hardware",
	Long: `putzeys-cli detects your hardware (GPU, VRAM, RAM, CPU) and shows which
Ollama models can run on your machine. You can also download models directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.RunTUI(cfg)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: error loading config: %v\n", err)
		cfg = config.DefaultConfig()
	}
}
