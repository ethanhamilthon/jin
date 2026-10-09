# Reading the jin database with sqlite3

All state is in one SQLite file: `~/.jin/jin.db` (`~/.jin-dev/jin.db` for source
builds). It uses WAL mode, so reading while jin runs is safe. Open it read-only to be
sure:

```sh
sqlite3 -readonly ~/.jin/jin.db
sqlite3 -readonly -header -column ~/.jin/jin.db "SELECT ..."
```

Do not write to it while jin runs unless you know what you do.

Every jin process (TUI, `jin -p`) holds a shared `flock` on
`~/.jin/.jin.lock` while it runs. The Reset and Swap config rows of `/settings` can reset or
swap the folder
only when Jin gets the lock exclusively; otherwise it refuses with "close other jin windows
first". When the last step of a swap fails, the earlier steps are undone.

## Tables

```sql
projects(id TEXT PK, path TEXT UNIQUE, name TEXT, last_session_id TEXT,
         created_at INTEGER, last_opened_at INTEGER,  -- unix seconds
         archived INTEGER)  -- 1: hidden from the sidebar of jin web, nothing deleted
sessions(id TEXT PK, path TEXT, project_id TEXT NULL FK projects(id),
         model TEXT, effort TEXT, title TEXT, provider TEXT,
         created_at INTEGER, updated_at INTEGER,       -- unix seconds
         input_tokens INTEGER, output_tokens INTEGER,
         context_tokens INTEGER, cost REAL)
messages(id INTEGER PK AUTOINCREMENT, session_id TEXT, data TEXT)   -- data is JSON
settings(key TEXT PK, value TEXT)
unread_sessions(session_id TEXT PK)
running_sessions(session_id TEXT PK, pid INTEGER)
todos(session_id TEXT, position INTEGER, text TEXT, status TEXT)   -- unused since v0.10, kept for older data
todo_state(session_id TEXT PK, edited INTEGER)   -- unused since v0.10, kept for older data
file_changes(id INTEGER PK, session_id, turn INTEGER, path, existed INTEGER,
             before TEXT, after TEXT)   -- edit/write results for /undo
```

- A project row records a normalized working directory even before it has a saved session.
  Project records are separate from sessions; adding a project does not create project files.
  Migration registers paths found in existing sessions and preserves the legacy `path` column.
- A session row is created on the first prompt, so an empty session leaves no session row.
- `sessions.path` remains the session's working directory for compatibility. `project_id`
  links it to `projects.id`; older or pathless rows may have no link. Existing session paths
  continue to be used when sessions are opened.
- `provider` is the id of the provider saved for that session. Changing the default provider
  does not change a live session's provider. An empty value uses the saved active provider
  when the session is opened.
- `messages.data` is an OpenAI-style message: `role`, `content` (string, or a list of
  parts when the message has images), optional `reasoning_content`, `tool_calls`,
  `tool_call_id`. Roles: `user`, `assistant`, `tool`. An assistant message may also hold
  `jin_native` (`kind`, `model`, `items`): the provider's own output items, such as
  encrypted reasoning or signed thinking blocks, replayed to the same kind and model.
- A compaction summary is a user message starting with `<conversation-summary>`.
- `running_sessions` is the owner of a session: the pid of the jin process whose request
  runs in it. A process cannot claim a session that another live process owns, and only
  the owner removes the row.
- Background tasks are not stored: the jin process that runs them keeps them in memory
  (see [tasks.md](tasks.md)). Databases of earlier versions may still hold the unused
  `async_tasks` and `async_events` tables.
- `file_changes` keeps, per agent turn, each file `edit` or `write` changed with its
  content before and after. `/undo` reverts the newest turn and deletes its rows.

## Settings keys

| Key | Value |
| --- | --- |
| `providers` | JSON list of providers: `id`, `name`, `kind` (`openai`, `responses` or `anthropic`), `base_url`, `api_key` (plain text, never print it) |
| `provider.active` | id of the default provider for new sessions |
| `provider.base_url`, `provider.api_key` | the v0.2 provider. v0.3 copies them into `providers` once and keeps updating them for the active provider, so v0.2 can still read them. Never print the key |
| `provider.stall_timeout` | seconds a response stream may stay silent before it is cancelled and tried once more; no key means a limit by reasoning effort, 90 s to 600 s |
| `update.latest`, `update.checked` | newest release tag seen and the Unix time of the last check (at most every 6 hours) |
| `model`, `effort` | current model and reasoning effort |
| `models.scope`, `models.scope.<id>` | JSON list of enabled models of a provider, empty means all. The plain key belongs to the `default` provider |
| `models.efforts` | JSON map model → last effort |
| `editor` | `nano`, `vim` or `hx` |
| `sound.enabled`, `sound.when`, `sound.volume` | `1`/`0`, `always`/`blur`, 0-100 |
| `ui.motion` | `off`, `slow`, `normal` (default) or `fast` |
| `ui.theme` | name of the color theme from `/theme` (built-in or from `~/.jin/themes`); no key means Jin Original |
| `web.accent` | accent color of the `jin web` page, set in its Settings (see [web.md](web.md)); no key means the default accent |
| `fold` | 0 everything, 1 no tool calls, 2 messages only, 3 tool output |
| `hooks.disabled` | JSON list of switched-off hook names; a project hook is listed by its file path |
| `hooks.trust` | JSON map folder → `true`/`false`: may the project hooks in `.jin/hooks` of that folder run |
| `tools.disabled` | JSON list of switched-off tool names |
| `prompts.disabled` | JSON list of switched-off prompt names |
| `release.seen` | the version whose "What's new" section the intro showed last |
| `migrated.0_4` | `1` after the 0.4 upgrade step ran |
| `models.cache`, `models.cache.<id>` | JSON list of model ids of a provider, from `jin refresh-models` |
| `models.levels`, `models.levels.<id>` | JSON map model → reasoning levels of a provider, from `jin refresh-models --efforts` |
| `docs.enabled` | `0` when Jin docs are switched off; anything else (or no key) means on |

Show settings without the secret:

```sh
sqlite3 -readonly -column ~/.jin/jin.db "SELECT key, value FROM settings WHERE key NOT IN ('provider.api_key', 'providers')"
```

## Recipes

Recent sessions with cost:

```sql
SELECT substr(id,1,8) AS id, datetime(updated_at,'unixepoch','localtime') AS updated,
       model, input_tokens, output_tokens, round(cost,4) AS cost, title
FROM sessions ORDER BY updated_at DESC LIMIT 20;
```

Sessions linked to a registered project:

```sql
SELECT p.name, s.id, s.title
FROM projects p JOIN sessions s ON s.project_id = p.id
WHERE p.path = '/Users/me/projects/app' ORDER BY s.updated_at DESC;
```

For legacy sessions that have no `project_id`, query the preserved path column:

```sql
SELECT id, title FROM sessions WHERE path = '/Users/me/projects/app' ORDER BY updated_at DESC;
```

Total spend per model:

```sql
SELECT model, count(*) AS sessions, round(sum(cost),4) AS cost
FROM sessions GROUP BY model ORDER BY cost DESC;
```

Spend per day:

```sql
SELECT date(updated_at,'unixepoch','localtime') AS day, round(sum(cost),4) AS cost
FROM sessions GROUP BY day ORDER BY day DESC LIMIT 14;
```

Read one conversation (role and text; needs the sqlite JSON1 functions, built in):

```sql
SELECT id, json_extract(data,'$.role') AS role,
       substr(coalesce(json_extract(data,'$.content'),''),1,200) AS text
FROM messages WHERE session_id LIKE 'abcd1234%' ORDER BY id;
```

Find sessions that mention a word:

```sql
SELECT DISTINCT s.id, s.title FROM messages m JOIN sessions s ON s.id = m.session_id
WHERE m.data LIKE '%needle%';
```

Tool calls made in a session:

```sql
SELECT json_extract(c.value,'$.function.name') AS tool,
       substr(json_extract(c.value,'$.function.arguments'),1,120) AS args
FROM messages m, json_each(m.data,'$.tool_calls') c
WHERE m.session_id LIKE 'abcd1234%' ORDER BY m.id;
```

Sessions that are running right now:

```sql
SELECT s.title, r.pid FROM running_sessions r JOIN sessions s ON s.id = r.session_id;
```
