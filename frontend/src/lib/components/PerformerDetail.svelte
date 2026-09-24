<script lang="ts">
  import { onDestroy } from "svelte";
  import { App } from "../../../bindings/case/internal/app";
  import type { PerformerDetailDTO } from "../../../bindings/case/internal/app/models";
  import { navigate } from "../router.svelte";
  import { openExternal } from "../external";

  let { id }: { id: number } = $props();

  let performer = $state<PerformerDetailDTO | null>(null);
  let loading = $state(false);
  let error = $state("");
  let requestId = 0;

  async function load(performerId: number) {
    const req = ++requestId;
    loading = true;
    error = "";
    performer = null;
    try {
      const data = await App.GetPerformer(performerId);
      if (req !== requestId) return;
      if (!data) {
        error = "演员不存在";
        return;
      }
      performer = data;
    } catch (e) {
      if (req !== requestId) return;
      error = `加载失败: ${e}`;
    } finally {
      if (req === requestId) loading = false;
    }
  }

  $effect(() => {
    load(id);
  });

  onDestroy(() => {
    requestId++;
  });

  let hue = $derived(performer ? (performer.id * 47) % 360 : 0);
  let initial = $derived((performer?.name?.trim() || "?").charAt(0).toUpperCase());
</script>

<main>
  <button class="back" onclick={() => navigate("/performers")}>← 返回演员</button>

  {#if loading}
    <p class="state">加载中...</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if performer}
    <div class="detail-top">
      <div class="avatar">
        {#if performer.imageUrl}
          <img src={performer.imageUrl} alt={performer.name || "avatar"} />
        {:else}
          <div class="avatar-placeholder" style={`--hue: ${hue}`}>{initial}</div>
        {/if}
      </div>

      <div class="head">
        <h1>
          {performer.name || "(未命名)"}
          {#if performer.disambiguation}<span class="disambig">（{performer.disambiguation}）</span>{/if}
        </h1>
        <div class="badges">
          <span class="badge muted">#{performer.id}</span>
          {#if performer.favorite}<span class="badge">收藏</span>{/if}
          {#if performer.rating > 0}
            <span class="badge">评分 {performer.rating}/100</span>
          {:else}
            <span class="badge muted">未评分</span>
          {/if}
          {#if performer.gender}<span class="badge muted">{performer.gender}</span>{/if}
        </div>

        <section class="meta">
          <div><span class="k">生日</span><span>{performer.birthdate || "-"}</span></div>
          <div><span class="k">去世</span><span>{performer.deathDate || "-"}</span></div>
          <div><span class="k">国家</span><span>{performer.country || "-"}</span></div>
          <div><span class="k">身高</span><span>{performer.height ?? "-"}</span></div>
          <div><span class="k">体重</span><span>{performer.weight ?? "-"}</span></div>
          <div><span class="k">三围</span><span>{performer.measurements || "-"}</span></div>
          <div><span class="k">出道</span><span>{performer.careerStart || "-"}</span></div>
          <div><span class="k">退役</span><span>{performer.careerEnd || "-"}</span></div>
          <div><span class="k">纹身</span><span>{performer.tattoos || "-"}</span></div>
          <div><span class="k">创建</span><span>{performer.createdAt || "-"}</span></div>
          <div><span class="k">更新</span><span>{performer.updatedAt || "-"}</span></div>
        </section>
      </div>
    </div>

    {#if (performer.aliases ?? []).length > 0}
      <section>
        <h2>别名</h2>
        <div class="chips">
          {#each performer.aliases ?? [] as alias}
            <span class="chip">{alias}</span>
          {/each}
        </div>
      </section>
    {/if}

    {#if (performer.tags ?? []).length > 0}
      <section>
        <h2>标签</h2>
        <div class="chips">
          {#each performer.tags ?? [] as tag (tag.id)}
            <span class="chip">{tag.name}</span>
          {/each}
        </div>
      </section>
    {/if}

    {#if performer.details}
      <section>
        <h2>简介</h2>
        <p class="details">{performer.details}</p>
      </section>
    {/if}

    {#if (performer.urls ?? []).length > 0}
      <section>
        <h2>链接</h2>
        <ul class="urls">
          {#each performer.urls ?? [] as url}
            <li>
              <a
                href={url}
                onclick={(e) => {
                  e.preventDefault();
                  openExternal(url);
                }}>{url}</a
              >
            </li>
          {/each}
        </ul>
      </section>
    {/if}
  {/if}
</main>

<style>
  main {
    padding: 2rem;
    font-family: system-ui, sans-serif;
    max-width: 1000px;
    margin: 0 auto;
  }
  .back {
    padding: 0.4rem 0.9rem;
    font-size: 0.9rem;
    background: #202b33;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    margin-bottom: 1.25rem;
  }
  .detail-top {
    display: flex;
    gap: 1.5rem;
    align-items: flex-start;
  }
  .avatar {
    flex: 0 0 200px;
  }
  .avatar img,
  .avatar-placeholder {
    width: 100%;
    aspect-ratio: 1 / 1;
    border-radius: 8px;
    display: block;
    object-fit: cover;
    background: #1c1c1c;
  }
  .avatar-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 3rem;
    font-weight: 700;
    user-select: none;
    background: hsl(var(--hue), 20%, 15%);
  }
  .head {
    flex: 1;
    min-width: 0;
  }
  h1 {
    margin: 0 0 0.75rem;
    font-size: 1.6rem;
  }
  .disambig {
    font-size: 1rem;
    color: #888;
    font-weight: normal;
  }
  .badges {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 1.25rem;
  }
  .badge {
    background: #4caf50;
    color: white;
    padding: 0.15rem 0.5rem;
    border-radius: 3px;
    font-size: 0.8rem;
  }
  .badge.muted {
    background: #e0e0e0;
    color: #444;
  }
  .state {
    color: #666;
    padding: 2rem 0;
  }
  .error {
    color: #c00;
    padding: 0.75rem;
    background: #fee;
    border-radius: 4px;
  }
  .meta {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 0.5rem 1.5rem;
  }
  .meta > div {
    display: flex;
    gap: 0.5rem;
    font-size: 0.9rem;
  }
  .meta .k {
    color: #888;
    min-width: 4rem;
  }
  section {
    margin-top: 1.5rem;
  }
  section h2 {
    font-size: 1rem;
    margin: 0 0 0.5rem;
    color: #333;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }
  .chip {
    background: #eef1f4;
    border: 1px solid #dde3e8;
    border-radius: 999px;
    padding: 0.15rem 0.7rem;
    font-size: 0.85rem;
  }
  .details {
    margin: 0;
    line-height: 1.6;
    white-space: pre-wrap;
  }
  .urls {
    margin: 0;
    padding-left: 1.1rem;
  }
  .urls a {
    color: #1a5fb4;
    word-break: break-all;
  }
</style>
