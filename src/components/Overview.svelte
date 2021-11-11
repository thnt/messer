<script>
  import { metricStore } from '../store';
  import device from '../assets/device.jpg';

  metricStore.getRecents(200);

  let currentValues;
  $: currentValues = $metricStore.recents[0] || {};
</script>

<div class="card">
  <div class="card-body">
    <div class="row text-center border-bottom pb-3">
      <div class="col">
        <div class="title">NORMAL</div>
        <div class="small">(NCM)</div>
        <div class="value mt-1">{currentValues.TTflowL1000 ?? '-'}</div>
      </div>
      <div class="col border-start border-end">
        <div class="title">OVER MAX</div>
        <div class="small">(NCM)</div>
        <div class="value mt-1">{currentValues.TTflowG1000 ?? '-'}</div>
      </div>
      <div class="col">
        <div class="title">TOTAL FLOW</div>
        <div class="small">(NCM)</div>
        <div class="value mt-1">{currentValues.TotalFlow ?? '-'}</div>
      </div>
    </div>
    <div class="text-center device mt-3">
      <img class="h-100" src={device} alt="" />
    </div>
    <ul class="list-group mb-4 mx-auto">
      <li class="list-group-item d-flex flex-wrap justify-content-between align-items-center">
        <div class="title me-3">Current flow:</div>
        <div class="value">
          {currentValues.Massflow ?? '-'} <span class="unit">NCMH</span>
        </div>
      </li>
      <li class="list-group-item d-flex flex-wrap justify-content-between align-items-center">
        <div class="title me-3">Temperature:</div>
        <div class="value">
          {currentValues.Temperature ?? '-'} <span class="unit">°C</span>
        </div>
      </li>
      <li class="list-group-item d-flex flex-wrap justify-content-between align-items-center">
        <div class="title me-3">Pressure:</div>
        <div class="value">
          {currentValues.Pressure ?? '-'} <span class="unit">Barg</span>
        </div>
      </li>
    </ul>
  </div>
</div>

<style>
  .title,
  .value {
    font-weight: 600;
  }
  .value {
    font-size: 1.4rem;
    white-space: nowrap;
  }
  .device {
    height: 120px;
    margin-bottom: -30px;
  }
  .list-group {
    max-width: 500px;
    background-color: #fff;
  }
  .unit {
    width: 4.5rem;
    font-size: 1rem;
    text-align: right;
    display: inline-block;
  }
</style>
