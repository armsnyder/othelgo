<script lang="ts">
  import type { Cell } from "../../types/boardTypes";

  export let value: Cell = 0;
</script>

<style>
  div {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: 1px solid var(--fg-color);
  }

  .player-0 {
    cursor: pointer;
    border: 0;
  }

  .player-1 {
    background-color: magenta;
  }

  .player-2 {
    background-color: green;
  }
</style>

<!-- https://svelte.dev/tutorial/event-forwarding -->
<div class="player-{value}" on:click/> 
