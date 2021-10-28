<script>
  import dayjs from 'dayjs';

  import Table from '../components/Table.svelte';
  import Footer from '../components/Footer.svelte';
  import Header from '../components/Header.svelte';
  import Overview from '../components/Overview.svelte';
  import Button from '../components/Button.svelte';
  import { metricStore } from '../store';

  const columns = [
    { key: 'Timestamp', name: 'Timestamp', format: v => dayjs.unix(v).format('D/M/YYYY H:mm:ss') },
    { key: 'Pressure', name: 'Pressure (Barg)' },
    { key: 'Temperature', name: 'Temperature (°C)' },
    { key: 'Massflow', name: 'Flow (NCMH)' },
    { key: 'TTflowG1000', name: 'High (NCM)' },
    { key: 'TTflowL1000', name: 'Low (NCM)' },
  ];

  const pagesize = 10;
  let from, to;
  const getMetricPage = p => {
    paging.current = p;
    const params = {
      limit: pagesize,
      skip: (p - 1) * pagesize,
    };
    if (from) {
      params.from = dayjs(from).startOf('d').unix();
    }
    if (to) {
      params.to = dayjs(to).endOf('d').unix();
    }
    metricStore.getMetrics(params);
  };

  let rows = [];
  let paging = {
    total: 0,
    size: pagesize,
    current: 1,
  };

  metricStore.subscribe(s => {
    rows = s.metrics;
    paging = { ...paging, total: s.total };
  });

  getMetricPage(1);
</script>

<Header />
<main class="container-lg pt-4">
  <h1 class="heading text-center mb-3 text-white">PLC AND REMOTE MONITOR SYSTEM</h1>
  <Overview />

  <Table title="Detail" className="mt-4" {columns} {rows} {paging} onChange={getMetricPage}>
    <div slot="filter" class="filter row g-2 mb-3">
      <div class="col-sm-6 col-md-4 col-lg-3">
        <input class="form-control" type="date" bind:value={from} />
      </div>
      <div class="col-sm-6 col-md-4 col-lg-3">
        <input class="form-control" type="date" bind:value={to} />
      </div>
      <div class="col-sm-6 col-md-4 col-lg-3">
        <Button on:click={() => getMetricPage(1)}>OK</Button>
      </div>
    </div>
  </Table>
</main>
<Footer />

<style>
  .heading {
    font-size: calc(0.9rem + 0.6vw);
  }
  @media (min-width: 1400px) {
    .heading {
      font-size: 1.5rem;
    }
  }
</style>
