<script lang="ts">
  import { createEventDispatcher } from "svelte";

  // Current input value.
  export let value = "";

  export let maxlength = 10;

  // We render a custom terminal carrot instead of the HTML form carrot.
  let carotIndex = -1;

  // Eliminate illegal characters from the input value as the user types.
  $: value = value.replaceAll(/[^a-zA-Z0-9 ]/g, "").toLowerCase();

  $: displayText = value.padEnd(maxlength, "_");

  // HTML input event handlers:

  function blur() {
    carotIndex = -1;
  }

  function unSelect(e: { currentTarget: HTMLInputElement }) {
    const target = e.currentTarget;
    // Defer state change until after the event has been processed by the element.
    setTimeout(() => {
      target.selectionStart = target.selectionEnd;
    }, 0);
  }

  function focus(e: { currentTarget: HTMLInputElement }) {
    setCarotIndex(e.currentTarget);
  }

  const dispatch = createEventDispatcher();

  function keyDown(e: KeyboardEvent & { currentTarget: HTMLInputElement }) {
    // Filter out keys which can be used to blur.
    if (e.key !== "Tab" && e.key !== "Enter") {
      setCarotIndex(e.currentTarget);
    }

    // Emit an event for the enter key.
    if (e.key === "Enter") {
      dispatch("enter");
    }
  }

  function setCarotIndex(target: HTMLInputElement) {
    // Defer state change until after the event has been processed by the element.
    setTimeout(() => {
      carotIndex = Math.min(target.selectionEnd ?? 0, maxlength - 1);
    }, 0);
  }
</script>

<style lang="less">
  input {
    position: absolute;
    top: 0;
    left: 0;
    border: 0;
    outline: 0;
    padding: 0;
    background-color: transparent;
    width: 100%;
    cursor: text;

    /* Inputted text is rendered by a separate element. */
    color: transparent;

    /* Even though text is transparent,
    this ensures the carot is lined up with the display text. */
    font-size: inherit;
    font-family: inherit;

    &::selection {
      color: transparent;
    }
  }

  .wrapper {
    position: relative;
    &:not(:first-child) {
      margin-block-start: var(--default-margin);
    }
  }

  .character {
    white-space: pre;
    text-transform: uppercase;
  }

  .carot {
    animation: blink 500ms linear infinite alternate;

    @keyframes blink {
      0%,
      40% {
        background-color: var(--accent-color);
        color: var(--bg-color);
      }
      60%,
      100% {
        background-color: var(--bg-color);
        color: var(--fg-color);
      }
    }
  }
</style>

<div class="wrapper">
  {#each displayText as char, i}
    <span class="character" class:carot={i === carotIndex}>{char}</span>
  {/each}

  <input
    spellcheck={false}
    bind:value
    on:blur={blur}
    on:select={unSelect}
    on:focus={focus}
    on:keydown={keyDown}
    on:mousedown={focus}
    {maxlength} />
</div>
