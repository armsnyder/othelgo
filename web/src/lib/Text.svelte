<script lang="ts">
  export let color = "var(--fg-color)";
  export let bold = false;
  export let alignStart = false;
  export let alignEnd = false;
</script>

<style lang="less">
  p {
    margin: 0;
    &:not(:first-child) {
      margin-block-start: var(--default-margin);
    }
    color: var(--color);
  }

  .bold {
    font-weight: bold;
  }

  .alignEnd {
    align-self: flex-end;
  }

  .alignStart {
    align-self: flex-start;
  }
</style>

<p class:bold class:alignStart class:alignEnd style="--color: {color}">
  <slot />
</p>
