<script lang="ts">
  import { createStore } from "./stores/localStorage";
  import Game from "./scenes/game/Game.svelte";
  import Nickname from "./scenes/nickname/Nickname.svelte";
  import Join from "./scenes/join/Join.svelte";

  const nickname = createStore("nickname");
  let host = "";
  let opponent = "";

  let currentScene = $nickname === "" ? "nickname" : "join";

  function saveNickname(e: CustomEvent<string>) {
    nickname.set(e.detail);
    currentScene = "join";
  }

  function startGame(e: CustomEvent<{ host: string; opponent: string }>) {
    host = e.detail.host;
    opponent = e.detail.opponent;
    currentScene = "game";
  }
</script>

{#if currentScene === 'nickname'}
  <Nickname value={$nickname} on:submit={saveNickname} />
{:else if currentScene === 'join'}
  <Join nickname={$nickname} on:submit={startGame} />
{:else if currentScene === 'game'}
  <Game
    nickname={$nickname}
    {host}
    {opponent}
    on:changeNickname={() => (currentScene = "nickname")} />
{/if}
