<script>
  import { userStore } from '../store';
  import Button from './Button.svelte';

  const onLogout = () => {
    if (!confirm('Are you sure want to logout?')) return;

    userStore.logout();
  };
</script>

<header class="navbar navbar-expand-md navbar-dark navbar-overlap d-print-none">
  <div class="container-xl">
    <h1 class="navbar-brand navbar-brand-autodark d-none-navbar-horizontal pe-0 pe-md-3">MESSER</h1>
    <div class="order-md-last">
      <div class="d-flex align-items-center">
        <strong class="me-1">{$userStore.user.name}</strong>
        <Button ghost type="danger" on:click={onLogout}>Logout</Button>
      </div>
    </div>
  </div>
</header>
