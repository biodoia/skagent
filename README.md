# 🚀 SkAgent

**AI-Powered Spec-Driven Development Assistant**

SkAgent is a terminal-based AI assistant that helps you plan and execute software projects using spec-driven development methodologies. It integrates with GitHub Spec-Kit, Plandex, and various AI providers to streamline your development workflow.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ Features

- 🤖 **Multi-Provider AI Support** - OpenRouter (35+ free models), Claude Max, Gemini CLI, DeepSeek, and more
- 📝 **Spec-Driven Development** - Integration with GitHub Spec-Kit workflow
- 🔧 **Tool Integration** - GitHub, web search, and planning tools built-in
- 🎨 **Beautiful TUI** - Modern terminal interface with Charm stack
- ⚡ **Autonomous Mode** - Let the agent work independently on your project
- 🆓 **100% Free Options** - Use OpenRouter's free models at zero cost

## 📦 Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/biodoia/skagent.git
cd skagent

# Build
make build

# Install to ~/bin (no sudo required)
make install-user

# Or install system-wide
make install
```

### Pre-built Binaries

Download from [Releases](https://github.com/biodoia/skagent/releases).

## 🚀 Quick Start

```bash
# Run SkAgent (first run will start setup wizard)
skagent

# Or run setup directly
skagent setup
```

### First-Time Setup

The setup wizard will guide you through:
1. Selecting an AI provider
2. Entering API keys (if required)
3. Choosing a model

**Recommended: OpenRouter** - 35+ free models, no credit card required!

Get your free API key at [openrouter.ai](https://openrouter.ai)

## 🎮 Usage

### Commands

| Command | Description |
|---------|-------------|
| `/auto` | Toggle autonomous mode |
| `/provider` | Show current AI provider |
| `/models` | List available free models |
| `/clear` | Clear conversation |
| `/help` | Show help |
| `/quit` | Exit application |

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+C` | Exit |
| `Esc` | Cancel/Exit |
| `↑/↓` | Scroll messages |

### Example Session

```
You: I want to build a CLI tool for managing dotfiles

Agent: I'll help you create a comprehensive plan for a dotfiles manager CLI.

Based on my analysis, here's what I recommend:

## Project: DotManager

### Key Features
1. Symlink management for config files
2. Git-based backup and sync
3. Cross-platform support (Linux, macOS, Windows)
...

Shall I generate the full SpecKit specification?
```

## 🔧 Supported AI Providers

| Provider | Type | Cost | Setup |
|----------|------|------|-------|
| **OpenRouter** | API | Free models available | API key |
| **Claude Max** | OAuth | Subscription | `claude` CLI |
| **Gemini CLI** | CLI | Free tier | `gemini` CLI |
| **Codex CLI** | CLI | Varies | `codex` CLI |
| **Kimi** | API | Free tier | API key |
| **GLM-4** | API | Free tier | API key |
| **DeepSeek** | API | Very cheap | API key |
| **Minimax** | API | Varies | API key |

### Best Free Models on OpenRouter

- ⭐ **Qwen3 Coder 480B** - Best for coding tasks
- ⭐ **Mistral Devstral** - Developer-focused
- ⭐ **DeepSeek R1** - Complex reasoning
- ⭐ **Gemini 2.0 Flash** - 1M context window
- ⭐ **Llama 3.3 70B** - Best open-source general

## 📐 Spec-Kit Workflow

SkAgent follows the GitHub Spec-Kit methodology:

1. **SPECIFY** → Define what and why
2. **PLAN** → Technical blueprint
3. **TASKS** → Atomic work items
4. **IMPLEMENT** → Build with TDD

### The Nine Articles

1. Library-First: Features as standalone libraries
2. CLI Mandate: All functionality via CLI
3. Test-First: No implementation without tests
4. Simplicity: Max 3 projects per implementation
5. Anti-Abstraction: Use frameworks directly
6. Integration-First: Test in realistic environments

## 🛠️ Development

```bash
# Download dependencies
make deps

# Build
make build

# Run tests
make test

# Cross-compile for all platforms
make cross-compile

# Clean build artifacts
make clean
```

### Project Structure

```
skagent/
├── cmd/skagent/          # Main application entry
├── internal/
│   ├── ai/               # AI providers and client
│   ├── config/           # Configuration management
│   ├── docs/             # Documentation loader
│   ├── setup/            # Setup wizard
│   ├── tools/            # Tool integrations
│   └── tui/              # Terminal UI
├── build/                # Build output
├── Makefile              # Build automation
└── README.md
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- [Charm](https://charm.sh) - Beautiful TUI components
- [OpenRouter](https://openrouter.ai) - Free AI model access
- [GitHub Spec-Kit](https://github.com/github/spec-kit) - Spec-driven methodology

---

Made with ❤️ by [@biodoia](https://github.com/biodoia)
