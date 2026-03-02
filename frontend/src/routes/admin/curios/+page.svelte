<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	const { q: initialQ } = data;
	let q = $state(initialQ);
</script>

<svelte:head>
	<title>Curios — Admin — tiny-ils</title>
</svelte:head>

<div class="header-row">
	<h1>Curios ({data.total})</h1>
	<a href="/admin/curios/new" class="btn-primary">+ Add curio</a>
</div>

{#if form?.error}
	<p class="error">{form.error}</p>
{/if}

<form
	class="search-bar"
	onsubmit={(e) => {
		e.preventDefault();
		goto(`/admin/curios?q=${encodeURIComponent(q)}`, { replaceState: true });
	}}
>
	<input type="search" bind:value={q} placeholder="Search…" />
	<button type="submit">Search</button>
</form>

{#if data.curios.length === 0}
	<p class="empty">No curios found.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Title</th>
				<th>Type</th>
				<th>Format</th>
				<th>Tags</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each data.curios as c (c.id)}
				<tr>
					<td>{c.title}</td>
					<td>{c.mediaType}</td>
					<td>{c.formatType}</td>
					<td class="tags">{c.tags?.join(', ') ?? ''}</td>
					<td class="actions">
						<a href="/admin/curios/{c.id}/edit">Edit</a>
						<form method="POST" action="?/delete" class="inline-form">
							<input type="hidden" name="id" value={c.id} />
							<button
								type="submit"
								class="btn-danger"
								onclick={(e) => {
									if (!confirm(`Delete "${c.title}"?`)) e.preventDefault();
								}}
							>
								Delete
							</button>
						</form>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

<style>
	.header-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
	h1 { margin: 0; }
	.btn-primary {
		padding: 0.5rem 1rem;
		background: #111;
		color: #fff;
		border-radius: 4px;
		text-decoration: none;
		font-size: 0.875rem;
	}
	.search-bar { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
	.search-bar input {
		flex: 1;
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
	}
	.search-bar button {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		background: #fff;
		cursor: pointer;
		font-size: 0.875rem;
	}
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	th { text-align: left; padding: 0.5rem; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-weight: 500; }
	td { padding: 0.5rem; border-bottom: 1px solid #f3f4f6; vertical-align: middle; }
	.tags { color: #9ca3af; font-size: 0.75rem; max-width: 200px; }
	.actions { display: flex; gap: 0.5rem; align-items: center; white-space: nowrap; }
	.actions a { color: #374151; text-decoration: none; font-size: 0.875rem; }
	.actions a:hover { text-decoration: underline; }
	.inline-form { display: inline; }
	.btn-danger {
		padding: 0.2rem 0.5rem;
		background: none;
		border: 1px solid #fca5a5;
		border-radius: 4px;
		color: #dc2626;
		cursor: pointer;
		font-size: 0.75rem;
	}
	.error { color: #dc2626; }
	.empty { color: #6b7280; }
</style>
