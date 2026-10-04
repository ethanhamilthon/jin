package cli

const usage = `jin: a minimal terminal coding agent

Usage:
  jin                          open the TUI
  jin -p [flags] [prompt...]   run one request without the TUI (see docs/headless.md)
                               --cwd <dir> runs in that directory, --provider <id> uses a saved provider
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
  jin async run "<cmd>" --session <id>
                               run a command in the background; its result
                               comes back to the session as a message
  jin async check --id <task> [--limit <n>]
                               status and the last n characters of the output
  jin async input --id <task> "<text>"
                               write a line to the stdin of a task
  jin async stop --id <task>   stop a task
  jin update [--check]         install the latest release over this binary
                               (--check only tells whether one exists)
  jin --version                print the version
  jin --help                   print this help
`
