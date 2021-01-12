<script lang="ts">
  import { onMount } from "svelte";
  import Board from "./Board.svelte";
  import Alert from "./Alert.svelte";
  import type { Decorate, Error } from "./messageTypes";
  import { createMessageReceiver, sendMessage } from "./websocket";

  onMount(() => sendMessage({ action: "hello", version: "0.0.0" }));

  const decorate = createMessageReceiver<Decorate>({
    action: "decorate",
    decoration: "",
  });

  const error = createMessageReceiver<Error>({ action: "error", error: "" });
</script>

<style>
  :root {
    font-family: Arial, sans-serif;
  }
</style>

{#if $error.error}
  <Alert>Error from server: {$error.error}</Alert>
{/if}

<p>{$decorate.decoration || 'Waiting to be decorated...'}</p>

<Board />
