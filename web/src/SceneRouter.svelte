<script lang="ts">
  import { createStore } from "./stores/localStorage";
  import Game from "./scenes/game/Game.svelte";
  import Nickname from "./scenes/nickname/Nickname.svelte";

  const nickname = createStore("nickname");
  let pickingNickname = $nickname === "";

  function changeNickname() {
    pickingNickname = true;
  }

  function saveNickname(e: CustomEvent<string>) {
    nickname.set(e.detail);
    pickingNickname = false;
  }
</script>

{#if pickingNickname}
  <Nickname value={$nickname} on:submit={saveNickname} />
{:else}
  <Game nickname={$nickname} on:changeNickname={changeNickname} />
{/if}
