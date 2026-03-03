<script lang="ts">
	import type { PageData } from './$types';

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

<div class="header">
	<h1>Loans</h1>
	<div class="filter">
		<a href="?active=true" class:active={data.activeOnly}>Active</a>
		<a href="?active=false" class:active={!data.activeOnly}>All</a>
	</div>
</div>

{#if data.loans.length === 0}
	<p class="empty">No {data.activeOnly ? 'active ' : ''}loans found.</p>
{:else}
	<p class="count">{data.total} loan{data.total === 1 ? '' : 's'}</p>
	<div class="table-wrap">
		<table>
			<thead>
				<tr>
					<th>Title</th>
					<th>Copy ID</th>
					<th>User ID</th>
					<th>Checked out</th>
					<th>Due date</th>
					{#if !data.activeOnly}
						<th>Returned</th>
					{/if}
				</tr>
			</thead>
			<tbody>
				{#each data.loans as loan}
					<tr class:overdue={isOverdue(loan.due_date, loan.returned_at)}>
						<td>
							<a href="/admin/catalog/{loan.curio_id}">{loan.curio_title || '(unknown)'}</a>
						</td>
						<td class="mono">{loan.copy_id.slice(0, 8)}&hellip;</td>
						<td class="mono">{loan.user_id.slice(0, 8)}&hellip;</td>
						<td>{fmt(loan.checked_out)}</td>
						<td class:due-soon={!parseInt(loan.returned_at) && parseInt(loan.due_date) * 1000 - Date.now() < 86400000 * 3}>
							{fmt(loan.due_date)}
						</td>
						{#if !data.activeOnly}
							<td>{fmt(loan.returned_at)}</td>
						{/if}
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	{#if totalPages > 1}
		<div class="pagination">
			{#if data.page > 1}
				<a href="?active={data.activeOnly}&page={data.page - 1}">&larr; Prev</a>
			{/if}
			<span>Page {data.page} of {totalPages}</span>
			{#if data.page < totalPages}
				<a href="?active={data.activeOnly}&page={data.page + 1}">Next &rarr;</a>
			{/if}
		</div>
	{/if}
{/if}

<style>
	.header { display: flex; align-items: baseline; gap: 1.5rem; margin-bottom: 1.5rem; }
	h1 { margin: 0; }
	.filter { display: flex; gap: 0.5rem; }
	.filter a {
		font-size: 0.8125rem;
		padding: 0.25rem 0.75rem;
		border-radius: 999px;
		border: 1px solid #d1d5db;
		color: #374151;
		text-decoration: none;
	}
	.filter a.active { background: #111; color: #fff; border-color: #111; }
	.count { font-size: 0.8125rem; color: #6b7280; margin: 0 0 0.75rem; }
	.empty { color: #6b7280; }
	.table-wrap { overflow-x: auto; }
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	th { text-align: left; font-size: 0.75rem; font-weight: 600; color: #6b7280; padding: 0.5rem 0.75rem; border-bottom: 1px solid #e5e7eb; white-space: nowrap; }
	td { padding: 0.625rem 0.75rem; border-bottom: 1px solid #f3f4f6; vertical-align: middle; }
	td a { color: #111; text-decoration: none; }
	td a:hover { text-decoration: underline; }
	.mono { font-family: monospace; font-size: 0.8125rem; color: #6b7280; }
	tr.overdue td { color: #dc2626; }
	tr.overdue td a { color: #dc2626; }
	td.due-soon { color: #d97706; font-weight: 600; }
	.pagination { display: flex; align-items: center; gap: 1rem; margin-top: 1.5rem; font-size: 0.875rem; }
	.pagination a { color: #111; text-decoration: none; }
	.pagination a:hover { text-decoration: underline; }
	.pagination span { color: #6b7280; }
</style>
