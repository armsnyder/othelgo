<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import type { UpdateBoard } from "../../types/messageTypes";
  import { createMessageReceiver, sendMessage } from "../../stores/websocket";
  import Board from "./Board.svelte";
  import Text from "../../lib/Text.svelte";
  import Button from "../../lib/Button.svelte";
  import Uppercase from "../../lib/Uppercase.svelte";

  export let nickname: string;

  onMount(() => sendMessage({ action: "hostGame", nickname }));

  onDestroy(() =>
    sendMessage({ action: "leaveGame", nickname, host: nickname })
  );

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

  function changeNickname() {
    dispatch("changeNickname");
  }
</script>

<Text alignEnd>
  Did you know? Your name is
  <Uppercase>{nickname}</Uppercase>!
</Text>
<Button alignEnd on:click={changeNickname}>CHANGE NICKNAME</Button>
<Board data={$boardUpdate.board} />
