<script lang="ts">
  import { act } from "../lib/actions";
  import type { Question } from "../lib/types";

  let { id, questions }: { id: string; questions: Question[] } = $props();
  let picks = $state<boolean[][]>([]);
  let texts = $state<string[]>([]);

  $effect(() => {
    picks = questions.map((q) => (q.options ?? []).map(() => false));
    texts = questions.map(() => "");
  });

  function pick(q: number, o: number) {
    if (!questions[q].multiple) picks[q] = picks[q].map(() => false);
    picks[q][o] = !picks[q][o];
  }

  function answer(q: number): string {
    if (texts[q]?.trim()) return texts[q].trim();
    const chosen = (questions[q].options ?? []).filter((_, o) => picks[q]?.[o]);
    if (!questions[q].multiple) return chosen[0] ?? "";
    return chosen.length ? "- " + chosen.join("\n- ") : "";
  }

  const ready = $derived(questions.every((_, q) => answer(q) !== ""));

  function submit(event: Event) {
    event.preventDefault();
    if (ready) act(id, "answer", { answers: questions.map((_, q) => answer(q)) });
  }
</script>

<form class="ask" onsubmit={submit}>
  <span class="label">[ the agent asks ]</span>
  {#each questions as question, q (q)}
    <fieldset>
      <legend>{question.question}</legend>
      {#each question.options ?? [] as option, o (o)}
        <button type="button" class="option" class:on={picks[q]?.[o]} onclick={() => pick(q, o)}>
          <span class="box">{question.multiple ? (picks[q]?.[o] ? "■" : "□") : picks[q]?.[o] ? "●" : "○"}</span>{option}
        </button>
      {/each}
      <input class="field" placeholder={question.options?.length ? "Or type your own answer" : "Your answer"} bind:value={texts[q]} />
    </fieldset>
  {/each}
  <div class="actions"><button class="btn primary" disabled={!ready}>Answer</button></div>
</form>

<style>
  .ask {
    background: var(--card); border: 1px solid var(--accent); box-shadow: var(--glow);
    padding: 12px 14px; display: grid; gap: 12px; max-height: 50vh; overflow-y: auto;
  }
  fieldset { border: 0; margin: 0; padding: 0; display: grid; gap: 6px; }
  legend { color: var(--text-strong); margin-bottom: 6px; padding: 0; }
  .option {
    display: flex; gap: 10px; text-align: left; background: transparent; border: 1px solid var(--raised);
    border-radius: var(--radius); padding: 6px 10px; cursor: pointer; color: var(--text-dim);
  }
  .option:hover { background: var(--raised); }
  .option.on { border-color: var(--accent); color: var(--text-strong); }
  .box { color: var(--accent); }
  .actions { display: flex; justify-content: flex-end; }
</style>
