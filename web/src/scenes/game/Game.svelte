<script lang="ts">
  import { onDestroy } from "svelte";
  import type { Joined, UpdateBoard } from "../../types/messageTypes";
  import { createMessageReceiver, sendMessage } from "../../stores/websocket";
  import Board from "./Board.svelte";
  import Text from "../../lib/Text.svelte";
  import { host, nickname } from "../../stores/global";

  let isHost = $host === $nickname;

  export let opponent = isHost ? "" : $host;

  onDestroy(() =>
    sendMessage({ action: "leaveGame", nickname: $nickname, host: $host })
  );

  const joined = createMessageReceiver<Joined>({
    action: "joined",
    nickname: "",
  });

  $: opponent ||= $joined.nickname;

  const boardUpdate = createMessageReceiver<UpdateBoard>({
    action: "updateBoard",
    board: [],
    player: 1,
    p1score: 0,
    p2score: 0,
    x: 0,
    y: 0,
  });

  $: yourScore = isHost ? $boardUpdate.p1score : $boardUpdate.p2score;
  $: opponentScore = isHost ? $boardUpdate.p2score : $boardUpdate.p1score;

  function handleClickCell(event: CustomEvent<{ x: number; y: number }>) {
    sendMessage({
      action: "placeDisk",
      nickname: $nickname,
      host: $host,
      x: event.detail.x,
      y: event.detail.y,
    });
  }
</script>

{#if !opponent}
  <Text color="lightgreen">Waiting for opponent to join...</Text>
{/if}

<Text alignStart>{$nickname.toUpperCase()}: {yourScore}</Text>

<Text alignStart>
  {opponent.toUpperCase() || '[OPPONENT]'}:
  {opponentScore}
</Text>

<Board data={$boardUpdate.board} on:clickCell={handleClickCell} />
