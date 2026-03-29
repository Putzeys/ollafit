package cmd

import (
	"fmt"
	"os"

	"github.com/putzeys/ollafit/internal/ollama"
	"github.com/putzeys/ollafit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	modelsSearch string
	modelsLocal  bool
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available Ollama models",
	Long:  "Lists models from the ollamadb.dev registry or locally installed models.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if modelsLocal {
			return listLocalModels()
		}
		return listRemoteModels(modelsSearch)
	},
}

func init() {
	modelsCmd.Flags().StringVarP(&modelsSearch, "search", "s", "", "Search query to filter models")
	modelsCmd.Flags().BoolVarP(&modelsLocal, "local", "l", false, "List only locally installed models")
	rootCmd.AddCommand(modelsCmd)
}

func listLocalModels() error {
	client := ollama.NewClient(cfg.OllamaHost)

	if !client.IsRunning() {
		fmt.Fprintln(os.Stderr, ui.WarningStyle.Render("Ollama is not running at "+cfg.OllamaHost))
		fmt.Fprintln(os.Stderr, ui.DimStyle.Render("Start Ollama with: ollama serve"))
		return fmt.Errorf("ollama not running")
	}

	localModels, err := client.ListLocalModels()
	if err != nil {
		return fmt.Errorf("listing local models: %w", err)
	}

	if len(localModels) == 0 {
		fmt.Println(ui.WarningStyle.Render("No models installed locally."))
		fmt.Println(ui.DimStyle.Render("Pull a model with: ollafit pull <model>"))
		return nil
	}

	names := make([]string, len(localModels))
	descs := make([]string, len(localModels))
	for i, m := range localModels {
		names[i] = m.Name
		descs[i] = fmt.Sprintf("%s | %s | %s",
			m.Details.ParameterSize,
			m.Details.QuantizationLevel,
			formatModelSize(m.Size))
	}

	fmt.Print(ui.RenderModelList(names, descs))
	return nil
}

func listRemoteModels(search string) error {
	registry := ollama.NewRegistry()
	remoteModels, err := registry.SearchModels(search, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.WarningStyle.Render("Registry unavailable, showing popular models"))
		remoteModels = ollama.GetPopularModels()
	}

	if len(remoteModels) == 0 {
		// Fallback to popular models
		remoteModels = ollama.GetPopularModels()
	}

	names := make([]string, len(remoteModels))
	descs := make([]string, len(remoteModels))
	for i, m := range remoteModels {
		names[i] = m.Name
		descs[i] = m.Description
	}

	fmt.Print(ui.RenderModelList(names, descs))
	return nil
}

func formatModelSize(bytes int64) string {
	const (
		MB = 1024 * 1024
		GB = 1024 * MB
	)
	if bytes >= GB {
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	}
	return fmt.Sprintf("%d MB", bytes/MB)
}
