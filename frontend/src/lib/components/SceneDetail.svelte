<script lang="ts">
  import { onDestroy } from "svelte";
  import { App } from "../../../bindings/case/internal/app";
  import type { SceneDetailDTO } from "../../../bindings/case/internal/app/models";
  import { navigate } from "../router.svelte";
  import { formatDuration, formatBitrate } from "../format";
  import { openExternal } from "../external";

  let { id }: { id: number } = $props();

  let scene = $state<SceneDetailDTO | null>(null);
  let loading = $state(false);
  let error = $state("");
  // Monotonic request id: stale responses and destroyed components are ignored.
  let requestId = 0;

  let hue = $derived(scene ? (scene.id * 47) % 360 : 0);
  let initial = $derived(
    (scene?.title?.trim() || scene?.code?.trim() || "?").charAt(0).toUpperCase(),
  );

  async function load(sceneId: number) {
    const req = ++requestId;
    loading = true;
    error = "";
    scene = null;

    try {
      const data = await App.GetScene(sceneId);
      if (req !== requestId) return;
      if (!data) {
        error = "场景不存在";
        return;
      }
      scene = data;
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
</script>

<main>
  <button class="back" onclick={() => navigate("/")}>← 返回列表</button>

  {#if loading}
    <p class="state">加载中...</p>
  {:else if error}
    <p class="error">{error}</p>
  {:else if scene}
    <div class="detail-top">
      <div class="cover">
        {#if scene.coverUrl}
          <img src={scene.coverUrl} alt={scene.title || scene.code || "cover"} />
        {:else}
          <div class="cover-placeholder" style={`--hue: ${hue}`}>{initial}</div>
        {/if}
      </div>

      <div class="head">
        <header>
          <h1>{scene.title || scene.code || "(无标题)"}</h1>
          <div class="badges">
            {#if scene.organized}<span class="badge">已整理</span>{/if}
            {#if scene.rating > 0}
              <span class="badge">评分 {scene.rating}/100</span>
            {:else}
              <span class="badge muted">未评分</span>
            {/if}
            {#if scene.resolution}<span class="badge muted">{scene.resolution}</span>{/if}
            {#if scene.duration > 0}<span class="badge muted">{formatDuration(scene.duration)}</span>{/if}
          </div>
        </header>

        <section class="meta">
          <div><span class="k">番号</span><span>{scene.code || "-"}</span></div>
          <div><span class="k">工作室</span><span>{scene.studioName || "-"}</span></div>
          <div><span class="k">导演</span><span>{scene.director || "-"}</span></div>
          <div><span class="k">发行日期</span><span>{scene.date || "-"}</span></div>
          <div><span class="k">制作日期</span><span>{scene.productionDate || "-"}</span></div>
          <div><span class="k">创建</span><span>{scene.createdAt || "-"}</span></div>
          <div><span class="k">更新</span><span>{scene.updatedAt || "-"}</span></div>
        </section>
      </div>
    </div>

    {#if (scene.performers ?? []).length > 0}
      <section>
        <h2>演员</h2>
        <div class="chips">
          {#each scene.performers ?? [] as performer (performer.id)}
            <span class="chip">{performer.name}</span>
          {/each}
        </div>
      </section>
    {/if}

    {#if (scene.tags ?? []).length > 0}
      <section>
        <h2>标签</h2>
        <div class="chips">
          {#each scene.tags ?? [] as tag (tag.id)}
            <span class="chip">{tag.name}</span>
          {/each}
        </div>
      </section>
    {/if}

    {#if scene.details}
      <section>
        <h2>简介</h2>
        <p class="details">{scene.details}</p>
      </section>
    {/if}

    {#if (scene.attachments ?? []).length > 0}
      <section>
        <h2>附件</h2>
        <div class="attachments">
          {#each scene.attachments ?? [] as attachment}
            <figure>
              <img src={attachment.url} alt={attachment.name} loading="lazy" />
              <figcaption>{attachment.name}</figcaption>
            </figure>
          {/each}
        </div>
      </section>
    {/if}

    {#if (scene.urls ?? []).length > 0}
      <section>
        <h2>链接</h2>
        <ul class="urls">
          {#each scene.urls ?? [] as url}
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

    {#if (scene.files ?? []).length > 0}
      <section>
        <h2>文件</h2>
        <table>
          <thead>
            <tr>
              <th>文件</th>
              <th>时长</th>
              <th>分辨率</th>
              <th>视频</th>
              <th>音频</th>
              <th>帧率</th>
              <th>码率</th>
              <th>格式</th>
            </tr>
          </thead>
          <tbody>
            {#each scene.files ?? [] as file (file.id)}
              <tr>
                <td class="path" title={file.path}>{file.basename || file.path}</td>
                <td>{formatDuration(file.duration)}</td>
                <td>{file.width > 0 && file.height > 0 ? `${file.width}×${file.height}` : "-"}</td>
                <td>{file.videoCodec || "-"}</td>
                <td>{file.audioCodec || "-"}</td>
                <td>{file.frameRate > 0 ? file.frameRate.toFixed(2) : "-"}</td>
                <td>{formatBitrate(file.bitrate)}</td>
                <td>{file.format || "-"}</td>
              </tr>
            {/each}
          </tbody>
        </table>
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
  .cover {
    flex: 0 0 220px;
  }
  .cover img,
  .cover-placeholder {
    width: 100%;
    aspect-ratio: 2 / 3;
    border-radius: 8px;
    display: block;
    object-fit: cover;
    background: #1c1c1c;
  }
  .cover-placeholder {
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
  header h1 {
    margin: 0 0 0.75rem;
    font-size: 1.6rem;
  }
  .badges {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
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
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 0.5rem 1.5rem;
    margin-bottom: 1.5rem;
  }
  .meta > div {
    display: flex;
    gap: 0.5rem;
    font-size: 0.9rem;
  }
  .meta .k {
    color: #888;
    min-width: 4.5rem;
  }
  section {
    margin-bottom: 1.5rem;
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
  .attachments {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 0.75rem;
  }
  .attachments figure {
    margin: 0;
  }
  .attachments img {
    width: 100%;
    aspect-ratio: 16 / 9;
    object-fit: cover;
    border-radius: 6px;
    display: block;
    background: #1c1c1c;
  }
  .attachments figcaption {
    font-size: 0.75rem;
    color: #888;
    margin-top: 0.25rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .urls a {
    color: #1a5fb4;
    word-break: break-all;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }
  th,
  td {
    text-align: left;
    padding: 0.4rem 0.6rem;
    border-bottom: 1px solid #eee;
    white-space: nowrap;
  }
  td.path {
    max-width: 260px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: monospace;
  }
</style>
