<script lang="ts">
  import type { Error } from "./types/messageTypes";
  import { onMount } from "svelte";
  import { createMessageReceiver, sendMessage } from "./stores/websocket";
  import Title from "./lib/Title.svelte";
  import Text from "./lib/Text.svelte";
  import CenterLayout from "./lib/CenterLayout.svelte";
  import SceneRouter from "./SceneRouter.svelte";

  onMount(() => sendMessage({ action: "hello", version: "0.0.0" }));

  const error = createMessageReceiver<Error>({ action: "error", error: "" });
</script>

<style lang="less">
  .app {
    display: flex;
    justify-content: center;
    height: 100%;
    font-size: 14pt;
    background: var(--bg-color);
    color: var(--fg-color);
    font-family: "Source Code Pro", monospace;
  }
</style>

<div class="app">
  <CenterLayout>
    <Title />

    {#if $error.error}
      <Text color="palevioletred" bold>
        Error from server:
        {$error.error}
      </Text>
    {/if}

    <SceneRouter />
  </CenterLayout>
</div>
