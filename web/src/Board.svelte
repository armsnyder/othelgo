<script lang="ts">
  import { onMount } from "svelte";

  import Cell from "./Cell.svelte";
  import type { UpdateBoard } from "./messageTypes";
  import { createMessageReceiver, sendMessage } from "./websocket";

  onMount(() => sendMessage({ action: "hostGame", nickname: "adam" }));

  const boardUpdate = createMessageReceiver<UpdateBoard>({
    action: "updateBoard",
    board: [],
    player: 1,
    p1score: 0,
    p2score: 0,
    x: 0,
    y: 0,
  });
</script>

<style>
  table {
    border: 1px solid black;
    user-select: none;
  }

  td {
    border: 1px solid black;
    height: 68px;
    width: 68px;
    text-align: center;
    vertical-align: middle;
    font-size: 42px;
    padding: 0;
  }
</style>

<table>
  {#each $boardUpdate.board as row}
    <tr>
      {#each row as value}
        <td>
          <Cell {value} />
        </td>
      {/each}
    </tr>
  {/each}
</table>
