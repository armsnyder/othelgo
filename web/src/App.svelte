<script lang="ts">
  import { onMount } from "svelte";
  import type { Decorate } from "./messageTypes";
  import { createMessageReceiver, sendMessage } from "./websocket";

  onMount(() => sendMessage({ action: "hello", version: "0.0.0" }));

  const message = createMessageReceiver<Decorate>("decorate");
</script>

<p>{$message?.decoration ?? 'Waiting to be decorated...'}</p>
