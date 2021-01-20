<script lang="ts">
  import { onMount, createEventDispatcher } from "svelte";
  import type { OpenGames } from "../../types/messageTypes";
  import { createMessageReceiver, sendMessage } from "../../stores/websocket";

  import Text from "../../lib/Text.svelte";
  import Button from "../../lib/Button.svelte";

  export let nickname: string;

  onMount(() => sendMessage({ action: "listOpenGames" }));

  const openGamesMessageStore = createMessageReceiver<OpenGames>({
    action: "openGames",
    hosts: [],
  });

  const dispatch = createEventDispatcher();

  function handleClickClosure(host: string) {
    function handleClick() {
      sendMessage({
        action: "joinGame",
        nickname,
        host,
      });
      dispatch("submit", {host, opponent: host});
    }
    return handleClick;
  }

  function handleHostGame() {
    sendMessage({
      action: "hostGame",
      nickname,
    });
    dispatch("submit", {host:nickname, opponent: ""})
  }
</script>

<Button on:click={handleHostGame}>HOST GAME</Button>

<Text>OPEN GAMES</Text>
{#each $openGamesMessageStore.hosts as host}
  <Button on:click={handleClickClosure(host)}>{host.toUpperCase()}</Button>
{/each}
