# Reading the jin database with sqlite3

All state is in one SQLite file: `~/.jin/jin.db` (`~/.jin-dev/jin.db` for source
builds). It uses WAL mode, so reading while jin runs is safe. Open it read-only to be
sure:

```sh
sqlite3 -readonly ~/.jin/jin.db
sqlite3 -readonly -header -column ~/.jin/jin.db "SELECT ..."
```

Do not write to it while jin runs unless you know what you do.

## Tables

```sql
sessions(id TEXT PK, path TEXT, model TEXT, effort TEXT, title TEXT,
         created_at INTEGER, updated_at INTEGER,       -- unix seconds
         input_tokens INTEGER, output_tokens INTEGER,
         context_tokens INTEGER, cost REAL)
messages(id INTEGER PK AUTOINCREMENT, session_id TEXT, data TEXT)   -- data is JSON
settings(key TEXT PK, value TEXT)
unread_sessions(session_id TEXT PK)
running_sessions(session_id TEXT PK, pid INTEGER)
todos(session_id TEXT, position INTEGER, text TEXT, status TEXT)   -- PK (session_id, position)
todo_state(session_id TEXT PK, edited INTEGER)   -- 1 when the user edited the list
```

- A session row is created on the first prompt, so empty chats leave no trace.
- `path` is the working directory the session belongs to.
- `messages.data` is an OpenAI-style message: `role`, `content` (string, or a list of
  parts when the message has images), optional `reasoning_content`, `tool_calls`,
  `tool_call_id`. Roles: `user`, `assistant`, `tool`.
- A compaction summary is a user message starting with `<conversation-summary>`.

## Settings keys

| Key | Value |
| --- | --- |
| `provider.base_url` | provider URL |
| `provider.api_key` | API key, plain text. Never print it |
| `model`, `effort` | current model and reasoning effort |
| `models.scope` | JSON list of enabled models, empty means all |
| `models.efforts` | JSON map model → last effort |
| `editor` | `nano`, `vim` or `hx` |
| `sound.enabled`, `sound.when`, `sound.volume` | `1`/`0`, `always`/`blur`, 0-100 |
| `fold` | 0, 1 or 2 |
| `hooks.disabled` | JSON list of switched-off hook names |
| `tools.disabled` | JSON list of switched-off tool names |
| `docs.enabled` | `1` when Jin docs are on |

Show settings without the secret:

```sh
sqlite3 -readonly -column ~/.jin/jin.db "SELECT key, value FROM settings WHERE key <> 'provider.api_key'"
```

## Recipes

Recent sessions with cost:

```sql
SELECT substr(id,1,8) AS id, datetime(updated_at,'unixepoch','localtime') AS updated,
       model, input_tokens, output_tokens, round(cost,4) AS cost, title
FROM sessions ORDER BY updated_at DESC LIMIT 20;
```

Sessions of one project:

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
