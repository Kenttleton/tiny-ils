<script lang="ts">
	import { page } from '$app/stores';
	import { isManager } from '$lib/auth';

	let { data, children } = $props();

	const user = $derived(data.user);
	const nodeId = $derived(data.nodeId);
	const manager = $derived(user ? isManager(user.claims, nodeId) : false);
	const isAuthPage = $derived($page.url.pathname.startsWith('/auth/'));
</script>

<svelte:head>
	<title>tiny-ils</title>
</svelte:head>

{#if !isAuthPage && user}
	<nav>
		<a href="/" class="brand">tiny-ils</a>
		<div class="links">
			<a href="/browse">Browse</a>
			<a href="/loans">My Loans</a>
			{#if manager}
				<a href="/admin">Admin</a>
			{/if}
		</div>
		<form method="POST" action="/auth/logout">
			<button type="submit">Sign out</button>
		</form>
	</nav>
{/if}

<main>
	{@render children()}
</main>

<style>
	nav {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		padding: 0.75rem 1.5rem;
		border-bottom: 1px solid #e5e7eb;
		background: #fff;
	}
	.brand {
		font-weight: 700;
		font-size: 1.1rem;
		text-decoration: none;
		color: #111;
	}
	.links {
		display: flex;
		gap: 1rem;
		flex: 1;
	}
	.links a {
		color: #374151;
		text-decoration: none;
	}
	.links a:hover {
		color: #111;
	}
	button {
		background: none;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		padding: 0.25rem 0.75rem;
		cursor: pointer;
		color: #374151;
	}
	main {
		padding: 2rem 1.5rem;
		max-width: 1100px;
		margin: 0 auto;
	}
</style>
