<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';
	import * as Table from '$lib/components/ui/table';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faArrowLeft, faFloppyDisk, faPlus } from '@fortawesome/free-solid-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	const { curio, copies } = $derived(data);

	const MEDIA_TYPES = ['THING', 'BOOK', 'VIDEO', 'AUDIO', 'GAME'];
	const FORMAT_TYPES = ['PHYSICAL', 'DIGITAL', 'BOTH'];
	const CONDITIONS = ['NEW', 'GOOD', 'FAIR', 'POOR'];

	function statusClass(status: string): string {
		const map: Record<string, string> = {
			AVAILABLE: 'bg-green-100 text-green-800',
			ON_LOAN: 'bg-yellow-100 text-yellow-800',
			REQUESTED: 'bg-blue-100 text-blue-800',
			IN_TRANSIT: 'bg-purple-100 text-purple-800',
		};
		return map[status] ?? 'bg-zinc-100 text-zinc-700';
	}
</script>

<svelte:head>
	<title>Edit {curio.title} — Admin — tiny-ils</title>
</svelte:head>

<a href="/admin/curios" class="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"><FontAwesomeIcon icon={faArrowLeft} class="h-3.5 w-3.5" /> Back to curios</a>
<h1 class="mb-6 text-2xl font-bold">Edit curio</h1>

{#if form?.error}
	<Alert.Root variant="destructive" class="mb-4">
		<Alert.Description>{form.error}</Alert.Description>
	</Alert.Root>
{/if}

<form method="POST" action="?/update" class="flex max-w-[600px] flex-col gap-4">
	<div class="flex flex-col gap-1.5">
		<Label for="mediaType">Media type</Label>
		<select
			id="mediaType"
			name="mediaType"
			class="rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
		>
			{#each MEDIA_TYPES as mt}
				<option value={mt} selected={mt === curio.mediaType}>{mt}</option>
			{/each}
		</select>
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="formatType">Format type</Label>
		<select
			id="formatType"
			name="formatType"
			class="rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
		>
			{#each FORMAT_TYPES as ft}
				<option value={ft} selected={ft === curio.formatType}>{ft}</option>
			{/each}
		</select>
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="title">Title *</Label>
		<Input id="title" type="text" name="title" value={curio.title} required />
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="description">Description</Label>
		<textarea
			id="description"
			name="description"
			rows="4"
			class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
		>{curio.description ?? ''}</textarea>
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="tags">Tags (comma-separated)</Label>
		<Input id="tags" type="text" name="tags" value={curio.tags?.join(', ') ?? ''} />
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="barcode">Barcode</Label>
		<Input id="barcode" type="text" name="barcode" value={curio.barcode ?? ''} />
	</div>

	<div>
		<Button type="submit"><FontAwesomeIcon icon={faFloppyDisk} class="mr-1.5 h-3.5 w-3.5" />Save changes</Button>
	</div>
</form>

{#if curio.formatType === 'PHYSICAL' || curio.formatType === 'BOTH'}
<section class="mt-10 max-w-[600px] border-t border-border pt-6">
	<h2 class="mb-3 text-base font-semibold">Physical copies ({copies.length})</h2>

	{#if form?.copyError}
		<Alert.Root variant="destructive" class="mb-4">
			<Alert.Description>{form.copyError}</Alert.Description>
		</Alert.Root>
	{/if}

	{#if copies.length > 0}
		<div class="overflow-x-auto">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Condition</Table.Head>
						<Table.Head>Location</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head>ID</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each copies as copy}
						<Table.Row>
							<Table.Cell>{copy.condition}</Table.Cell>
							<Table.Cell>{copy.location || '—'}</Table.Cell>
							<Table.Cell>
								<span class="rounded-full px-2 py-0.5 text-xs font-semibold {statusClass(copy.status)}">
									{copy.status}
								</span>
							</Table.Cell>
							<Table.Cell class="font-mono text-xs text-muted-foreground">{copy.id.slice(0, 8)}…</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	{:else}
		<p class="text-muted-foreground text-sm">No physical copies yet.</p>
	{/if}

	<form method="POST" action="?/addCopy" class="mt-4">
		<h3 class="mb-2 text-sm font-semibold">Add a copy</h3>
		<div class="flex flex-wrap items-end gap-3">
			<div class="flex flex-1 flex-col gap-1.5">
				<Label for="condition">Condition</Label>
				<select
					id="condition"
					name="condition"
					class="rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
				>
					{#each CONDITIONS as c}
						<option value={c} selected={c === 'GOOD'}>{c}</option>
					{/each}
				</select>
			</div>
			<div class="flex flex-1 flex-col gap-1.5">
				<Label for="location">Location</Label>
				<Input id="location" type="text" name="location" placeholder="e.g. Shelf A3" />
			</div>
			<Button type="submit" variant="outline"><FontAwesomeIcon icon={faPlus} class="mr-1.5 h-3.5 w-3.5" />Add copy</Button>
		</div>
	</form>
</section>
{/if}
