<script lang="ts">
  import Button from "../../lib/Button.svelte";
  import Input from "../../lib/Input.svelte";
  import Text from "../../lib/Text.svelte";
  import { currentScene, nickname } from "../../stores/global";

  let value = $nickname;

  function submit() {
    if (value !== "") {
      nickname.set(value);
      currentScene.set("join");
    }
  }
</script>

<Text>Enter your name:</Text>
<Input bind:value on:submit={submit} />
<Button on:click={submit}>NEXT</Button>
