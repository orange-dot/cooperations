# Cooperations

Local-first AI mob programming orchestration for software development. The system coordinates multiple models in specialized roles (architect, implementer, reviewer, navigator) using a CLI and a futuristic GPU-accelerated GUI.

## Status

- CLI: usable for local workflows
- GUI: functional 3-panel layout with real-time workflow visualization

## Quick Start

Requirements:
- Go 1.23+
- API keys: `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`

Build:
```bash
cp .env.example .env   # Add your API keys
go build -o coop ./cmd/coop
```

Run CLI:
```bash
./coop run "implement a REST API for users"
./coop status
./coop history
```

Run GUI:
```bash
./coop gui "implement a REST API"       # Real mode (requires API keys)
./coop gui --demo "test the interface"  # Demo mode (stub progress)
```

## Examples

The repository includes two example sets:
- `examples/` (baseline implementations)
- `examples2/` (GPT-5.2 implementations)

Run:
```
go run ./examples
go run ./examples2
```

## GUI

The GUI provides a futuristic 3-panel interface built with [Gio](https://gioui.org):

- **Sidebar** (left): Workflow steps with progress indicators, handoff history
- **Main Panel** (center): Activity log, syntax-highlighted code display
- **Bottom Panel** (appears when needed): Human decision prompts (Approve/Reject/Edit)

Features:
- Dark theme with neon cyan/magenta accents
- Real-time progress streaming from orchestrator
- Syntax highlighting via Chroma
- GPU-accelerated rendering

## Project Layout

- `cmd/coop/` CLI entrypoint
- `internal/orchestrator/` workflow coordination with stream events
- `internal/agents/` role-specific agents (architect, implementer, reviewer, navigator)
- `internal/adapters/` model adapters (Claude, Codex)
- `internal/gui/` Gio-based GUI application
- `internal/gui/widgets/` custom neon-styled widgets
- `internal/gui/stream/` event streaming for real-time updates
- `docs/` strategy and plans
- `examples/`, `examples2/` algorithm demos

## Security

See `SECURITY.md`.

## Contributing

See `CONTRIBUTING.md`.

## License

GNU Affero General Public License v3.0. See `LICENSE`.
