# putzeys-cli

CLI tool that detects your hardware (GPU, VRAM, RAM, CPU) and shows which [Ollama](https://ollama.com) models can run on your machine.

## Features

- **Interactive TUI** — full-screen interface with 4 tabs (Hardware, Compatibilidade, Modelos, Pull)
- **Hardware detection** — GPU (NVIDIA, AMD, Apple Silicon), CPU, RAM
- **Model compatibility** — color-coded table showing what runs on your hardware
- **Model browser** — 370+ models from ollama.com library, with search and filter
- **Download models** — pull models via Ollama API with progress bar
- **Cross-platform** — Linux, macOS, Windows (single binary, no CGO)

## Install

### From source (requires Go 1.24+)

```bash
git clone https://github.com/putzeys/putzeys-cli.git
cd putzeys-cli
go install .
```

The binary will be installed to `~/go/bin/putzeys-cli`. Make sure `~/go/bin` is in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="$HOME/go/bin:$PATH"
```

### Cross-compile for other platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o putzeys-cli-linux .

# Windows
GOOS=windows GOARCH=amd64 go build -o putzeys-cli.exe .

# Mac Intel
GOOS=darwin GOARCH=amd64 go build -o putzeys-cli-intel .

# Mac Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o putzeys-cli-arm .
```

## Usage

### Interactive TUI (recommended)

```bash
putzeys-cli
```

Opens a full-screen TUI with 4 tabs:

| Tab | Content |
|---|---|
| **1 Hardware** | CPU, GPU, RAM detected |
| **2 Compatibilidade** | Color-coded table for 370+ models, cycle quantization with `q`, search with `/` |
| **3 Modelos** | Browse remote/local models, toggle with `l`, search with `/` |
| **4 Pull** | Type model name, check compatibility, download with progress bar |

#### Keyboard shortcuts

| Key | Action |
|---|---|
| `Tab` / `1-4` | Switch tab |
| `j/k` or `Up/Down` | Navigate list |
| `/` | Search (Compatibilidade & Modelos tabs) |
| `Enter` | Confirm action |
| `Esc` | Cancel / exit search |
| `q` (Compatibilidade) | Cycle quantization: Q4_K_M → Q8_0 → FP16 |
| `l` (Modelos) | Toggle local/remote |
| `ctrl+c` | Quit |

### CLI commands

```bash
# Detect your hardware
putzeys-cli scan

# Check which models can run on your machine
putzeys-cli check

# Search for specific models
putzeys-cli check --search llama

# Check with different quantization
putzeys-cli check --quant FP16

# JSON output
putzeys-cli check --json

# List available models
putzeys-cli models --search llama

# List locally installed models
putzeys-cli models --local

# Download a model (requires Ollama running)
putzeys-cli pull llama3.2:1b
```

## Compatibility Verdicts

| Status | Meaning |
|---|---|
| **CAN RUN** (green) | Model fits in GPU VRAM |
| **DEGRADED** (yellow) | Needs CPU/RAM offload (slower) |
| **CAN'T RUN** (red) | Insufficient total memory |

## Configuration

Create `~/.config/putzeys-cli/config.yaml`:

```yaml
cli_name: "putzeys-cli"
ollama_host: "http://localhost:11434"
model_source: "ollamadb"
vram_overhead_percent: 20.0
gpu_memory_fraction: 0.75  # Apple Silicon: % of unified memory for GPU
```

## Requirements

- **Go 1.24+** to build from source
- [Ollama](https://ollama.com) must be running for `models --local` and `pull` commands
- GPU detection requires: `nvidia-smi` (NVIDIA), `rocm-smi` (AMD), or Apple Silicon

## License

MIT
