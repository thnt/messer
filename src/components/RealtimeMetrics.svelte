<script>
  import { metricStore } from '../store';
  import MetricBox from './MetricBox.svelte';

  metricStore.getRecents(100);
</script>

<div class="row row-deck row-cards">
  <div class="col-12">
    <slot name="header"><h3>Overview</h3></slot>
  </div>
  <div class="col-sm-6 col-lg-4">
    <MetricBox title="Temperature" unit="°C" dataKey="Temperature" data={$metricStore.recents} />
  </div>
  <div class="col-sm-6 col-lg-4">
    <MetricBox title="Pressure" unit="Barg" dataKey="Pressure" data={$metricStore.recents} />
  </div>
  <div class="col-12 col-lg-4">
    <MetricBox title="Massflow" unit="NCMH" dataKey="Massflow" data={$metricStore.recents} />
  </div>
</div>
