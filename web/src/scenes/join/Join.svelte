<script lang="ts">
  import { onMount } from "svelte";
  import type { OpenGames } from "../../types/messageTypes";
  import { createMessageReceiver, sendMessage } from "../../stores/websocket";
  import { host, currentScene, nickname } from "../../stores/global";

  import Text from "../../lib/Text.svelte";
  import Button from "../../lib/Button.svelte";

  onMount(() => sendMessage({ action: "listOpenGames" }));

  const openGames = createMessageReceiver<OpenGames>({
    action: "openGames",
    hosts: [],
  });

  function handleHost() {
    sendMessage({
      action: "hostGame",
      nickname: $nickname,
    });

    host.set($nickname);
    currentScene.set("game");
  }

  function handleJoin(selectedHost: string) {
    return () => {
      sendMessage({
        action: "joinGame",
        nickname: $nickname,
        host: selectedHost,
      });

      host.set(selectedHost);
      currentScene.set("game");
    };
  }
</script>

<Text alignEnd>Did you know? Your name is {$nickname.toUpperCase()}!</Text>
<Button alignEnd on:click={() => currentScene.set('nickname')}>
  CHANGE NICKNAME
</Button>

<Button on:click={handleHost}>HOST GAME</Button>

<Text header>OPEN GAMES</Text>
{#each $openGames.hosts as host}
  <Button on:click={handleJoin(host)}>{host.toUpperCase()}</Button>
{/each}
