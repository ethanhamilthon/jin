package cli

const usage = `jin: a minimal terminal coding agent

Usage:
  jin                          open the TUI
  jin -p [flags] [prompt...]   run one request without the TUI (see docs/headless.md)
                               --cwd <dir> runs in that directory, --provider <id> uses a saved provider,
                               --max-cost <usd> and --max-turns <n> limit the run
  jin models [--all] [--provider <id>]
                               list the models you use, with price and context window
  jin refresh-models           refresh the cached model list
  jin sessions list [--all] [--format json]
                               list saved sessions (newest first)
  jin sessions search <words...> [--all] [--format json]
                               search session titles and user messages
  jin export <session-id> [--md|--json]
                               print a saved session (an id prefix is enough)
  jin hooks list               list global and project hooks
  jin hooks add <url|path> [--name n] [--project]
                               copy a markdown hook into ~/.jin/hooks
                               (or ./.jin/hooks with --project)
  jin update [--check]         install the latest release over this binary
                               (--check only tells whether one exists)
  jin --version                print the version
  jin --help                   print this help
`
