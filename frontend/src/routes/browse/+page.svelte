<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faMagnifyingGlass } from '@fortawesome/free-solid-svg-icons';
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

<h1 class="mb-6 text-2xl font-bold">Browse</h1>

<form class="mb-4 flex gap-2" onsubmit={(e) => { e.preventDefault(); search(); }}>
	<Input type="search" bind:value={q} placeholder="Search curios…" class="flex-1" />
	<select
		bind:value={mediaType}
		onchange={search}
		class="rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
	>
		{#each MEDIA_TYPES as mt}
			<option value={mt}>{mt || 'All types'}</option>
		{/each}
	</select>
	<Button type="submit"><FontAwesomeIcon icon={faMagnifyingGlass} class="mr-1.5 h-3.5 w-3.5" />Search</Button>
</form>

<p class="mb-4 text-sm text-muted-foreground">{data.total} item{data.total !== 1 ? 's' : ''}</p>

{#if data.curios.length === 0}
	<p class="text-muted-foreground text-sm">No curios found.</p>
{:else}
	<ul class="grid gap-3 p-0" style="list-style:none">
		{#each data.curios as curio (curio.id)}
			<li>
				<a
					href="/browse/{curio.id}"
					class="flex flex-col gap-1 rounded-md border border-border p-4 text-inherit no-underline transition-colors hover:border-muted-foreground"
				>
					<span class="font-semibold">{curio.title}</span>
					<span class="text-xs uppercase tracking-wide text-muted-foreground">{curio.mediaType} · {curio.formatType}</span>
					{#if curio.tags?.length}
						<span class="text-xs text-muted-foreground/70">{curio.tags.join(', ')}</span>
					{/if}
				</a>
			</li>
		{/each}
	</ul>
{/if}
