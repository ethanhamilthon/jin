## Project development rules

1. Keep files small: ~100 lines max per file.
2. Clear architectural separation: UI, Core, Provider.
3. Do not make design decisions without the user's agreement. Always ask when it is not fully clear how a feature should work.
4. Adding libraries or dependencies requires the user's confirmation.
5. Minimal comments. The code itself, the architecture, and the naming should explain everything.
6. Always commit after a feature is done.
7. English only: all UI text, code comments, commit messages, and project reports.

## Running the web UI in dev

Start two background tasks (never in the foreground):

1. Backend: `go run . web --port 7374 --no-open`. Port 7373 is often taken by the user's own `jin web`; never stop that process, pick a free port. The backend prints `http://127.0.0.1:<port>/?token=<token>`.
2. Vite: `cd web && JIN_WEB=http://127.0.0.1:<port> npm run dev` (serves `http://localhost:5173/`, proxies `/api` to the backend).

Tell the user to open this URL once so the backend sets the auth cookie:
`http://localhost:5173/api/state?token=<token>`
Then `http://localhost:5173/` works. Opening `/?token=...` on the Vite address does not work: Vite serves it itself, so no cookie is set and `/api/events` answers 401 (blank page).

Notes:
- `internal/web/dist` is the embedded production build. If its files belong to root, `npm run build` fails with `ENOTEMPTY`; build elsewhere with `npx vite build --outDir /tmp/jin-dist --emptyOutDir` or ask the user to fix ownership.
- A root-owned `web/node_modules/.vite-temp` makes `npm run check` fail with `EACCES`; remove that empty directory.
