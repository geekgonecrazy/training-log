<script lang="ts">
  import { authStore } from '$lib/stores/auth';
  import { ApiError } from '$lib/api/client';
  import { auth as authApi } from '$lib/api/endpoints';
  import Icon from '$lib/components/Icon.svelte';

  let email = '';
  let password = '';
  let name = '';
  let rememberMe = true;
  let mode: 'login' | 'register' = 'login';
  let busy = false;
  let error = '';

  async function submit() {
    busy = true;
    error = '';
    try {
      if (mode === 'register') {
        await authApi.register(email, password, name);
      }
      await authStore.login(email, password, rememberMe);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 403) {
          error = 'Registration is closed on this server.';
        } else if (err.status === 401) {
          error = 'Invalid email or password.';
        } else if (err.status === 409) {
          error = 'An account with that email already exists.';
        } else {
          error = err.message;
        }
      } else {
        error = String(err);
      }
    } finally {
      busy = false;
    }
  }
</script>

<div class="app-shell no-tabs login-shell">
  <div class="brand">
    <span class="icon-circle big"><Icon name="flame" size={36} /></span>
    <h1 class="title-large" style="margin-top: 0.75rem;">Training Log</h1>
    <p class="subtitle">{mode === 'login' ? 'Welcome back.' : 'Build the habit.'}</p>
  </div>

  <div class="card-elevated card">
    <form on:submit|preventDefault={submit} class="stack">
      {#if mode === 'register'}
        <div class="field">
          <label for="name">Display name</label>
          <input id="name" type="text" bind:value={name} autocomplete="name" placeholder="Aaron" />
        </div>
      {/if}

      <div class="field">
        <label for="email">Email</label>
        <input id="email" type="email" bind:value={email} autocomplete="email" required placeholder="you@example.com" />
      </div>

      <div class="field">
        <label for="password">Password</label>
        <input
          id="password"
          type="password"
          bind:value={password}
          autocomplete={mode === 'login' ? 'current-password' : 'new-password'}
          required
          minlength={mode === 'register' ? 8 : 1}
          placeholder="••••••••"
        />
      </div>

      <label class="row" style="text-transform: none; letter-spacing: 0; margin-bottom: 0.25rem; color: var(--text-2); font-size: 0.95rem; font-weight: 500;">
        <input type="checkbox" bind:checked={rememberMe} style="width: auto; margin: 0;" />
        <span>Remember me</span>
      </label>

      <button class="primary" type="submit" disabled={busy}>
        {busy ? '…' : mode === 'login' ? 'Sign in' : 'Create account'}
      </button>

      {#if error}<p class="error">{error}</p>{/if}
    </form>
  </div>

  <div style="text-align: center; margin-top: 1.25rem;">
    <button
      class="ghost"
      on:click={() => {
        mode = mode === 'login' ? 'register' : 'login';
        error = '';
      }}
      style="background: transparent; border: none; min-height: 0; color: var(--muted); padding: 0.5rem;"
    >
      {mode === 'login' ? 'Need an account? Register' : 'Have an account? Sign in'}
    </button>
  </div>
</div>

<style>
  .login-shell {
    min-height: 100vh;
    min-height: 100dvh;
    display: flex;
    flex-direction: column;
    justify-content: center;
    max-width: 420px;
  }
  .brand { text-align: center; margin-bottom: 1.5rem; }
  .brand .icon-circle.big {
    width: 64px;
    height: 64px;
    border-radius: 20px;
    color: var(--accent);
    background: rgba(0, 210, 168, 0.08);
    border-color: rgba(0, 210, 168, 0.25);
    margin: 0 auto;
  }
  .brand h1 { margin: 0; }
</style>
