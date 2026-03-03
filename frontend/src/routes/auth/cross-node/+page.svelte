<script lang="ts">
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
	<title>Partner library sign-in — tiny-ils</title>
</svelte:head>

<div class="auth-card">
	<h1>Sign in from another library</h1>
	<p class="description">
		If your account is registered at a different partner library, enter that library's connection
		details and your user ID to sign in here.
	</p>

	{#if form?.error}
		<p class="error">{form.error}</p>
	{/if}

	<form method="POST">
		<input type="hidden" name="next" value={data.next} />

		<label>
			Home library address
			<input
				type="text"
				name="home_node_address"
				value={form?.homeNodeAddress ?? ''}
				required
				placeholder="e.g. library.example.org:50153"
				autocomplete="off"
			/>
			<span class="hint">The gRPC address of the library where your account lives.</span>
		</label>

		<label>
			Home library ID
			<input
				type="text"
				name="home_node_id"
				value={form?.homeNodeId ?? ''}
				required
				placeholder="e.g. abc123..."
				autocomplete="off"
			/>
			<span class="hint">Found under Network on your home library's admin page.</span>
		</label>

		<label>
			Your user ID
			<input
				type="text"
				name="user_id"
				value={form?.userId ?? ''}
				required
				placeholder="e.g. 550e8400-..."
				autocomplete="off"
			/>
			<span class="hint">Found on your profile page at your home library.</span>
		</label>

		<button type="submit">Sign in</button>
	</form>

	<p class="alt">
		Have an account here? <a href="/auth/login">Local sign in</a>
	</p>
</div>

<style>
	.auth-card {
		max-width: 440px;
		margin: 4rem auto;
		padding: 2rem;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
	}
	h1 {
		margin: 0 0 0.5rem;
		font-size: 1.5rem;
	}
	.description {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0 0 1.5rem;
		line-height: 1.5;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
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
		font-family: monospace;
	}
	.hint {
		font-size: 0.75rem;
		color: #9ca3af;
		font-weight: 400;
		font-family: inherit;
	}
	button {
		padding: 0.6rem;
		background: #111;
		color: #fff;
		border: none;
		border-radius: 4px;
		font-size: 1rem;
		cursor: pointer;
		margin-top: 0.25rem;
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
