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

	function daysUntilDue(dueDateStr: string): number {
		return Math.ceil((parseInt(dueDateStr, 10) * 1000 - Date.now()) / 86400000);
	}

	const totalPages = $derived(Math.ceil(data.total / data.limit));
</script>

<svelte:head>
	<title>My Loans — tiny-ils</title>
</svelte:head>

<div class="mb-6 flex items-baseline gap-6">
	<h1 class="text-2xl font-bold">My Loans</h1>
	<div class="flex gap-2">
		<a
			href="?active=true"
			class="rounded-full border px-3 py-1 text-xs no-underline {data.activeOnly ? 'border-foreground bg-foreground text-background' : 'border-border text-foreground'}"
		>Active</a>
		<a
			href="?active=false"
			class="rounded-full border px-3 py-1 text-xs no-underline {!data.activeOnly ? 'border-foreground bg-foreground text-background' : 'border-border text-foreground'}"
		>History</a>
	</div>
</div>

{#if data.loans.length === 0}
	<p class="text-muted-foreground text-sm">
		{data.activeOnly ? 'You have no active loans.' : 'No loan history found.'}
	</p>
{:else}
	<div class="flex max-w-[600px] flex-col gap-3">
		{#each data.loans as loan}
			{@const overdue = isOverdue(loan.due_date, loan.returned_at)}
			{@const returned = !!parseInt(loan.returned_at, 10)}
			{@const days = !returned ? daysUntilDue(loan.due_date) : null}
			<div class="flex flex-col gap-1.5 rounded-lg border px-5 py-4 {overdue ? 'border-red-300 bg-red-50/50' : 'border-border'} {returned ? 'opacity-70' : ''}">
				<div>
					<a href="/browse/{loan.curio_id}" class="font-semibold text-foreground hover:underline">{loan.curio_title || '(unknown)'}</a>
				</div>
				<div class="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
					<span>Checked out {fmt(loan.checked_out)}</span>
					{#if returned}
						<span class="rounded-full bg-green-100 px-2 py-0.5 text-xs font-semibold text-green-800">Returned {fmt(loan.returned_at)}</span>
					{:else if overdue}
						<span class="rounded-full bg-red-100 px-2 py-0.5 text-xs font-semibold text-red-700">Overdue — due {fmt(loan.due_date)}</span>
					{:else if days !== null && days <= 3}
						<span class="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-700">Due in {days} day{days === 1 ? '' : 's'}</span>
					{:else}
						<span>Due {fmt(loan.due_date)}</span>
					{/if}
				</div>
			</div>
		{/each}
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
