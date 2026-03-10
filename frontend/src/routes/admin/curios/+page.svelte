<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData, ActionData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Alert from '$lib/components/ui/alert';
	import * as Table from '$lib/components/ui/table';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faPlus, faPen, faTrash } from '@fortawesome/free-solid-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	const { q: initialQ } = data;
	let q = $state(initialQ);
</script>

<svelte:head>
	<title>Curios — Admin — tiny-ils</title>
</svelte:head>

<div class="mb-6 flex items-center justify-between">
	<h1 class="text-2xl font-bold">Curios ({data.total})</h1>
	<Button href="/admin/curios/new"><FontAwesomeIcon icon={faPlus} class="mr-1.5 h-3.5 w-3.5" />Add curio</Button>
</div>

{#if form?.error}
	<Alert.Root variant="destructive" class="mb-4">
		<Alert.Description>{form.error}</Alert.Description>
	</Alert.Root>
{/if}

<form
	class="mb-4 flex gap-2"
	onsubmit={(e) => {
		e.preventDefault();
		goto(`/admin/curios?q=${encodeURIComponent(q)}`, { replaceState: true });
	}}
>
	<Input type="search" bind:value={q} placeholder="Search…" class="flex-1" />
	<Button type="submit" variant="outline">Search</Button>
</form>

{#if data.curios.length === 0}
	<p class="text-muted-foreground text-sm">No curios found.</p>
{:else}
	<div class="overflow-x-auto">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Title</Table.Head>
					<Table.Head>Type</Table.Head>
					<Table.Head>Format</Table.Head>
					<Table.Head>Tags</Table.Head>
					<Table.Head></Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each data.curios as c (c.id)}
					<Table.Row>
						<Table.Cell>{c.title}</Table.Cell>
						<Table.Cell>{c.mediaType}</Table.Cell>
						<Table.Cell>{c.formatType}</Table.Cell>
						<Table.Cell class="max-w-[200px] text-xs text-muted-foreground">{c.tags?.join(', ') ?? ''}</Table.Cell>
						<Table.Cell>
							<div class="flex items-center gap-2 whitespace-nowrap">
								<a href="/admin/curios/{c.id}/edit" class="inline-flex items-center gap-1 text-sm text-foreground hover:underline"><FontAwesomeIcon icon={faPen} class="h-3.5 w-3.5" />Edit</a>
								<form method="POST" action="?/delete" class="inline">
									<input type="hidden" name="id" value={c.id} />
									<button
										type="submit"
										class="rounded border border-red-300 px-2 py-0.5 text-xs text-red-600 hover:bg-red-50"
										onclick={(e) => {
											if (!confirm(`Delete "${c.title}"?`)) e.preventDefault();
										}}
									>
										<FontAwesomeIcon icon={faTrash} class="mr-1.5 h-3.5 w-3.5" />Delete
									</button>
								</form>
							</div>
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>
{/if}
