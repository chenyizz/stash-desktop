<script lang="ts">
  import SearchBox from "./SearchBox.svelte";

  type Option = { id: number; name: string };

  let {
    placeholder = "搜索...",
    load,
    selected = $bindable<Option[]>([]),
  }: {
    placeholder?: string;
    load: (query: string) => Promise<Option[]>;
    selected?: Option[];
  } = $props();

  let results = $state<Option[]>([]);
  let loading = $state(false);

  async function search(q: string) {
    if (!q.trim()) {
      results = [];
      return;
    }
    loading = true;
    try {
      results = (await load(q)).slice(0, 20);
    } catch {
      results = [];
    } finally {
      loading = false;
    }
  }

  function add(option: Option) {
    if (!selected.some((s) => s.id === option.id)) {
      selected = [...selected, option];
    }
    results = [];
  }

  function remove(id: number) {
    selected = selected.filter((s) => s.id !== id);
  }
</script>

<div class="entity-picker">
  {#if selected.length > 0}
    <div class="chips">
      {#each selected as item (item.id)}
        <span class="chip">
          {item.name}
          <button onclick={() => remove(item.id)} aria-label="移除">×</button>
        </span>
      {/each}
    </div>
  {/if}

  <SearchBox {placeholder} onchange={search} />

  {#if loading}
    <p class="hint">搜索中...</p>
  {/if}

  {#if results.length > 0}
    <ul class="results">
      {#each results as option (option.id)}
        <li>
          <button onclick={() => add(option)}>{option.name}</button>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .entity-picker {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    background: #eef1f4;
    border: 1px solid #dde3e8;
    border-radius: 999px;
    padding: 0.05rem 0.5rem;
    font-size: 0.8rem;
  }
  .chip button {
    border: none;
    background: transparent;
    cursor: pointer;
    color: #888;
    font-size: 0.9rem;
    line-height: 1;
  }
  .hint {
    margin: 0;
    font-size: 0.8rem;
    color: #888;
  }
  .results {
    list-style: none;
    margin: 0;
    padding: 0;
    max-height: 180px;
    overflow: auto;
    border: 1px solid #ddd;
    border-radius: 4px;
  }
  .results li button {
    width: 100%;
    text-align: left;
    border: none;
    background: transparent;
    padding: 0.35rem 0.6rem;
    cursor: pointer;
    font-size: 0.85rem;
  }
  .results li button:hover {
    background: #f2f5f8;
  }
</style>
