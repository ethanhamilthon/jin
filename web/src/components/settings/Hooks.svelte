<script lang="ts">
  import { app } from "../../lib/app.svelte";
  import HookList from "./HookList.svelte";
</script>

<HookList project={false} dir={app.projectPath} />
<p class="set-note">Hooks of one project and their trust live in that project's Project pane.</p>
