<script>
  export let title = '';
  export let className = '';
  export let columns = [];
  export let rows = [];
  export let rowId = '';

  let pages = 0;
  let currentPage = 0;
  let sliceStart = 0;
  const sliceSize = 10;

  export let paging = {
    total: 0,
    size: 10,
    current: 0,
  };
  export let onChange = p => {
    currentPage = p;
  };

  $: {
    currentPage = paging.current;
    pages = Math.ceil((paging.total || rows.length) / paging.size);

    if (currentPage < sliceStart + 1) {
      sliceStart = Math.max(0, currentPage - sliceSize);
    } else if (currentPage > sliceStart + sliceSize) {
      sliceStart = Math.min(pages - 1, currentPage - 1);
    }
  }

  const onSelectPage = p => {
    if (p === currentPage) return;
    onChange(p);
  };
</script>

<div class={className}>
  <slot name="header"><h3>{title}</h3></slot>
  <div class="card">
    <div class:card-body={!!$$slots.filter}>
      <slot name="filter" />
      <div class="table-responsive">
        <table class="table table-vcenter card-table">
          <thead>
            <tr>
              {#each columns as c (c.key)}
                <th>{c.name}</th>
              {/each}
            </tr>
          </thead>
          <tbody>
            {#each rows as row, i (rowId ? row[rowId] : i)}
              <tr>
                {#each columns as c}
                  <td>{c.format ? c.format(row[c.key]) : row[c.key] || '-'}</td>
                {/each}
              </tr>
            {/each}
            {#if !rows.length}
              <tr>
                <td class="text-muted text-center" colspan={columns.length}>No data</td>
              </tr>
            {/if}
          </tbody>
        </table>
      </div>
    </div>
    <div class="card-footer d-flex align-items-center">
      <p class="m-0 text-muted">Total: {paging.total}</p>
      {#if pages > 1}
        <ul class="pagination m-0 ms-auto">
          {#if pages > 1}
            <li class="page-item" class:disabled={currentPage === 1}>
              <a
                class="page-link"
                href="#"
                on:click|preventDefault={() => onSelectPage(currentPage - 1)}>Prev</a
              >
            </li>
          {/if}
          {#if sliceStart > 0}
            <li class="page-item mx-2 disabled">...</li>
          {/if}
          {#each [...Array(sliceSize).keys()] as p}
            <li class="page-item" class:active={p + sliceStart + 1 === currentPage}>
              <a
                class="page-link"
                href="#"
                on:click|preventDefault={() => onSelectPage(p + sliceStart + 1)}
                >{p + sliceStart + 1}</a
              >
            </li>
          {/each}
          {#if sliceStart + sliceSize < pages - 1}
            <li class="page-item mx-2 disabled">...</li>
          {/if}
          <li class="page-item" class:disabled={currentPage >= pages}>
            <a
              class="page-link"
              href="#"
              on:click|preventDefault={() => onSelectPage(currentPage + 1)}
            >
              Next
            </a>
          </li>
        </ul>
      {/if}
    </div>
  </div>
</div>
