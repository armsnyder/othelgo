<script lang="ts">
  export let alignStart = false;
  export let alignEnd = false;
</script>

<style lang="less">
  button {
    &:not(:first-child) {
      margin-block-start: var(--default-margin);
    }
    padding: 0;
    border: 0;
    outline: 0;
    font-family: inherit;
    font-size: inherit;
    cursor: pointer;
    color: inherit;
    background-color: inherit;

    &:hover {
      background-color: var(--accent-color);
      color: var(--bg-color);
    }

    &:active {
      background-color: var(--fg-color);
    }
  }

  .alignEnd {
    align-self: flex-end;
  }

  .alignStart {
    align-self: flex-start;
  }
</style>

<button class:alignStart class:alignEnd on:click>[
  <slot />
  ]</button>
