<script lang="ts">
  import { onMount } from "svelte";
  import SceneList from "./lib/components/SceneList.svelte";
  import SceneDetail from "./lib/components/SceneDetail.svelte";
  import { initRouter, route } from "./lib/router.svelte";

  onMount(() => initRouter());
</script>

{#if route.name === "detail"}
  <SceneDetail id={route.id} />
{:else}
  <SceneList />
{/if}
