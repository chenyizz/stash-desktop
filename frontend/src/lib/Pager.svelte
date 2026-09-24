<script lang="ts">
  let {
    page,
    totalPages,
    loading = false,
    onGoto,
  }: {
    page: number;
    totalPages: number;
    loading?: boolean;
    onGoto: (target: number) => void;
  } = $props();
</script>

{#if totalPages > 1}
  <div class="pager">
    <button onclick={() => onGoto(page - 1)} disabled={page <= 1 || loading}>上一页</button>
    <span>第 {page} / {totalPages} 页</span>
    <button onclick={() => onGoto(page + 1)} disabled={page >= totalPages || loading}>下一页</button>
  </div>
{/if}

<style>
  .pager {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    margin-top: 1.5rem;
    font-size: 0.9rem;
    color: #555;
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
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
