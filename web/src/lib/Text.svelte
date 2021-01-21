<script lang="ts">
  export let color = "var(--fg-color)";
  export let bold = false;
  export let alignStart = false;
  export let alignEnd = false;
  export let header = false;
</script>

<style lang="less">
  p {
    margin: 0;
    &:not(:first-child) {
      margin-block-start: var(--default-margin);
    }
    color: var(--color);
  }

  .header {
    &:not(:first-child) {
      margin-block-start: calc(2 * var(--default-margin));
    }
    text-decoration: underline;
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

<p
  class:header
  class:bold
  class:alignStart
  class:alignEnd
  style="--color: {color}">
  <slot />
</p>
