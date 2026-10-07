<script lang="ts">
  import type { Image } from "../lib/draft";

  let { images, remove }: { images: Image[]; remove: (label: string) => void } = $props();
</script>

<div class="thumbs">
  {#each images as image (image.label)}
    <figure title={image.path}>
      <img src={image.url} alt={image.label} />
      <button onclick={() => remove(image.label)} aria-label="Remove {image.label}">✕</button>
    </figure>
  {/each}
</div>

<style>
  .thumbs { display: flex; gap: 8px; flex-wrap: wrap; }
  figure { position: relative; margin: 0; width: 64px; height: 64px; border: 1px solid var(--raised); border-radius: var(--radius); overflow: hidden; background: var(--void); }
  img { width: 100%; height: 100%; object-fit: cover; display: block; }
  button {
    position: absolute; top: 2px; right: 2px; width: 18px; height: 18px; padding: 0; border: 0; border-radius: 2px;
    background: rgb(0 0 0 / 0.7); color: var(--text-strong); font-size: 10px; cursor: pointer;
  }
  button:hover { background: var(--error); }
</style>
