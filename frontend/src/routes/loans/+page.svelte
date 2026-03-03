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

	function isOverdue(dueDateStr: string, closed: boolean): boolean {
		if (closed) return false;
		return Date.now() > parseInt(dueDateStr, 10) * 1000;
	}

	function daysUntilDue(dueDateStr: string): number {
		return Math.ceil((parseInt(dueDateStr, 10) * 1000 - Date.now()) / 86400000);
	}

	const totalPages = $derived(
		data.crossNode ? 1 : Math.ceil((data.total ?? 0) / (data.limit ?? 25))
	);
</script>

<svelte:head>
	<title>My Loans — tiny-ils</title>
</svelte:head>

<div class="header">
	<h1>My Loans</h1>
	<div class="filter">
		<a href="?active=true" class:active={data.activeOnly}>Active</a>
		<a href="?active=false" class:active={!data.activeOnly}>History</a>
	</div>
</div>

{#if data.crossNode}
	<p class="network-note">Showing loans across all connected partner libraries.</p>
{/if}

{#if data.loans.length === 0}
	<p class="empty">
		{data.activeOnly ? 'You have no active loans.' : 'No loan history found.'}
	</p>
{:else}
	<div class="loan-list">
		{#each data.loans as loan}
			{@const overdue = isOverdue(loan.due_date, loan.closed)}
			{@const days = !loan.closed ? daysUntilDue(loan.due_date) : null}
			<div class="loan-card" class:overdue class:returned={loan.closed}>
				<div class="loan-title">
					<a href="/catalog/{loan.curio_id}">{loan.curio_title || '(unknown)'}</a>
					{#if loan.is_digital}
						<span class="badge digital-badge">Digital</span>
					{/if}
				</div>
				<div class="loan-meta">
					{#if loan.node_name}
						<span class="node-tag">{loan.node_name}</span>
					{/if}
					<span>Checked out {fmt(loan.checked_out)}</span>
					{#if loan.closed}
						<span class="badge returned-badge">Returned</span>
					{:else if overdue}
						<span class="badge overdue-badge">Overdue — due {fmt(loan.due_date)}</span>
					{:else if days !== null && days <= 3}
						<span class="badge due-soon-badge">Due in {days} day{days === 1 ? '' : 's'}</span>
					{:else}
						<span>Due {fmt(loan.due_date)}</span>
					{/if}
				</div>
			</div>
		{/each}
	</div>

	{#if !data.crossNode && totalPages > 1}
		<div class="pagination">
			{#if (data.page ?? 1) > 1}
				<a href="?active={data.activeOnly}&page={(data.page ?? 1) - 1}">&larr; Prev</a>
			{/if}
			<span>Page {data.page ?? 1} of {totalPages}</span>
			{#if (data.page ?? 1) < totalPages}
				<a href="?active={data.activeOnly}&page={(data.page ?? 1) + 1}">Next &rarr;</a>
			{/if}
		</div>
	{/if}
{/if}

<style>
	.header { display: flex; align-items: baseline; gap: 1.5rem; margin-bottom: 1rem; }
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
	.network-note {
		font-size: 0.8125rem;
		color: #6b7280;
		margin: 0 0 1.25rem;
	}
	.empty { color: #6b7280; }
	.loan-list { display: flex; flex-direction: column; gap: 0.75rem; max-width: 600px; }
	.loan-card {
		padding: 1rem 1.25rem;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}
	.loan-card.overdue { border-color: #fca5a5; background: #fff5f5; }
	.loan-card.returned { opacity: 0.7; }
	.loan-title {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.loan-title a { font-weight: 600; color: #111; text-decoration: none; font-size: 0.9375rem; }
	.loan-title a:hover { text-decoration: underline; }
	.loan-meta { font-size: 0.8125rem; color: #6b7280; display: flex; gap: 1rem; align-items: center; flex-wrap: wrap; }
	.node-tag {
		font-size: 0.75rem;
		background: #f3f4f6;
		color: #374151;
		padding: 0.1rem 0.5rem;
		border-radius: 4px;
		font-weight: 500;
	}
	.badge { padding: 0.15rem 0.6rem; border-radius: 999px; font-size: 0.75rem; font-weight: 600; }
	.digital-badge { background: #eff6ff; color: #2563eb; }
	.returned-badge { background: #f0fdf4; color: #16a34a; }
	.overdue-badge { background: #fef2f2; color: #dc2626; }
	.due-soon-badge { background: #fffbeb; color: #d97706; }
	.pagination { display: flex; align-items: center; gap: 1rem; margin-top: 1.5rem; font-size: 0.875rem; }
	.pagination a { color: #111; text-decoration: none; }
	.pagination a:hover { text-decoration: underline; }
	.pagination span { color: #6b7280; }
</style>
