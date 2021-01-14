<script lang="ts">
  import { createEventDispatcher } from "svelte";

  import type { Cell } from "../../types/boardTypes";
  import Disk from "./Disk.svelte";

  export let data: Cell[][] = [];
  const dispatch = createEventDispatcher();
</script>

<style>
  table {
    margin-block-start: calc(2 * var(--default-margin));
    border-collapse: collapse;
  }

  td {
    border: 1px solid var(--fg-color);
    padding: 8px;
  }
</style>

<table>
  {#each data as row, x}
    <tr>
      {#each row as value, y}
        <td>
          <Disk {value} on:click={() => dispatch('clickCell', { x, y })} />
        </td>
      {/each}
    </tr>
  {/each}
</table>
