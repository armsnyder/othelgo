<script lang="ts">
  import { onMount } from "svelte";
  import { message, sendMessage } from "./websocket";

  onMount(() => {
      setTimeout(() => {
          sendMessage({ action: "hello", version: "0.0.0" })
      }, 1000)
  });
</script>

<p>{$message?.decoration ?? 'Waiting to be decorated...'}</p>
