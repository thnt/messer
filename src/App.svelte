<script>
  import dayjs from 'dayjs';
  import relativeTime from 'dayjs/plugin/relativeTime';

  import Dashboard from './pages/Dashboard.svelte';
  import Login from './pages/Login.svelte';
  import Main from './pages/Main.svelte';
  import { userStore } from './store';

  dayjs.extend(relativeTime);

  userStore.auth();
</script>

{#if $userStore.user?.id}
  {#if location.pathname.match(/\/dashboard\/?/)}
    <Dashboard />
  {:else}
    <Main />
  {/if}
{:else if !$userStore.init}
  <div class="text-center mt-5">
    <div class="spinner-border text-primary" role="status">
      <span class="visually-hidden">Loading...</span>
    </div>
  </div>
{:else}
  <Login />
{/if}
