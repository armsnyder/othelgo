<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Button from "../../lib/Button.svelte";
  import Input from "../../lib/Input.svelte";
  import Text from "../../lib/Text.svelte";

  export let value = "";

  const dispatch = createEventDispatcher<{ submit: string }>();

  function submit() {
    if (value !== "") {
      dispatch("submit", value);
    }
  }
</script>

<Text>Enter your name:</Text>
<Input bind:value on:enter={submit} />
<Button on:click={submit}>NEXT</Button>
