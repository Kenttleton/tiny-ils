<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	// Destructure initial values so $state doesn't capture reactive props directly.
	const { q: initialQ, mediaType: initialMediaType } = data;
	let q = $state(initialQ);
	let mediaType = $state(initialMediaType);

	const MEDIA_TYPES = ['', 'BOOK', 'VIDEO', 'AUDIO', 'GAME', 'THING'];

	function search() {
		const params = new URLSearchParams();
		if (q) params.set('q', q);
		if (mediaType) params.set('mediaType', mediaType);
		goto(`/browse?${params}`, { replaceState: true });
	}
</script>

<svelte:head>
	<title>Browse — tiny-ils</title>
</svelte:head>

<h1>Browse</h1>

<form class="search-bar" onsubmit={(e) => { e.preventDefault(); search(); }}>
	<input type="search" bind:value={q} placeholder="Search curios…" />
	<select bind:value={mediaType} onchange={search}>
		{#each MEDIA_TYPES as mt}
			<option value={mt}>{mt || 'All types'}</option>
		{/each}
	</select>
	<button type="submit">Search</button>
</form>

<p class="count">{data.total} item{data.total !== 1 ? 's' : ''}</p>

{#if data.curios.length === 0}
	<p class="empty">No curios found.</p>
{:else}
	<ul class="curio-list">
		{#each data.curios as curio (curio.id)}
			<li>
				<a href="/browse/{curio.id}" class="curio-card">
					<span class="title">{curio.title}</span>
					<span class="meta">{curio.mediaType} · {curio.formatType}</span>
					{#if curio.tags?.length}
						<span class="tags">{curio.tags.join(', ')}</span>
					{/if}
				</a>
			</li>
		{/each}
	</ul>
{/if}

<style>
	h1 { margin: 0 0 1.5rem; }
	.search-bar {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}
	.search-bar input {
		flex: 1;
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 1rem;
	}
	.search-bar select, .search-bar button {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
		background: #fff;
		cursor: pointer;
	}
	.count { color: #6b7280; font-size: 0.875rem; margin: 0 0 1rem; }
	.empty { color: #6b7280; }
	.curio-list { list-style: none; padding: 0; margin: 0; display: grid; gap: 0.75rem; }
	.curio-card {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		padding: 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		text-decoration: none;
		color: inherit;
		transition: border-color 0.1s;
	}
	.curio-card:hover { border-color: #9ca3af; }
	.title { font-weight: 600; }
	.meta { font-size: 0.75rem; color: #6b7280; text-transform: uppercase; letter-spacing: 0.05em; }
	.tags { font-size: 0.75rem; color: #9ca3af; }
</style>
