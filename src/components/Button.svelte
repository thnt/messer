<script>
  /** @type {'primary' | 'secondary' | 'success' | 'warning' | 'danger' | 'info' | 'light' | 'dark'} */
  export let type = 'primary';
  export let ghost = false;
  export let outline = false;
  /** @type {'pill' | 'square'} */
  export let shape = undefined;
  export let className = '';
  export let disabled = false;
  /** @type {'lg' | 'sm'}*/
  export let size = undefined;
  export let submit = false;
  export let loading = false;

  let stylePrefix = 'btn-';
  if (ghost) {
    stylePrefix += 'ghost-';
  } else if (outline) {
    stylePrefix += 'outline-';
  }

  const classNames = ['btn', className, stylePrefix + type];
  if (shape) {
    classNames.push(`btn-${shape}`);
  }
  if (size) {
    classNames.push(`btn-${size}`);
  }
</script>

<button
  type={submit ? 'submit' : 'button'}
  class={classNames.join(' ').trim()}
  on:click
  disabled={disabled || loading}
>
  {#if loading}
    <span class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true" />
  {/if}
  <slot />
</button>
