<script>
  import { userStore } from '../store';
  import Button from '../components/Button.svelte';

  let username, password;

  const onLogin = () => {
    userStore.login(username, password);
  };
</script>

<div class="page page-center">
  <div class="container-tight py-4">
    <h1 class="text-center mb-4">MESSER</h1>
    <form
      on:submit|preventDefault={onLogin}
      class="card card-md"
      action="."
      method="get"
      autocomplete="off"
    >
      <div class="card-body">
        <h2 class="card-title text-center mb-4">Login to your account</h2>
        <div class="mb-3">
          <label class="form-label" for="username">Username</label>
          <input
            type="text"
            class="form-control"
            placeholder="Username"
            required
            bind:value={username}
          />
        </div>
        <div class="mb-2">
          <label class="form-label" for="password">Password</label>
          <div class="input-group input-group-flat">
            <input
              type="password"
              class="form-control"
              placeholder="Password"
              required
              bind:value={password}
            />
          </div>
        </div>
        <div class="form-footer">
          {#if $userStore.error}
            <div class="alert alert-danger py-2" role="alert">{$userStore.error}</div>
          {/if}
          <Button loading={$userStore.loading} submit className="w-100">Log me in</Button>
        </div>
      </div>
    </form>
  </div>
</div>
