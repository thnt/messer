<script>
  import ApexCharts from 'apexcharts';
  import dayjs from 'dayjs';
  import { onMount } from 'svelte';

  export let title = 'Metric';
  export let data = [];
  export let dataKey = '';

  let chartNode, chart;

  const chartOpts = {
    chart: {
      type: 'area',
      height: 80,
      sparkline: {
        enabled: true,
      },
    },
    colors: ['#206bc4'],
    stroke: { width: 1.5 },
    tooltip: {
      x: { show: false },
      y: {
        title: {
          formatter: (_, s) =>
            dayjs(s.series[s.seriesIndex][s.dataPointIndex][0]).format('D/M/YYYY H:mm:ss') + ':',
        },
      },
      marker: { show: false },
    },
    xaxis: {
      type: 'datetime',
    },
    series: [
      {
        data: data
          .reverse()
          .map(d => [d.Timestamp*1000, d[dataKey]])
          .reverse(),
      },
    ],
  };

  let max, min;

  $: {
    if (chart) {
      const nextData = data.map(d => [d.Timestamp*1000, d[dataKey]]).reverse();
      const values = nextData.map(v => v[1]);
      max = Math.max(...values);
      min = Math.min(...values);
      chart.updateSeries([{ data: nextData }]);
    }
  }

  onMount(() => {
    chart = new ApexCharts(chartNode, chartOpts);
    chart.render();

    return () => chart.destroy();
  });
</script>

<div class="card">
  <div class="card-body d-flex mt-3 align-items-center">
    <h3 class="me-2 title">
      {title}
      {#if data.length}
        <div
          class="text-muted timestamp"
          title={dayjs.unix(data[0].Timestamp).format('D/M/YYYY H:mm:ss')}
        >
          {dayjs.unix(data[0].Timestamp).fromNow()}
        </div>
      {/if}
    </h3>
    <div class="values ms-auto">
      <div class="max text-danger">MAX: <span>{max || '-'}</span></div>
      <div class="value">{data?.[0]?.[dataKey] || 0}</div>
      <div class="min text-green">MIN: <span>{min || '-'}</span></div>
    </div>
  </div>
  <div bind:this={chartNode} />
</div>

<style>
  .title {
    font-size: 1.2rem;
  }
  .value {
    font-weight: 600;
    font-size: 2rem;
  }
  .timestamp {
    font-weight: normal;
    font-size: 0.8rem;
  }
  .max,
  .min {
    font-size: 0.9rem;
  }
  .max span,
  .min span {
    font-weight: 600;
  }
</style>
