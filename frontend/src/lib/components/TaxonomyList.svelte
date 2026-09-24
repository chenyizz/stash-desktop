<script lang="ts">
  import { onMount } from "svelte";
  import Pager from "../Pager.svelte";
  import SearchBox from "./SearchBox.svelte";

  type Item = { id: number; name: string };
  type PageResult = { items: Item[]; total: number; page: number };

  let {
    title,
    load,
  }: {
    title: string;
    load: (query: string, page: number, pageSize: number) => Promise<PageResult>;
  } = $props();

  const PAGE_SIZE = 60;

  let items = $state<Item[]>([]);
  let total = $state(0);
  let page = $state(1);
  let loading = $state(false);
  let error = $state("");
  let query = $state("");
  let requestId = 0;

  async function loadPage(target = page, q = query) {
    const req = ++requestId;
    loading = true;
    error = "";
    try {
      const result = await load(q, target, PAGE_SIZE);
      if (req !== requestId) return;
      items = result.items ?? [];
      total = result.total ?? 0;
      page = result.page ?? target;
    } catch (e) {
      if (req !== requestId) return;
      error = `加载失败: ${e}`;
    } finally {
      if (req === requestId) loading = false;
    }
  }

  function onSearch(q: string) {
    query = q;
    loadPage(1, q);
  }

  let totalPages = $derived(Math.max(1, Math.ceil(total / PAGE_SIZE)));

  function gotoPage(target: number) {
    const next = Math.min(Math.max(1, target), totalPages);
    if (next !== page) {
      loadPage(next, query);
    }
  }

  function avatarHue(id: number): number {
    return (id * 47) % 360;
  }

  function avatarInitial(name: string): string {
    return (name?.trim() || "?").charAt(0).toUpperCase();
  }

  onMount(() => loadPage(1));
</script>

<main>
  <div class="header">
    <h2>{title}（共 {total}）</h2>
    <div class="header-actions">
      <SearchBox placeholder={`搜索${title}`} onchange={onSearch} />
      <button onclick={() => loadPage(page, query)} disabled={loading}>
        {loading ? "加载中..." : "刷新"}
      </button>
    </div>
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if items.length === 0 && !loading}
    <p class="empty">还没有数据。</p>
  {/if}

  <div class="grid">
    {#each items as item (item.id)}
      <div class="card">
        <div class="avatar" style={`--hue: ${avatarHue(item.id)}`}>
          {avatarInitial(item.name)}
        </div>
        <div class="name" title={item.name}>{item.name || "(未命名)"}</div>
      </div>
    {/each}
  </div>

  <Pager {page} {totalPages} loading onGoto={gotoPage} />
</main>

<style>
  main {
    padding: 2rem;
    font-family: system-ui, sans-serif;
    max-width: 1200px;
    margin: 0 auto;
  }
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    gap: 0.5rem;
  }
  .header-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  button {
    padding: 0.5rem 1rem;
    font-size: 1rem;
    background: #202b33;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .error {
    color: #c00;
    padding: 0.5rem;
    background: #fee;
    border-radius: 4px;
  }
  .empty {
    color: #888;
    text-align: center;
    padding: 2rem;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 1rem;
  }
  .card {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 0.75rem;
    background: #fafafa;
    text-align: center;
  }
  .avatar {
    width: 100%;
    aspect-ratio: 1 / 1;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 1.6rem;
    font-weight: 700;
    background: hsl(var(--hue), 20%, 15%);
    user-select: none;
    margin-bottom: 0.5rem;
  }
  .name {
    font-size: 0.85rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
