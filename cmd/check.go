package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/putzeys/putzeys-cli/internal/hardware"
	"github.com/putzeys/putzeys-cli/internal/models"
	"github.com/putzeys/putzeys-cli/internal/ollama"
	"github.com/putzeys/putzeys-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	checkSearch string
	checkQuant  string
	checkJSON   bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check model compatibility with your hardware",
	Long: `Detects your hardware and checks which Ollama models can run on your machine.
Shows a color-coded table: green (CAN RUN), yellow (DEGRADED), red (CAN'T RUN).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Detect hardware
		hw, err := hardware.DetectAll()
		if err != nil {
			return fmt.Errorf("detecting hardware: %w", err)
		}

		if !checkJSON {
			fmt.Print(ui.RenderHardwareInfo(hw))
			fmt.Println()
		}

		// 2. Get models to check
		quant := models.ParseQuantization(checkQuant)
		variants := getModelsToCheck(checkSearch, quant)

		if len(variants) == 0 {
			fmt.Println(ui.WarningStyle.Render("No models found to check."))
			return nil
		}

		// 3. Run compatibility check
		results := models.CheckAll(variants, hw, cfg.VRAMOverheadPercent, cfg.GPUMemoryFraction)

		// 4. Output
		if checkJSON {
			jsonOut := ui.RenderCompatJSON(results)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(jsonOut)
		}

		fmt.Print(ui.RenderCompatTable(results))
		fmt.Println()
		fmt.Println(ui.Legend())
		fmt.Println()

		return nil
	},
}

func init() {
	checkCmd.Flags().StringVarP(&checkSearch, "search", "s", "", "Filter models by name")
	checkCmd.Flags().StringVarP(&checkQuant, "quant", "q", "Q4_K_M", "Quantization level (Q4_K_M, Q8_0, FP16)")
	checkCmd.Flags().BoolVar(&checkJSON, "json", false, "Output results as JSON")
	rootCmd.AddCommand(checkCmd)
}

func getModelsToCheck(search string, quant models.Quantization) []models.ModelVariant {
	// Try to get models from registry
	registry := ollama.NewRegistry()
	remoteModels, err := registry.SearchModels(search, 0)
	if err != nil || len(remoteModels) == 0 {
		remoteModels = ollama.GetPopularModels()
	}

	// Filter by search if needed
	if search != "" {
		var filtered []ollama.RemoteModel
		for _, m := range remoteModels {
			if strings.Contains(strings.ToLower(m.Name), strings.ToLower(search)) {
				filtered = append(filtered, m)
			}
		}
		if len(filtered) > 0 {
			remoteModels = filtered
		}
	}

	// Convert to ModelVariants
	variants := make([]models.ModelVariant, 0, len(remoteModels))
	for _, m := range remoteModels {
		params := ollama.ParseModelParams(m.Name)
		est := models.EstimateVRAM(params, quant, cfg.VRAMOverheadPercent)

		variants = append(variants, models.ModelVariant{
			Name:         m.Name,
			Description:  m.Description,
			Parameters:   params,
			Quantization: quant,
			VRAMRequired: est.TotalVRAM,
		})
	}

	return variants
}
