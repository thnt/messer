<script>
  import Dashboard from '../components/Dashboard.svelte';
  const params = new URLSearchParams(location.search);
  const src = params.get('src') || '';
</script>

<div class="dashboard">
  <Dashboard {src} />
</div>

<style>
  .dashboard {
    min-width: 1024px;
    background: #fff;
    overflow: auto;
    flex: 1;
  }
</style>
