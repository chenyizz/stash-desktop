<script lang="ts">
  import { onMount } from "svelte";
  import { App } from "../bindings/case/internal/app";

  let version = $state("");
  let pong = $state("");
  let dataDir = $state("");

  onMount(async () => {
    version = await App.GetVersion();
    pong = await App.Ping();
    dataDir = await App.GetDataDir();
  });
</script>

<main>
  <h1>Case</h1>
  <p>Version: {version}</p>
  <p>Ping: {pong}</p>
  <p>DataDir: {dataDir}</p>
</main>
