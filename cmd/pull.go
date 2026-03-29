package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/putzeys/ollafit/internal/hardware"
	"github.com/putzeys/ollafit/internal/models"
	"github.com/putzeys/ollafit/internal/ollama"
	"github.com/putzeys/ollafit/internal/ui"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <model>",
	Short: "Download a model via Ollama",
	Long:  "Downloads a model from the Ollama registry with a progress bar. Warns if hardware may not support it.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelName := args[0]
		client := ollama.NewClient(cfg.OllamaHost)

		// Check if Ollama is running
		if !client.IsRunning() {
			fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("Ollama is not running at "+cfg.OllamaHost))
			fmt.Fprintln(os.Stderr, ui.DimStyle.Render("Start Ollama with: ollama serve"))
			return fmt.Errorf("ollama not running")
		}

		// Check hardware compatibility first
		hw, err := hardware.DetectAll()
		if err == nil {
			params := ollama.ParseModelParams(modelName)
			quant := models.QuantQ4KM
			result := models.CheckCompatibility(
				models.ModelVariant{
					Name:         modelName,
					Parameters:   params,
					Quantization: quant,
				},
				hw,
				cfg.VRAMOverheadPercent,
				cfg.GPUMemoryFraction,
			)

			switch result.Verdict {
			case models.VerdictCannotRun:
				fmt.Println(ui.VerdictCannotRun.Render("WARNING: Your hardware may not support this model"))
				fmt.Println(ui.DimStyle.Render(result.Reason))
				fmt.Println()
			case models.VerdictDegraded:
				fmt.Println(ui.VerdictDegraded.Render("NOTE: This model will run in degraded mode"))
				fmt.Println(ui.DimStyle.Render(result.Reason))
				fmt.Println()
			case models.VerdictCanRun:
				fmt.Println(ui.VerdictCanRun.Render("Hardware check: OK"))
				fmt.Println()
			}
		}

		fmt.Printf("Pulling %s...\n", ui.LabelStyle.Render(modelName))

		// Set up bubbletea progress UI
		prog := ui.NewPullProgress()
		p := tea.NewProgram(prog)

		// Run pull in background, send messages to bubbletea
		go func() {
			err := client.PullModel(modelName, func(status string, total, completed int64) {
				p.Send(ui.PullProgressMsg{
					Status:    status,
					Total:     total,
					Completed: completed,
				})
			})
			p.Send(ui.PullDoneMsg{Err: err})
		}()

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("progress UI error: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
