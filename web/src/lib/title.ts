import { post } from "./api";
import { fail } from "./app.svelte";

// generateTitle has the title model name the session now. The pane header
// and the session picker update from the state and sessions events.
export async function generateTitle(id: string) {
  await post<{ title: string }>(`/api/sessions/${id}/generate-title`).catch(fail);
}
