<script lang="ts">
	let { children } = $props();
</script>

<div class="admin-layout">
	<aside>
		<nav>
			<p class="section-label">Admin</p>
			<a href="/admin">Dashboard</a>
			<a href="/admin/curios">Curios</a>
			<a href="/admin/loans">Loans</a>
			<a href="/admin/transfers">Transfers</a>
			<a href="/admin/users">Users</a>
			<a href="/admin/peers">Network</a>
			<a href="/admin/settings">Settings</a>
		</nav>
	</aside>
	<div class="content">
		{@render children()}
	</div>
</div>

<style>
	.admin-layout {
		display: flex;
		gap: 2rem;
	}
	aside {
		width: 160px;
		flex-shrink: 0;
	}
	aside nav {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		position: sticky;
		top: 1rem;
	}
	.section-label {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: #9ca3af;
		margin: 0 0 0.5rem;
	}
	aside a {
		display: block;
		padding: 0.4rem 0.6rem;
		border-radius: 4px;
		text-decoration: none;
		color: #374151;
		font-size: 0.875rem;
	}
	aside a:hover {
		background: #f3f4f6;
		color: #111;
	}
	.content {
		flex: 1;
		min-width: 0;
	}
</style>
