<script lang="ts">
  let { kind }: { kind: "working" | "tasks" } = $props();
</script>

<div class="glow {kind}" aria-hidden="true"></div>

<style>
  .glow {
    --c: var(--accent);
    position: absolute; left: 0; right: 0; bottom: 0; height: 240px; z-index: 1; pointer-events: none;
    background: radial-gradient(
      ellipse 70% 100% at 50% 100%,
      color-mix(in srgb, var(--c) 55%, transparent) 0%,
      color-mix(in srgb, var(--c) 18%, transparent) 45%,
      transparent 75%
    );
    animation: glow 2.4s ease-in-out infinite;
  }
  .glow::after {
    content: ""; position: absolute; left: 8%; right: 8%; bottom: 0; height: 1px;
    background: linear-gradient(90deg, transparent, var(--c), transparent);
  }
  .tasks { --c: var(--violet); }
  @keyframes glow { 0%, 100% { opacity: 0.3; } 50% { opacity: 1; } }
  @media (prefers-reduced-motion: reduce) { .glow { animation: none; opacity: 0.6; } }
</style>
