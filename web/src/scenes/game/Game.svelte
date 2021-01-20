<script lang="ts">
  import { onDestroy, createEventDispatcher } from "svelte";
  import type { UpdateBoard } from "../../types/messageTypes";
  import { createMessageReceiver, sendMessage } from "../../stores/websocket";
  import Board from "./Board.svelte";
  import Text from "../../lib/Text.svelte";
  import Button from "../../lib/Button.svelte";

  export let nickname: string;
  export let host: string;
  export let opponent: string;

  onDestroy(() => sendMessage({ action: "leaveGame", nickname, host }));

  const boardUpdate = createMessageReceiver<UpdateBoard>({
    action: "updateBoard",
    board: [],
    player: 1,
    p1score: 0,
    p2score: 0,
    x: 0,
    y: 0,
  });

  const dispatch = createEventDispatcher();

  function handleClickCell(event: { detail: { x: number; y: number } }) {
    sendMessage({
      action: "placeDisk",
      nickname,
      host,
      x: event.detail.x,
      y: event.detail.y,
    });
  }
</script>

<Text alignEnd>Did you know? Your name is {nickname.toUpperCase()}!</Text>
<Text alignStart>Opponent: {opponent}</Text>
<Button alignEnd on:click={() => dispatch('changeNickname')}>
  CHANGE NICKNAME
</Button>
<Board data={$boardUpdate.board} on:clickCell={handleClickCell} />
