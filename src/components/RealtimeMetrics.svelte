<script>
  import { metricStore } from '../store';
  import MetricBox from './MetricBox.svelte';

  metricStore.getRecents(100);
</script>

<div class="row row-deck row-cards">
  <div class="col-12">
    <slot name="header"><h3>Overview</h3></slot>
  </div>
  <div class="col-md-4">
    <MetricBox title="Pressure" dataKey="Pressure" data={$metricStore.recents} />
  </div>
  <div class="col-md-4">
    <MetricBox title="Temperature" dataKey="Temperature" data={$metricStore.recents} />
  </div>
  <div class="col-md-4">
    <MetricBox title="Massflow" dataKey="Massflow" data={$metricStore.recents} />
  </div>
</div>
