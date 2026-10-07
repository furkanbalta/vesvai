---
icon: lucide/terminal
---

# CLI

The `vesvai` binary is both the TUI launcher and a full command-line tool. Run
`vesvai help` or `vesvai <command> --help` at any time.

## Invocation behavior

| Invocation | Behavior |
|---|---|
| `vesvai` on a terminal | Launches the [TUI](tui.md) |
| `vesvai "message"` | Runs the orchestrator on the message (headless if stdin is not a terminal) |
| `echo "msg" \| vesvai` | Reads stdin, appends it to any arguments, and runs the orchestrator |
| `vesvai <command>` | Runs the subcommand |

Errors are printed to stderr as `vesvai: <error>` and the process exits with code
`1`. Successful commands exit `0`. Interactive prompts cancelled with ++ctrl+c++
also exit `1`.

## `vesvai run`

Run the orchestrator agent on a message, streaming its progress, tools, thinking,
and responses live. Without a message (or with `--chat`) it starts an interactive
chat in the same session.

```bash
vesvai run "Explain the build system"
vesvai run --provider openai --model gpt-4o "Review internal/agent"
vesvai run --model smart-router "Fix the flaky tests"   # needs smart_router.enabled
vesvai run --session 3f2b1c... "Continue where we left off"
```

| Flag | Default | Description |
|---|---|---|
| `--provider` | auto | Provider to use |
| `--model` | auto | Model to use. `smart-router` selects the best model per task/agent (see [Smart Router](../providers-and-models.md#smart-router)) |
| `--select-model` | `false` | List available models and let you pick one (overrides `--model`) |
| `--session` | — | Session ID to continue |
| `--select-session` | `false` | List sessions and pick one to continue |
| `--file` | — | File to attach (repeatable) |
| `--show-thinking` | `false` | Print raw reasoning content instead of a `Thinking...` indicator |
| `--show-subagent` | `false` | Print subagent messages instead of a `Subagent` indicator |
| `--chat` | `false` | Keep chatting after the first run ends |

Model resolution follows the rules in
[Providers & Models](../providers-and-models.md#model-selection). When resuming a
session without an explicit provider/model, the session's stored model is reused.

### Output

Streaming output is plain text, with ANSI colors only when stdout is a terminal:

```
Running orchestrator (openai/gpt-4o)

> Explain the build system
tool [orchestrator] read(path=Makefile)
result [orchestrator] read: Path: Makefile | ...
The build is driven by a small Makefile...
• orchestrator finished (2 iterations, 1.2K tokens, $0.003100)
```

- Tool calls are shown as `tool [agent] name(args...)`; results are truncated to 500
  characters.
- `--show-thinking` prints the model's raw reasoning; otherwise a `Thinking...`
  spinner is shown on a TTY.
- `--show-subagent` prints subagent content instead of a `Subagent <name>` line.
- When the context is compacted mid-run, a line like
  `↻ Context compacted (sliding-window) — 30 messages, 8000 tokens` is printed.

### Chat mode

With `--chat`, or when the message is empty, `run` enters a line-based chat loop:

```
chat mode: type a message, exit/quit to end
> add a test for the parser
...
```

Exit words: `exit`, `quit`, `/exit`, `/quit`, `bye`.

### Attachments

`--file` can be repeated. The MIME type is inferred from the extension — images
(`.png`, `.jpg`, `.jpeg`, `.gif`, `.webp`) and audio (`.mp3`, `.wav`) are sent as
attachments, PDFs and text files as files. See
[Adding Context](../features/adding-context.md).

### Session selection

`--select-session` opens a paginated picker (10 per page) filtered to sessions from
the current project directory, showing title, provider/model, and creation time.

## `vesvai serve`

Start a server: the HTTP REST API by default, or ACP with `--acp`.

```bash
vesvai serve                          # HTTP API on 127.0.0.1:8080
vesvai serve --port 9000 --host 0.0.0.0
vesvai serve --acp                    # ACP over HTTP at /acp
vesvai serve --acp --stdio            # ACP over stdin/stdout
```

| Flag | Default | Description |
|---|---|---|
| `--port` | config (`8080`) | Port to listen on |
| `--host` | config (`127.0.0.1`) | Host to bind to |
| `--acp` | `false` | Start the ACP server instead of the HTTP API |
| `--stdio` | `false` | Use the stdio transport (only with `--acp`) |

See [HTTP](http.md) and [ACP](acp.md).

## `vesvai login`

Add or update a provider with an API key.

```bash
vesvai login
vesvai login --provider anthropic --api-key sk-ant-...
```

| Flag | Description |
|---|---|
| `--provider` | Provider name. Interactive picker when omitted |
| `--api-key` | API key. Masked prompt when omitted; may be empty |

The provider's model list is fetched before the config is saved, so a login only
succeeds when the credentials work.

## `vesvai connect`

Connect this machine's agent to a website backend so you can chat with it from a
browser. Only a pairing token is needed — create one on the website under
**Agents → Connect a device**. The full walkthrough lives in
[Connect](../features/connect.md).

```bash
vesvai connect <token>     # first run: connect and remember the token
vesvai connect             # every run after that
```

| Subcommand | Description |
|---|---|
| `vesvai connect login [token]` | Save a pairing token without connecting (masked prompt when omitted) |
| `vesvai connect logout` | Forget the saved token (`--all` also resets the device identity) |
| `vesvai connect status` | Relay, saved token (masked), and device identity |
| `vesvai connect whoami` | This machine's device ID and public key |
| `vesvai connect methods` | List every method the website may call |
| `vesvai connect pair` | Print the device identity to register with a backend |

| Flag | Description |
|---|---|
| `--token` | Pairing token (may also be passed as the argument) |
| `--auto-approve` | Skip local confirmation for sensitive remote actions (not recommended) |
| `--insecure` | Allow `ws://` to a non-loopback host (not recommended) |
| `--device-name` | Human-friendly name for this device |
| `--allowed-method` | Restrict methods the website may call (repeatable, supports `prefix*`) |
| `--max-concurrency` | Maximum parallel runs (default from config) |
| `--response-timeout` | Seconds a run may produce no output before it is failed |

The relay endpoint is built into the app, so nothing but a token is needed; point a
self-hosted website at `connect.relay_url` in `~/.vesvai/vesvai.json`. The pairing
token is stored there too, and is only ever displayed masked.

## `vesvai providers`

| Command | Description |
|---|---|
| `vesvai providers add` | Add a provider with full config options |
| `vesvai providers list` | List providers with masked key and cached model count |
| `vesvai providers refresh [--provider NAME]` | Re-fetch models, bypassing the cache |
| `vesvai providers remove NAME` | Remove a provider and its cached models |

`vesvai providers add` exposes every provider config field:

```bash
vesvai providers add --name anthropic --api-key [**REDACTED**]
vesvai providers add --driver openai --base-url http://localhost:11434/v1 --api-key ollama
```

| Flag | Description |
|---|---|
| `--name` | Registered provider name, or a custom name when combined with `--driver`. Interactive picker (with a "Custom endpoint" option) when omitted |
| `--driver` | Driver for custom endpoints (`openai`, `claude`, `gemini`). Requires `--base-url` |
| `--api-key` | API key. Masked prompt in interactive mode; may be empty |
| `--base-url` | Base URL. Required for driver entries |
| `--header KEY=VALUE` | Request header (repeatable) |
| `--timeout` | Request timeout in seconds |
| `--max-retries` | Maximum retry count |

Running it with any of `--name`, `--driver`, or `--base-url` skips all prompts.
Models are fetched in the background after saving; a failed fetch does not block
the add — run `vesvai providers refresh` to retry. Registered names use built-in
endpoints; `--base-url`/`--driver` only affect driver entries and custom names.

## `vesvai models`

List cached models.

```bash
vesvai models
vesvai models --provider deepseek
```

| Flag | Description |
|---|---|
| `--provider` | Only show models for this provider |

## `vesvai doctor`

Check provider connectivity and look for updates.

```
checking providers...
  openai OK (64 models)
  groq FAIL  llm: provider error: status 401
checking for updates...
  You are running the latest version
```

Each provider is queried with a fresh (uncached) model fetch. If a newer release
exists, `doctor` asks `Do you want to update? (y/N)` and can install it.

## `vesvai config`

```bash
vesvai config show
```

`show` prints the resolved configuration as pretty JSON with API keys masked. It is
the only `config` subcommand.

## `vesvai sessions`

| Command | Description |
|---|---|
| `vesvai sessions list` | List sessions for the current project |
| `vesvai sessions show ID` | Show session metadata and messages |

`list` flags:

| Flag | Default | Description |
|---|---|---|
| `--search` | — | Search sessions by title |
| `--all` | `false` | Include sessions from all projects |
| `--page` | `1` | Page number |
| `--size` | `50` | Page size |

```bash
vesvai sessions list --all --search refactor
vesvai sessions show 3f2b1c8e-...
```

See [Sessions](../features/sessions.md).

## `vesvai logs`

View application logs.

| Flag | Default | Description |
|---|---|---|
| `--search` | — | Search log messages |
| `--level` | — | Filter by `DEBUG`, `INFO`, `WARN`, or `ERROR` |
| `--page` | `1` | Page number |
| `--size` | `50` | Page size |

```bash
vesvai logs --level ERROR
vesvai logs --search "mcp" --size 100
```

With the SQLite logger (default) logs come from `~/.vesvai/logs.db`. With the
console logger, only the current process's in-memory buffer is available.

## `vesvai cache`

```bash
vesvai cache clear
vesvai cache clear --provider openai
```

| Flag | Description |
|---|---|
| `--provider` | Only clear this provider's cached models |

## `vesvai files`

Print system file paths and whether they exist:

```
config  exists   /home/user/.vesvai/vesvai.json
logs    exists   /home/user/.vesvai/logs.db
cache   exists   /home/user/.vesvai/cache.db
sessions exists  /home/user/.vesvai/sessions.db
lsps    missing  /home/user/.vesvai/lsps
```

## `vesvai mcp`

| Command | Description |
|---|---|
| `vesvai mcp list [--global\|--project]` | List configured servers |
| `vesvai mcp add [flags]` | Add a server (interactive unless flags are given) |
| `vesvai mcp remove NAME [--global\|--project]` | Remove a server |
| `vesvai mcp tools [--server NAME]` | List tools exposed by connected servers |

`add` flags: `--global`, `--project`, `--name`, `--transport local|http`,
`--command`, `--arg` (repeatable), `--env KEY=VALUE` (repeatable), `--url`,
`--header KEY=VALUE` (repeatable). See [MCP](../configurations/mcp.md).

## `vesvai lsp`

| Command | Description |
|---|---|
| `vesvai lsp list [--global\|--project]` | List language servers |
| `vesvai lsp add [flags]` | Add a server (interactive unless flags are given) |
| `vesvai lsp remove NAME [--global\|--project]` | Remove a server |
| `vesvai lsp diag --file PATH` | Print cached diagnostics for a file |

`add` flags: `--global`, `--project`, `--name`, `--command`, `--arg` (repeatable),
`--filetype` (repeatable), `--rootmarker` (repeatable), `--download URL`,
`--env KEY=VALUE` (repeatable). See [LSP](../configurations/lsp.md).

## `vesvai tui`

Launch the terminal UI explicitly. Equivalent to running `vesvai` on a terminal.

## `vesvai version`

```bash
vesvai version
# vesvai version 0.1.0
```

## `vesvai update`

Check for a newer release and install it in place.

```
Current version: 0.1.0
Latest version: 0.2.0
Updating to version 0.2.0...
Successfully updated to version 0.2.0
```

## Shell completion

Cobra provides completion scripts for bash, zsh, fish, and PowerShell:

```bash
vesvai completion bash > /etc/bash_completion.d/vesvai
vesvai completion zsh > "${fpath[1]}/_vesvai"
vesvai completion fish > ~/.config/fish/completions/vesvai.fish
```

## Headless and scripting examples

```bash
# One-shot prompt, output to stdout
vesvai run "List the public API of this package"

# Pipe context in
cat error.log | vesvai "Explain the root cause"

# Run in CI with an explicit model and no interactive pickers
vesvai run --provider openai --model gpt-4o-mini "Run the tests and fix failures"

# JSON-free plain text is emitted when stdout is not a terminal
OUTPUT="$(vesvai run --provider groq --model llama-3.3-70b-versatile 'Summarize README.md')"
```

For structured output, use the [HTTP API](http.md) instead.
