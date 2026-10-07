<script lang="ts">
  import ProjectList from "./ProjectList.svelte";
</script>

<aside>
  <ProjectList />
</aside>

<style>
  aside {
    grid-area: side; display: flex; flex-direction: column; min-height: 0;
    background: var(--canvas); border-right: 1px solid var(--raised);
  }
</style>
