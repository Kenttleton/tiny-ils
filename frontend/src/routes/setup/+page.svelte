<script lang="ts">
  import type { ActionData, PageData } from "./$types";

  let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
  <title>Setup — tiny-ils</title>
</svelte:head>

<div class="setup-card">
  <h1>Welcome to tiny-ils</h1>
  <p class="subtitle">Create your admin account to get started.</p>

  {#if form?.error}
    <p class="error">{form.error}</p>
  {/if}

  <form method="POST">
    <label>
      Public URL <span class="hint"
        >The URL users access this server from. Used for CORS and link
        generation.</span
      >
      <input
        type="url"
        name="publicUrl"
        value={form?.publicUrl ?? data.detectedPublicUrl}
        required
        placeholder="https://ils.example.com"
      />
    </label>

    <hr />

    <label>
      Display name (optional)
      <input
        type="text"
        name="displayName"
        value={form?.displayName ?? ""}
        autocomplete="name"
      />
    </label>

    <label>
      Email
      <input
        type="email"
        name="email"
        value={form?.email ?? ""}
        required
        autocomplete="email"
      />
    </label>

    <label>
      Password
      <input
        type="password"
        name="password"
        required
        autocomplete="new-password"
        minlength="8"
      />
    </label>

    <label>
      Confirm password
      <input
        type="password"
        name="confirm"
        required
        autocomplete="new-password"
        minlength="8"
      />
    </label>

    <button type="submit">Complete setup</button>
  </form>
</div>

<style>
  .setup-card {
    max-width: 480px;
    margin: 4rem auto;
    padding: 2rem;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
  }
  h1 {
    margin: 0 0 0.25rem;
    font-size: 1.5rem;
  }
  .subtitle {
    margin: 0 0 1.5rem;
    color: #6b7280;
    font-size: 0.9rem;
  }
  hr {
    border: none;
    border-top: 1px solid #e5e7eb;
    margin: 0.25rem 0;
  }
  form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.875rem;
    font-weight: 500;
  }
  .hint {
    font-weight: 400;
    font-size: 0.8rem;
    color: #6b7280;
  }
  input {
    padding: 0.5rem 0.75rem;
    border: 1px solid #d1d5db;
    border-radius: 4px;
    font-size: 1rem;
  }
  button {
    padding: 0.6rem;
    background: #111;
    color: #fff;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    margin-top: 0.5rem;
  }
  .error {
    color: #dc2626;
    font-size: 0.875rem;
    margin: 0 0 1rem;
  }
</style>
