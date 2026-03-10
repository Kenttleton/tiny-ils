<script lang="ts">
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faBook, faNetworkWired, faArrowRight } from '@fortawesome/free-solid-svg-icons';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Admin — tiny-ils</title>
</svelte:head>

<h1 class="mb-6 text-2xl font-bold">Dashboard</h1>

<div class="mb-8 flex flex-wrap gap-4">
	<Card.Root class="min-w-[140px]">
		<Card.Content class="flex flex-col gap-1 p-6">
			<FontAwesomeIcon icon={faBook} class="h-5 w-5 text-muted-foreground" />
			<span class="text-3xl font-bold">{data.total}</span>
			<span class="text-xs text-muted-foreground">Curios in catalog</span>
		</Card.Content>
	</Card.Root>
	<Card.Root class="min-w-[140px]">
		<Card.Content class="flex flex-col gap-1 p-6">
			<FontAwesomeIcon icon={faNetworkWired} class="h-5 w-5 text-muted-foreground" />
			<span class="text-3xl font-bold">{data.peers.length}</span>
			<span class="text-xs text-muted-foreground">Partner libraries</span>
		</Card.Content>
	</Card.Root>
</div>

<section class="mb-8">
	<h2 class="mb-3 flex items-baseline gap-3 text-base font-semibold">
		Network
		<a href="/admin/peers" class="inline-flex items-center gap-1 text-xs font-normal text-muted-foreground hover:text-foreground hover:underline">Manage <FontAwesomeIcon icon={faArrowRight} class="h-3.5 w-3.5" /></a>
	</h2>
	{#if data.peers.length === 0}
		<p class="text-muted-foreground text-sm">No peer nodes connected. <a href="/admin/peers" class="underline">Register one</a>.</p>
	{:else}
		<div class="overflow-x-auto">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Name</Table.Head>
						<Table.Head>Address</Table.Head>
						<Table.Head>Library ID</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.peers as peer}
						<Table.Row>
							<Table.Cell>{peer.displayName || '—'}</Table.Cell>
							<Table.Cell class="font-mono text-xs break-all">{peer.address}</Table.Cell>
							<Table.Cell class="font-mono text-xs break-all">{peer.nodeId}</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	{/if}
</section>

<section class="mb-8">
	<h2 class="mb-3 text-base font-semibold">Recent curios</h2>
	{#if data.recentCurios.length === 0}
		<p class="text-muted-foreground text-sm">No curios yet. <a href="/admin/curios/new" class="underline">Add one</a>.</p>
	{:else}
		<div class="overflow-x-auto">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Title</Table.Head>
						<Table.Head>Type</Table.Head>
						<Table.Head>Format</Table.Head>
						<Table.Head></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.recentCurios as c (c.id)}
						<Table.Row>
							<Table.Cell>{c.title}</Table.Cell>
							<Table.Cell>{c.mediaType}</Table.Cell>
							<Table.Cell>{c.formatType}</Table.Cell>
							<Table.Cell>
								<a href="/admin/curios/{c.id}/edit" class="text-sm text-foreground hover:underline">Edit</a>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
		<p class="mt-3 text-sm">
			<a href="/admin/curios" class="text-foreground hover:underline">View all curios →</a>
		</p>
	{/if}
</section>
