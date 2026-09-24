<script lang="ts">
  import type { ScenesFilter } from "../../../bindings/case/internal/app/models";
  import { App } from "../../../bindings/case/internal/app";
  import EntityPicker from "./EntityPicker.svelte";

  type Option = { id: number; name: string };

  let { onapply }: { onapply: (filter: ScenesFilter) => void } = $props();

  let open = $state(false);
  let organized = $state(""); // "" | "yes" | "no"
  let ratingMin = $state("");
  let dateFrom = $state("");
  let dateTo = $state("");

  let tagSel = $state<Option[]>([]);
  let performerSel = $state<Option[]>([]);
  let studioSel = $state<Option[]>([]);
  let tagAll = $state(false);
  let includeSubTags = $state(false);
  let performerAll = $state(true);
  let includeSubStudios = $state(false);

  let activeCount = $derived(
    (organized ? 1 : 0) +
      (ratingMin !== "" ? 1 : 0) +
      (dateFrom || dateTo ? 1 : 0) +
      tagSel.length +
      performerSel.length +
      studioSel.length,
  );

  async function loadTags(q: string): Promise<Option[]> {
    const r = await App.FindTags({ query: q, page: 1, pageSize: 20 });
    return (r?.tags ?? []).map((t) => ({ id: t.id, name: t.name }));
  }

  async function loadPerformers(q: string): Promise<Option[]> {
    const r = await App.FindPerformers({ query: q, page: 1, pageSize: 20 });
    return (r?.performers ?? []).map((p) => ({ id: p.id, name: p.name }));
  }

  async function loadStudios(q: string): Promise<Option[]> {
    const r = await App.FindStudios({ query: q, page: 1, pageSize: 20 });
    return (r?.studios ?? []).map((s) => ({ id: s.id, name: s.name }));
  }

  function build(): ScenesFilter {
    return {
      organized: organized === "" ? null : organized === "yes",
      ratingMin: ratingMin === "" ? null : Number(ratingMin),
      dateFrom,
      dateTo,
      tagIds: tagSel.map((s) => s.id),
      tagAll,
      includeSubTags,
      performerIds: performerSel.map((s) => s.id),
      performerAll,
      studioId: studioSel.length > 0 ? studioSel[0].id : null,
      includeSubStudios,
    };
  }

  function apply() {
    onapply(build());
    open = false;
  }

  function reset() {
    organized = "";
    ratingMin = "";
    dateFrom = "";
    dateTo = "";
    tagSel = [];
    performerSel = [];
    studioSel = [];
    tagAll = false;
    includeSubTags = false;
    performerAll = true;
    includeSubStudios = false;
    onapply(build());
  }
</script>

<div class="filter">
  <button class="toggle" onclick={() => (open = !open)}>
    过滤{activeCount > 0 ? `（${activeCount}）` : ""}
  </button>

  {#if open}
    <div class="panel">
      <label>
        <span>已整理</span>
        <select bind:value={organized}>
          <option value="">全部</option>
          <option value="yes">是</option>
          <option value="no">否</option>
        </select>
      </label>

      <label>
        <span>最低评分</span>
        <input type="number" min="0" max="100" bind:value={ratingMin} placeholder="0-100" />
      </label>

      <label>
        <span>日期从</span>
        <input type="date" bind:value={dateFrom} />
      </label>

      <label>
        <span>日期到</span>
        <input type="date" bind:value={dateTo} />
      </label>

      <div class="entity-block">
        <span class="block-title">标签</span>
        <EntityPicker bind:selected={tagSel} placeholder="搜索标签" load={loadTags} />
        <label class="inline"><input type="checkbox" bind:checked={tagAll} /> 全部命中</label>
        <label class="inline"><input type="checkbox" bind:checked={includeSubTags} /> 含子标签</label>
      </div>

      <div class="entity-block">
        <span class="block-title">演员</span>
        <EntityPicker bind:selected={performerSel} placeholder="搜索演员" load={loadPerformers} />
        <label class="inline"><input type="checkbox" bind:checked={performerAll} /> 全部命中</label>
      </div>

      <div class="entity-block">
        <span class="block-title">工作室</span>
        <EntityPicker bind:selected={studioSel} placeholder="搜索工作室" load={loadStudios} />
        <label class="inline"><input type="checkbox" bind:checked={includeSubStudios} /> 含子工作室</label>
      </div>

      <div class="actions">
        <button onclick={reset}>重置</button>
        <button class="primary" onclick={apply}>应用</button>
      </div>
    </div>
  {/if}
</div>

<style>
  .filter {
    position: relative;
  }
  .toggle {
    padding: 0.5rem 1rem;
    font-size: 1rem;
    background: #202b33;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
  }
  .panel {
    position: absolute;
    right: 0;
    top: calc(100% + 0.4rem);
    z-index: 10;
    background: white;
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1rem;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    min-width: 240px;
  }
  label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    font-size: 0.9rem;
  }
  label span {
    color: #555;
  }
  input,
  select {
    padding: 0.35rem 0.5rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-size: 0.9rem;
    width: 140px;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 0.25rem;
  }
  .entity-block {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    border-top: 1px solid #eee;
    padding-top: 0.6rem;
  }
  .block-title {
    font-size: 0.85rem;
    color: #555;
  }
  .inline {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.8rem;
    color: #666;
    justify-content: flex-start;
  }
  .inline input {
    width: auto;
  }
  .actions button {
    padding: 0.4rem 0.9rem;
    border: 1px solid #ccc;
    background: #fff;
    border-radius: 4px;
    cursor: pointer;
  }
  .actions .primary {
    background: #202b33;
    color: #fff;
    border-color: #202b33;
  }
</style>
