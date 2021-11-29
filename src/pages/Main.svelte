<script>
  import dayjs from 'dayjs';

  import Table from '../components/Table.svelte';
  import Footer from '../components/Footer.svelte';
  import Header from '../components/Header.svelte';
  import Overview from '../components/Overview.svelte';
  import Button from '../components/Button.svelte';
  import MetricChart from '../components/MetricChart.svelte';
  import { metricStore } from '../store';

  const columns = [
    {
      key: 'Timestamp',
      name: 'Timestamp',
      format: v => dayjs.unix(v).format('D/M/YYYY H:mm:ss'),
    },
    { key: 'Pressure', name: 'Pressure (Barg)' },
    { key: 'Temperature', name: 'Temperature (°C)' },
    { key: 'Massflow', name: 'Flow (NCMH)' },
    { key: 'TTflowL1000', name: 'Normal (NCM)' },
    { key: 'TTflowG1000', name: 'Over max (NCM)' },
    { key: 'TotalFlow', name: 'TotalFlow (NCM)' },
  ];

  const pagesize = 15;
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
  let recents = [];
  let paging = {
    total: 0,
    size: pagesize,
    current: 1,
  };
  let latestTs;
  let deviceErrorCode;

  metricStore.subscribe(s => {
    rows = s.metrics;
    paging = { ...paging, total: s.total };
    recents = s.recents;
    latestTs = recents?.[0]?.Timestamp || rows?.[0]?.Timestamp;
    deviceErrorCode = recents?.[0]?.DeviceErrorPLC;
  });

  let now = dayjs();
  setInterval(() => {
    now = dayjs();
  }, 1000);


  getMetricPage(1);
</script>

{#if (latestTs && now.diff(dayjs.unix(latestTs), 'm') > 10) || deviceErrorCode < 0}
  <div class="alert alert-danger mb-0 bg-red text-white text-center rounded-0">
    <!-- Download SVG icon from http://tabler-icons.io/i/alert-circle -->
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="icon"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      stroke-width="2"
      stroke="currentColor"
      fill="none"
      stroke-linecap="round"
      stroke-linejoin="round"
      ><path stroke="none" d="M0 0h24v24H0z" fill="none" /><circle cx="12" cy="12" r="9" /><line
        x1="12"
        y1="8"
        x2="12"
        y2="12"
      /><line x1="12" y1="16" x2="12.01" y2="16" /></svg
    >
    <span
      >{deviceErrorCode < 0
        ? `Disconnected from PCL device`
        : `No new data since ${dayjs.unix(latestTs).format("D/M/YYYY H:mm:ss")}`}</span
    >
  </div>
{/if}
<Header />
<main class="container-lg pt-4">
  <h1 class="heading text-center mb-3 text-white">TRINA SOLAR REMOTE MONITORING</h1>
  <Overview />
  <h3 class="mt-4">Recents</h3>
  <MetricChart data={recents} />
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
