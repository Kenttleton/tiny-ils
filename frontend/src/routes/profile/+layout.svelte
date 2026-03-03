<script lang="ts">
	let { children } = $props();
</script>

<div class="profile-layout">
	<aside>
		<nav>
			<p class="section-label">Profile</p>
			<a href="/profile/loans">My Loans</a>
			<a href="/profile/settings">Settings</a>
		</nav>
	</aside>
	<div class="content">
		{@render children()}
	</div>
</div>

<style>
	.profile-layout {
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
