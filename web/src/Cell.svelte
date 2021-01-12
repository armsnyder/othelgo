<script lang="ts">
  import type { Disk } from "./boardTypes";
  export let disk: Disk = 0;
</script>

<style>
    .player-0 {
        color: rgba(0, 0, 0, 0);
        cursor: pointer;
    }

    .player-1 {
        color: magenta;
    }

    .player-2 {
        color: green;
    }
</style>

<div class="player-{disk}">⬤</div>
