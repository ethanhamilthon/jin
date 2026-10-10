package cli

const usage = `jin: a minimal terminal coding agent

Usage:
  jin                          open the TUI
  jin web [--port N] [--no-open] [--cwd DIR]
                               open jin in the browser (see docs/web.md)
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
  jin sessions compact|handoff|rewind|context|reload <id>
                               act on one saved session (see docs/headless.md)
  jin projects list [--all] [--format json]
                               list registered projects (--all includes archived ones)
  jin projects add <path>
                               register a folder as a project, or restore it
  jin projects rename <path> <name>
                               set the display name of a project
  jin projects archive <path>
                               hide a project from the session picker of jin web
  jin projects restore <path>
                               show an archived project again
  jin export <session-id> [--md|--json]
                               print a saved session (an id prefix is enough)
  jin hooks list               list global and project hooks
  jin hooks add <url|path> [--name n] [--project]
                               copy a markdown hook into ~/.jin/hooks
                               (or ./.jin/hooks with --project)
  jin hooks render             print the enabled hooks, filled in (for the system prompt)
  jin docs [--list|<page>]     read the documentation built into jin
  jin update [--check]         install the latest release over this binary
                               (--check only tells whether one exists)
  jin --version                print the version
  jin --help                   print this help
`
