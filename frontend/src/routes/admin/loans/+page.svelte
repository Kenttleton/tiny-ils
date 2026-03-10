<script lang="ts">
	import type { PageData } from './$types';
	import * as Table from '$lib/components/ui/table';

	let { data }: { data: PageData } = $props();

	function fmt(unixStr: string): string {
		const n = parseInt(unixStr, 10);
		if (!n) return '—';
		return new Date(n * 1000).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function isOverdue(dueDateStr: string, returnedAtStr: string): boolean {
		if (parseInt(returnedAtStr, 10)) return false;
		return Date.now() > parseInt(dueDateStr, 10) * 1000;
	}

	const totalPages = $derived(Math.ceil(data.total / data.limit));
</script>

<svelte:head>
	<title>Loans — Admin — tiny-ils</title>
</svelte:head>

<div class="mb-6 flex items-baseline gap-6">
	<h1 class="text-2xl font-bold">Loans</h1>
	<div class="flex gap-2">
		<a
			href="?active=true"
			class="rounded-full border px-3 py-1 text-xs {data.activeOnly ? 'border-foreground bg-foreground text-background' : 'border-border text-foreground'} no-underline"
		>Active</a>
		<a
			href="?active=false"
			class="rounded-full border px-3 py-1 text-xs {!data.activeOnly ? 'border-foreground bg-foreground text-background' : 'border-border text-foreground'} no-underline"
		>All</a>
	</div>
</div>

{#if data.loans.length === 0}
	<p class="text-muted-foreground text-sm">No {data.activeOnly ? 'active ' : ''}loans found.</p>
{:else}
	<p class="mb-3 text-xs text-muted-foreground">{data.total} loan{data.total === 1 ? '' : 's'}</p>
	<div class="overflow-x-auto">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Title</Table.Head>
					<Table.Head>Copy ID</Table.Head>
					<Table.Head>User ID</Table.Head>
					<Table.Head>Checked out</Table.Head>
					<Table.Head>Due date</Table.Head>
					{#if !data.activeOnly}
						<Table.Head>Returned</Table.Head>
					{/if}
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each data.loans as loan}
					{@const overdue = isOverdue(loan.due_date, loan.returned_at)}
					{@const dueSoon = !parseInt(loan.returned_at) && parseInt(loan.due_date) * 1000 - Date.now() < 86400000 * 3}
					<Table.Row class={overdue ? 'text-red-600' : ''}>
						<Table.Cell>
							<a href="/admin/catalog/{loan.curio_id}" class={overdue ? 'text-red-600 hover:underline' : 'hover:underline'}>{loan.curio_title || '(unknown)'}</a>
						</Table.Cell>
						<Table.Cell class="font-mono text-xs text-muted-foreground">{loan.copy_id.slice(0, 8)}&hellip;</Table.Cell>
						<Table.Cell class="font-mono text-xs text-muted-foreground">{loan.user_id.slice(0, 8)}&hellip;</Table.Cell>
						<Table.Cell>{fmt(loan.checked_out)}</Table.Cell>
						<Table.Cell class={dueSoon ? 'font-semibold text-amber-600' : ''}>{fmt(loan.due_date)}</Table.Cell>
						{#if !data.activeOnly}
							<Table.Cell>{fmt(loan.returned_at)}</Table.Cell>
						{/if}
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	</div>

	{#if totalPages > 1}
		<div class="mt-6 flex items-center gap-4 text-sm">
			{#if data.page > 1}
				<a href="?active={data.activeOnly}&page={data.page - 1}" class="text-foreground hover:underline">&larr; Prev</a>
			{/if}
			<span class="text-muted-foreground">Page {data.page} of {totalPages}</span>
			{#if data.page < totalPages}
				<a href="?active={data.activeOnly}&page={data.page + 1}" class="text-foreground hover:underline">Next &rarr;</a>
			{/if}
		</div>
	{/if}
{/if}
