<script lang="ts">
  import type { GameOver, Joined, UpdateBoard } from "../../types/messageTypes";
  import { createMessageReceiver, sendMessage } from "../../stores/websocket";
  import Board from "./Board.svelte";
  import Text from "../../lib/Text.svelte";
  import { host, nickname, currentScene } from "../../stores/global";
  import Button from "../../lib/Button.svelte";

  let isHost = $host === $nickname;

  let opponent = isHost ? "" : $host;

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

  function handleClickQuit() {
    sendMessage({ action: "leaveGame", nickname: $nickname, host: $host });
  }

  const gameOver = createMessageReceiver<GameOver>({
    action: "gameOver",
    message: "",
  });

  $: if ($gameOver.message) {
    currentScene.set("join");
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

<Button on:click={handleClickQuit} alignEnd>LEAVE GAME</Button>
