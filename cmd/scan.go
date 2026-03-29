package cmd

import (
	"fmt"
	"os"

	"github.com/putzeys/ollafit/internal/hardware"
	"github.com/putzeys/ollafit/internal/ui"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Detect and display hardware specs",
	Long:  "Scans your system for GPU(s), VRAM, RAM, and CPU information.",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := hardware.DetectAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting hardware: %v\n", err)
			return err
		}

		fmt.Print(ui.RenderHardwareInfo(info))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
