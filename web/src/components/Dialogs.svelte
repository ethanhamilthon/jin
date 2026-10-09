<script lang="ts">
  import { app } from "../lib/app.svelte";
  import ModelDialog from "./dialogs/ModelDialog.svelte";
  import ProvidersDialog from "./dialogs/ProvidersDialog.svelte";
  import TasksDialog from "./dialogs/TasksDialog.svelte";
  import ContextDialog from "./dialogs/ContextDialog.svelte";
  import RewindDialog from "./dialogs/RewindDialog.svelte";
  import UndoDialog from "./dialogs/UndoDialog.svelte";
  import ProjectDialog from "./dialogs/ProjectDialog.svelte";
  import TrustDialog from "./dialogs/TrustDialog.svelte";

  const id = $derived(app.dialog?.session ?? "");
</script>

{#if app.dialog?.name === "model"}<ModelDialog {id} />
{:else if app.dialog?.name === "providers"}<ProvidersDialog />
{:else if app.dialog?.name === "tasks"}<TasksDialog />
{:else if app.dialog?.name === "context"}<ContextDialog {id} />
{:else if app.dialog?.name === "rewind"}<RewindDialog {id} />
{:else if app.dialog?.name === "undo"}<UndoDialog {id} />
{:else if app.dialog?.name === "addproject"}<ProjectDialog />
{:else if app.dialog?.name === "trust"}<TrustDialog dir={app.dialog.arg ?? ""} />
{/if}
