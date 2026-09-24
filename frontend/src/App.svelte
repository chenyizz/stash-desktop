<script lang="ts">
  import { onMount } from "svelte";
  import SceneList from "./lib/components/SceneList.svelte";
  import SceneDetail from "./lib/components/SceneDetail.svelte";
  import TaxonomyList from "./lib/components/TaxonomyList.svelte";
  import { App } from "../bindings/case/internal/app";
  import { initRouter, isActive, NAV_ITEMS, route } from "./lib/router.svelte";

  type Item = { id: number; name: string };
  type PageResult = { items: Item[]; total: number; page: number };

  async function loadPerformers(query: string, page: number, pageSize: number): Promise<PageResult> {
    const r = await App.FindPerformers({ query, page, pageSize });
    return {
      items: (r?.performers ?? []).map((p) => ({ id: p.id, name: p.name })),
      total: r?.total ?? 0,
      page: r?.page ?? page,
    };
  }

  async function loadTags(query: string, page: number, pageSize: number): Promise<PageResult> {
    const r = await App.FindTags({ query, page, pageSize });
    return {
      items: (r?.tags ?? []).map((t) => ({ id: t.id, name: t.name })),
      total: r?.total ?? 0,
      page: r?.page ?? page,
    };
  }

  async function loadStudios(query: string, page: number, pageSize: number): Promise<PageResult> {
    const r = await App.FindStudios({ query, page, pageSize });
    return {
      items: (r?.studios ?? []).map((s) => ({ id: s.id, name: s.name })),
      total: r?.total ?? 0,
      page: r?.page ?? page,
    };
  }

  onMount(() => initRouter());
</script>

<nav class="top-nav">
  {#each NAV_ITEMS as item (item.name)}
    <a href={`#${item.path}`} class:active={isActive(item.name)}>{item.label}</a>
  {/each}
</nav>

{#if route.name === "scene"}
  <SceneDetail id={route.id} />
{:else if route.name === "performers"}
  {#key "performers"}
    <TaxonomyList title="演员" load={loadPerformers} />
  {/key}
{:else if route.name === "tags"}
  {#key "tags"}
    <TaxonomyList title="标签" load={loadTags} />
  {/key}
{:else if route.name === "studios"}
  {#key "studios"}
    <TaxonomyList title="工作室" load={loadStudios} />
  {/key}
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
</style>
