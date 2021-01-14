<style lang="less">
  .wrapper {
    display: flex;
    flex: 1;
    flex-direction: column;
    max-width: 1000px;
    margin-inline: var(--default-margin);
  }

  .main {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    flex: 4;
  }

  .footer {
    flex: 1 0;
  }
</style>

<div class="wrapper">
  <div class="main">
    <slot />
  </div>
  <div class="footer" />
</div>
