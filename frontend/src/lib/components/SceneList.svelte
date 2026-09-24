<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { Events } from "@wailsio/runtime";
  import { App } from "../../../bindings/case/internal/app";
  import { navigate } from "../router.svelte";
  import { SCAN_COMPLETE } from "../events";

  let scanPath = $state("");
  let scanResult = $state("");
  let scanning = $state(false);

  let scenes: any[] = $state([]);
  let scenesLoading = $state(false);
  let scenesError = $state("");

  async function doScan() {
    if (!scanPath.trim()) {
      scanResult = "请输入文件夹路径";
      return;
    }
    scanning = true;
    scanResult = "";
    try {
      const jobID = await App.ScanLibrary(scanPath.trim());
      scanResult = `✅ 扫描任务已启动，Job ID: ${jobID}`;
    } catch (e) {
      scanResult = `❌ 扫描失败: ${e}`;
    } finally {
      scanning = false;
    }
  }

  async function loadScenes() {
    scenesLoading = true;
    scenesError = "";
    try {
      scenes = (await App.FindScenes(1, 50)) ?? [];
    } catch (e) {
      scenesError = `加载失败: ${e}`;
    } finally {
      scenesLoading = false;
    }
  }

  function openScene(id: number) {
    navigate(`/scenes/${id}`);
  }

  let offScan: (() => void) | undefined;

  onMount(() => {
    loadScenes();
    // 后端扫描/清理完成事件（替代 setTimeout 轮询）
    offScan = Events.On(SCAN_COMPLETE, () => {
      scanning = false;
      scanResult = "✅ 扫描完成，已刷新列表";
      loadScenes();
    });
  });

  onDestroy(() => offScan?.());
</script>

<main>
  <h1>Case</h1>

  <div class="scan-section">
    <h2>扫描媒体库</h2>
    <div class="scan-row">
      <input
        type="text"
        bind:value={scanPath}
        placeholder="输入文件夹路径，例如 E:\test"
        disabled={scanning}
      />
      <button onclick={doScan} disabled={scanning}>
        {scanning ? "启动中..." : "开始扫描"}
      </button>
    </div>
    {#if scanResult}
      <p class="result">{scanResult}</p>
    {/if}
  </div>

  <div class="scenes-section">
    <div class="header">
      <h2>场景列表 ({scenes.length})</h2>
      <button onclick={loadScenes} disabled={scenesLoading}>
        {scenesLoading ? "加载中..." : "刷新"}
      </button>
    </div>

    {#if scenesError}
      <p class="error">{scenesError}</p>
    {/if}

    {#if scenes.length === 0 && !scenesLoading}
      <p class="empty">还没有场景，先扫描一个文件夹试试。</p>
    {/if}

    <div class="scene-grid">
      {#each scenes as scene (scene.id)}
        <div
          class="scene-card"
          role="button"
          tabindex="0"
          onclick={() => openScene(scene.id)}
          onkeydown={(e) => {
            if (e.key === "Enter" || e.key === " ") openScene(scene.id);
          }}
        >
          <div class="scene-title">{scene.title || "(无标题)"}</div>
          <div class="scene-meta">
            {#if scene.organized}
              <span class="badge">已整理</span>
            {/if}
            <span class="hash">{scene.oshash}</span>
          </div>
          <div class="scene-path" title={scene.path}>{scene.path}</div>
          <div class="scene-date">{scene.createdAt}</div>
        </div>
      {/each}
    </div>
  </div>
</main>

<style>
  main {
    padding: 2rem;
    font-family: system-ui, sans-serif;
    max-width: 1200px;
    margin: 0 auto;
  }
  .scan-section,
  .scenes-section {
    margin-bottom: 2rem;
  }
  .scan-row {
    display: flex;
    gap: 0.5rem;
  }
  .scan-row input {
    flex: 1;
    padding: 0.5rem;
    font-size: 1rem;
    border: 1px solid #ccc;
    border-radius: 4px;
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
  .result {
    margin-top: 0.5rem;
    font-family: monospace;
    padding: 0.5rem;
    background: #eee;
    border-radius: 4px;
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
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }
  .scene-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1rem;
  }
  .scene-card {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1rem;
    background: #fafafa;
    transition: transform 0.15s;
    cursor: pointer;
  }
  .scene-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  .scene-card:focus-visible {
    outline: 2px solid #202b33;
    outline-offset: 2px;
  }
  .scene-title {
    font-weight: bold;
    margin-bottom: 0.5rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .scene-meta {
    display: flex;
    gap: 0.75rem;
    font-size: 0.85rem;
    color: #555;
    margin-bottom: 0.5rem;
  }
  .scene-path {
    font-size: 0.75rem;
    font-family: monospace;
    color: #888;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 0.25rem;
  }
  .badge {
    background: #4caf50;
    color: white;
    padding: 0.1rem 0.4rem;
    border-radius: 3px;
    font-size: 0.75rem;
  }
  .hash {
    font-family: monospace;
    font-size: 0.8rem;
    color: #666;
  }
  .scene-date {
    font-size: 0.75rem;
    color: #999;
    margin-top: 0.25rem;
  }
</style>
