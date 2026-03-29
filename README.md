# ollafit

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/putzeys/ollafit?include_prereleases)](https://github.com/putzeys/ollafit/releases)

Check which [Ollama](https://ollama.com) models can run on your hardware. Detects your GPU, VRAM, RAM, and CPU, then shows a color-coded compatibility table for 370+ models.

## Features

- **Interactive TUI** with 4 tabs: Hardware, Compatibility, Models, Pull
- **Hardware detection** for NVIDIA, AMD, Apple Silicon, and CPU-only setups
- **Model compatibility** with color-coded verdicts (CAN RUN / DEGRADED / CAN'T RUN)
- **Model browser** with search and local/remote toggle
- **Download models** directly with progress bar
- **Cross-platform** single binary (Linux, macOS, Windows) -- no CGO

## Install

### GitHub Releases (recommended)

Download the latest binary from [Releases](https://github.com/putzeys/ollafit/releases) and add it to your PATH.

### From source

```bash
go install github.com/putzeys/ollafit@latest
```

Or clone and build:

```bash
git clone https://github.com/putzeys/ollafit.git
cd ollafit
go install .
```

## Usage

### Interactive TUI

```bash
ollafit
```

Opens a full-screen interface with 4 tabs:

| Tab | Description |
|-----|-------------|
| **1 Hardware** | Detected CPU, GPU, RAM |
| **2 Compatibility** | Color-coded table for 370+ models |
| **3 Models** | Browse remote or local models |
| **4 Pull** | Check compatibility and download models |

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `Tab` / `1-4` | Switch tab |
| `j/k` or `Up/Down` | Navigate list |
| `/` | Search (Compatibility & Models) |
| `q` | Cycle quantization: Q4_K_M, Q8_0, FP16 (Compatibility) |
| `l` | Toggle local/remote (Models) |
| `Enter` | Confirm / send to Pull |
| `Esc` | Cancel |
| `Ctrl+C` | Quit |

### CLI commands

```bash
# Detect hardware
ollafit scan

# Check model compatibility
ollafit check
ollafit check --search llama
ollafit check --quant FP16
ollafit check --json

# Browse models
ollafit models
ollafit models --search llama
ollafit models --local

# Download a model (requires Ollama running)
ollafit pull llama3.2:1b
```

## Compatibility verdicts

| Status | Meaning |
|--------|---------|
| **CAN RUN** (green) | Model fits in GPU VRAM |
| **DEGRADED** (yellow) | Needs CPU/RAM offload (slower) |
| **CAN'T RUN** (red) | Insufficient total memory |

## Configuration

Create `~/.config/ollafit/config.yaml`:

```yaml
ollama_host: "http://localhost:11434"
model_source: "ollamadb"
vram_overhead_percent: 20.0
gpu_memory_fraction: 0.75  # Apple Silicon: % of unified memory for GPU
```

Environment variables are also supported with the `OLLAFIT_` prefix (e.g., `OLLAFIT_OLLAMA_HOST`).

## Requirements

- [Ollama](https://ollama.com) running locally for `models --local` and `pull` commands
- GPU detection uses: `nvidia-smi` (NVIDIA), `rocm-smi` (AMD), or Apple Silicon system APIs
- **Go 1.24+** only if building from source

## License

MIT
