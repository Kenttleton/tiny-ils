<script lang="ts">
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
</script>

<svelte:head>
	<title>Register — tiny-ils</title>
</svelte:head>

<div class="auth-card">
	<h1>Create account</h1>

	{#if form?.error}
		<p class="error">{form.error}</p>
	{/if}

	<form method="POST">
		<label>
			Display name (optional)
			<input
				type="text"
				name="displayName"
				value={form?.displayName ?? ''}
				autocomplete="name"
			/>
		</label>

		<label>
			Email
			<input type="email" name="email" value={form?.email ?? ''} required autocomplete="email" />
		</label>

		<label>
			Password
			<input type="password" name="password" required autocomplete="new-password" minlength="8" />
		</label>

		<button type="submit">Create account</button>
	</form>

	<p class="alt">Already have an account? <a href="/auth/login">Sign in</a></p>
</div>

<style>
	.auth-card {
		max-width: 400px;
		margin: 4rem auto;
		padding: 2rem;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	h1 {
		margin: 0 0 1.5rem;
		font-size: 1.5rem;
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
	}
	.error {
		color: #dc2626;
		font-size: 0.875rem;
		margin: 0 0 1rem;
	}
	.alt {
		font-size: 0.875rem;
		margin: 1rem 0 0;
	}
</style>
