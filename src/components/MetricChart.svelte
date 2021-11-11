<script>
  import ApexCharts from 'apexcharts';
  import { onMount } from 'svelte';

  export let data = [];

  let chartNode, chart;
  const metrics = [
    'TotalFlow',
    'Massflow',
    'Pressure',
    // 'TTflowG1000',
    // 'TTflowL1000',
    'Temperature',
  ];
  const colors = ['#206bc4', '#2fb344', '#ae3ec9', '#f76707', '#d63939', '#0ca678'];

  let selectedMetrics = {
    Massflow: true,
  };

  $: {
    if (chart) {
      chart.updateSeries(
        metrics
          .filter(m => selectedMetrics[m])
          .map(m => ({
            name: m,
            data: data.map(d => [d.Timestamp * 1000, d[m]]).reverse(),
          })),
      );
    }
  }

  onMount(() => {
    const chartOpts = {
      chart: {
        type: 'line',
        height: 480,
        // zoom: { enabled: false },
      },
      grid: { padding: { top: 20, bottom: 20 } },
      colors,
      stroke: { width: 2 },
      // dataLabels: { enabled: false },
      tooltip: {
        x: {
          format: 'dd/MM HH:mm:ss',
        },
      },
      xaxis: {
        type: 'datetime',
        labels: {
          datetimeUTC: false,
        },
      },
      series: metrics
        .filter(m => selectedMetrics[m])
        .map(m => ({
          name: m,
          data: data.map(d => [d.Timestamp * 1000, d[m]]).reverse(),
        })),
    };

    chart = new ApexCharts(chartNode, chartOpts);
    chart.render();

    return () => chart.destroy();
  });
</script>

<div class="card">
  <div class="card-body">
    <div class="metrics d-flex justify-content-center">
      {#each metrics as m}
        <label class="form-check me-3">
          <input
            class="form-check-input"
            type="checkbox"
            checked={!!selectedMetrics[m]}
            on:click={e => {
              selectedMetrics = { ...selectedMetrics, [m]: e.target.checked };
            }}
          />
          <span class="form-check-label">{m}</span>
        </label>
      {/each}
    </div>
    <div bind:this={chartNode} />
  </div>
</div>
