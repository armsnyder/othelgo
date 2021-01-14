<script lang="ts">
  import type { Cell } from "../../types/boardTypes";
  import Disk from "./Disk.svelte";

  export let data: Cell[][] = [];
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
  {#each data as row}
    <tr>
      {#each row as value}
        <td>
          <Disk {value} />
        </td>
      {/each}
    </tr>
  {/each}
</table>
