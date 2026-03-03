<script lang="ts">
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
	<title>Sign in — tiny-ils</title>
</svelte:head>

<div class="auth-card">
	<h1>Sign in</h1>

	{#if form?.error}
		<p class="error">{form.error}</p>
	{/if}

	<form method="POST">
		<input type="hidden" name="next" value={data.next} />

		<label>
			Email
			<input type="email" name="email" value={form?.email ?? ''} required autocomplete="email" />
		</label>

		<label>
			Password
			<input type="password" name="password" required autocomplete="current-password" />
		</label>

		<button type="submit">Sign in</button>
	</form>

	<p class="alt">
		Don't have an account? <a href="/auth/register">Register</a>
	</p>

	<hr />

	<a href="/auth/google" class="sso-btn">Sign in with Google</a>

	<p class="cross-node">
		Account at another library? <a href="/auth/cross-node">Partner library sign-in</a>
	</p>
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
	hr {
		margin: 1.5rem 0;
		border: none;
		border-top: 1px solid #e5e7eb;
	}
	.sso-btn {
		display: block;
		text-align: center;
		padding: 0.6rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		text-decoration: none;
		color: #374151;
		font-size: 0.875rem;
	}
	.sso-btn:hover {
		background: #f9fafb;
	}
	.cross-node {
		font-size: 0.8125rem;
		color: #6b7280;
		margin: 0.75rem 0 0;
		text-align: center;
	}
</style>
