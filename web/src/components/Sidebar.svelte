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
  @media (max-width: 700px) {
    aside {
      position: fixed; top: 0; bottom: 0; left: 0; z-index: 15; width: 84%; max-width: 360px;
      padding-top: env(safe-area-inset-top); padding-bottom: env(safe-area-inset-bottom);
    }
  }
</style>
