# Jin: a terminal coding agent in Go

![jin in a terminal](docs/screenshot.png)

Jin runs as one binary. It reads files, edits code and runs commands through a model API.
Sessions, a todo list, background tasks, undo, themes and headless mode are built in.
Choose a provider and model on the first run; there is no config file to write.

Save reusable instructions as prompts and call them with `#prompt-name`. To add a tool,
install its CLI and describe it in a hook. Jin calls it through the shell, without a
plugin or MCP layer. See [Extending jin](docs/extending.md).

## Install

macOS and Linux, arm64 and amd64:

```sh
curl -fsSL https://raw.githubusercontent.com/ethanhamilthon/jin/main/install.sh | sh
```

The script downloads the latest release, checks its SHA-256 and installs `jin` into
`/usr/local/bin` (or `~/.local/bin`). Run it again to update: it replaces the installed
`jin` and prints the version it installed. `JIN_VERSION=v0.7.2` pins a release. Later,
`jin update` does the same from jin itself. Then run `jin` in the project you want to work
on.

## Connect a model

The first start shows a setup screen: choose OpenAI Responses, OpenAI Chat Completions or
Anthropic, then enter the URL, key, model and effort. That covers OpenAI, Anthropic,
OpenRouter and local servers. Add more providers later with `/provider` and switch between
them.

Jin does not sign in with subscriptions (ChatGPT Plus/Pro, Claude Pro/Max and so on). To
use one, run [EasyCLIProxyAPI](https://github.com/router-for-me/EasyCLIProxyAPI), a desktop
app that serves your subscription as a local OpenAI-compatible API, and add its address in
`/provider`.

## Build from source

```sh
git clone https://github.com/ethanhamilthon/jin && cd jin
make build        # bin/jin, keeps its data in ~/.jin-dev
make install      # release build into /usr/local/bin, data in ~/.jin
```

Requires Go 1.27 or newer. Forks are welcome; read [AGENTS.md](AGENTS.md) for the rules of
the code base.

## Benchmarks

In the latest run (Aider Polyglot, 20 Python tasks, one run each), Codex passed 12 tasks,
pi 10, opencode 9 and jin v0.7.0 8. All used `gpt-6-luna` with effort `high`. Jin averaged
47 seconds and about 17k input tokens per task; Codex averaged 49 seconds and about 48k.
This is a small sample, and the v0.7.2 prompt changes have not been benchmarked.
Results and method: [docs/benchmarks.md](docs/benchmarks.md).

## Security

The agent has `bash` with no restrictions and no confirmation dialogs. `bash`, `write` and
`edit` act directly on your machine with your permissions. Run jin only in projects where
that is acceptable, and review project hooks before you trust a repository.

## Docs

[docs/](docs/README.md) covers keys and slash commands, sessions, hooks and prompts,
headless mode (`jin -p`), background tasks, settings and the database. Jin reads these
docs when you ask it about its own commands.

## License

[MIT](LICENSE) © 2026 Yerdana Yerbol.
