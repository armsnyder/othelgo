<script lang="ts">
  import Game from "./scenes/game/Game.svelte";
  import Nickname from "./scenes/nickname/Nickname.svelte";
  import Join from "./scenes/join/Join.svelte";
  import { currentScene, nickname } from "./stores/global";

  $: if (!$nickname) {
    currentScene.set("nickname");
  }
</script>

{#if $currentScene === 'nickname'}
  <Nickname />
{:else if $currentScene === 'join'}
  <Join />
{:else if $currentScene === 'game'}
  <Game />
{/if}
