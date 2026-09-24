<script lang="ts">
  import { onMount } from "svelte";
  import SceneList from "./lib/components/SceneList.svelte";
  import SceneDetail from "./lib/components/SceneDetail.svelte";
  import { initRouter, isActive, NAV_ITEMS, route } from "./lib/router.svelte";

  onMount(() => initRouter());
</script>

<nav class="top-nav">
  {#each NAV_ITEMS as item (item.name)}
    <a href={`#${item.path}`} class:active={isActive(item.name)}>{item.label}</a>
  {/each}
</nav>

{#if route.name === "scene"}
  <SceneDetail id={route.id} />
{:else if route.name === "performers" || route.name === "tags" || route.name === "studios"}
  <main class="placeholder">
    <p>{route.name} 开发中...</p>
  </main>
{:else}
  <SceneList />
{/if}

<style>
  .top-nav {
    display: flex;
    gap: 1rem;
    padding: 0.75rem 2rem;
    border-bottom: 1px solid #eee;
    font-family: system-ui, sans-serif;
  }
  .top-nav a {
    color: #555;
    text-decoration: none;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
  }
  .top-nav a.active {
    color: #202b33;
    font-weight: 600;
    background: #eef1f4;
  }
  .placeholder {
    padding: 2rem;
    font-family: system-ui, sans-serif;
    color: #888;
  }
</style>
