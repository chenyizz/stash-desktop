<script lang="ts">
  import { onDestroy } from "svelte";

  let {
    placeholder = "搜索...",
    debounceMs = 300,
    onchange,
  }: {
    placeholder?: string;
    debounceMs?: number;
    onchange: (query: string) => void;
  } = $props();

  let text = $state("");
  // IME（中文输入法）组合态：组合期间不触发搜索，组合结束后再防抖触发。
  let composing = $state(false);
  let timer: ReturnType<typeof setTimeout> | undefined;

  function schedule() {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => onchange(text.trim()), debounceMs);
  }

  function handleInput() {
    if (composing) return;
    schedule();
  }

  function handleCompositionStart() {
    composing = true;
  }

  function handleCompositionEnd() {
    composing = false;
    schedule();
  }

  function clear() {
    text = "";
    if (timer) clearTimeout(timer);
    onchange("");
  }

  onDestroy(() => {
    if (timer) clearTimeout(timer);
  });
</script>

<div class="search-box">
  <input
    type="text"
    bind:value={text}
    {placeholder}
    oninput={handleInput}
    oncompositionstart={handleCompositionStart}
    oncompositionend={handleCompositionEnd}
  />
  {#if text}
    <button class="clear" onclick={clear} aria-label="清空搜索">×</button>
  {/if}
</div>

<style>
  .search-box {
    position: relative;
    display: inline-flex;
    align-items: center;
  }
  .search-box input {
    padding: 0.5rem 2rem 0.5rem 0.75rem;
    font-size: 0.95rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    min-width: 220px;
  }
  .clear {
    position: absolute;
    right: 0.35rem;
    border: none;
    background: transparent;
    color: #888;
    font-size: 1.1rem;
    line-height: 1;
    cursor: pointer;
    padding: 0 0.25rem;
  }
  .clear:hover {
    color: #333;
  }
</style>
