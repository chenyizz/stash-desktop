<script lang="ts">
  import type { ScenesFilter } from "../../../bindings/case/internal/app/models";

  let { onapply }: { onapply: (filter: ScenesFilter) => void } = $props();

  let open = $state(false);
  let organized = $state(""); // "" | "yes" | "no"
  let ratingMin = $state("");
  let dateFrom = $state("");
  let dateTo = $state("");

  let activeCount = $derived(
    (organized ? 1 : 0) + (ratingMin !== "" ? 1 : 0) + (dateFrom || dateTo ? 1 : 0),
  );

  function build(): ScenesFilter {
    return {
      organized: organized === "" ? null : organized === "yes",
      ratingMin: ratingMin === "" ? null : Number(ratingMin),
      dateFrom,
      dateTo,
      tagIds: [],
      tagAll: false,
      includeSubTags: false,
      performerIds: [],
      performerAll: true,
      studioId: null,
      includeSubStudios: false,
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
