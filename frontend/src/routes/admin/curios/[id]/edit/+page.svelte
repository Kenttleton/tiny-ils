<script lang="ts">
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	const { curio, copies } = $derived(data);

	const MEDIA_TYPES = ['THING', 'BOOK', 'VIDEO', 'AUDIO', 'GAME'];
	const FORMAT_TYPES = ['PHYSICAL', 'DIGITAL', 'BOTH'];
	const CONDITIONS = ['NEW', 'GOOD', 'FAIR', 'POOR'];

	const statusColor: Record<string, string> = {
		AVAILABLE: 'green',
		ON_LOAN: 'yellow',
		REQUESTED: 'blue',
		IN_TRANSIT: 'orange'
	};
</script>

<svelte:head>
	<title>Edit {curio.title} — Admin — tiny-ils</title>
</svelte:head>

<a href="/admin/curios" class="back">← Back to curios</a>
<h1>Edit curio</h1>

{#if form?.error}
	<p class="msg error">{form.error}</p>
{/if}

<form method="POST" action="?/update" class="curio-form">
	<label>
		Media type
		<select name="mediaType">
			{#each MEDIA_TYPES as mt}
				<option value={mt} selected={mt === curio.mediaType}>{mt}</option>
			{/each}
		</select>
	</label>

	<label>
		Format type
		<select name="formatType">
			{#each FORMAT_TYPES as ft}
				<option value={ft} selected={ft === curio.formatType}>{ft}</option>
			{/each}
		</select>
	</label>

	<label>
		Title *
		<input type="text" name="title" value={curio.title} required />
	</label>

	<label>
		Description
		<textarea name="description" rows="4">{curio.description ?? ''}</textarea>
	</label>

	<label>
		Tags (comma-separated)
		<input type="text" name="tags" value={curio.tags?.join(', ') ?? ''} />
	</label>

	<label>
		Barcode
		<input type="text" name="barcode" value={curio.barcode ?? ''} />
	</label>

	<button type="submit" class="btn-primary">Save changes</button>
</form>

{#if curio.formatType === 'PHYSICAL' || curio.formatType === 'BOTH'}
<section class="copies-section">
	<h2>Physical copies ({copies.length})</h2>

	{#if form?.copyError}
		<p class="msg error">{form.copyError}</p>
	{/if}

	{#if copies.length > 0}
		<table>
			<thead>
				<tr>
					<th>Condition</th>
					<th>Location</th>
					<th>Status</th>
					<th>ID</th>
				</tr>
			</thead>
			<tbody>
				{#each copies as copy}
					<tr>
						<td>{copy.condition}</td>
						<td>{copy.location || '—'}</td>
						<td>
							<span class="badge badge-{statusColor[copy.status] ?? 'gray'}">
								{copy.status}
							</span>
						</td>
						<td class="mono">{copy.id.slice(0, 8)}…</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p class="empty">No physical copies yet.</p>
	{/if}

	<form method="POST" action="?/addCopy" class="add-copy-form">
		<h3>Add a copy</h3>
		<div class="add-copy-row">
			<label>
				Condition
				<select name="condition">
					{#each CONDITIONS as c}
						<option value={c} selected={c === 'GOOD'}>{c}</option>
					{/each}
				</select>
			</label>
			<label>
				Location
				<input type="text" name="location" placeholder="e.g. Shelf A3" />
			</label>
			<button type="submit" class="btn-secondary">Add copy</button>
		</div>
	</form>
</section>
{/if}

<style>
	.back { display: inline-block; margin-bottom: 1rem; color: #6b7280; text-decoration: none; font-size: 0.875rem; }
	h1 { margin: 0 0 1.5rem; }
	h2 { font-size: 1rem; margin: 0 0 0.75rem; }
	h3 { font-size: 0.875rem; margin: 1rem 0 0.5rem; }
	.curio-form { display: flex; flex-direction: column; gap: 1rem; max-width: 600px; }
	label { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.875rem; font-weight: 500; }
	input, select, textarea {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 1rem;
		font-family: inherit;
	}
	.btn-primary {
		padding: 0.6rem 1.25rem;
		background: #111;
		color: #fff;
		border: none;
		border-radius: 4px;
		font-size: 1rem;
		cursor: pointer;
		align-self: flex-start;
	}
	.btn-secondary {
		padding: 0.5rem 1rem;
		background: #fff;
		color: #111;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
		cursor: pointer;
		align-self: flex-end;
	}
	.btn-secondary:hover { background: #f3f4f6; }
	.msg { font-size: 0.875rem; margin: 0 0 0.75rem; }
	.error { color: #dc2626; }
	.copies-section { margin-top: 2.5rem; padding-top: 1.5rem; border-top: 1px solid #e5e7eb; max-width: 600px; }
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; margin-bottom: 1rem; }
	th { text-align: left; padding: 0.4rem 0.5rem; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-weight: 500; }
	td { padding: 0.4rem 0.5rem; border-bottom: 1px solid #f3f4f6; }
	.mono { font-family: monospace; font-size: 0.8rem; color: #6b7280; }
	.badge {
		display: inline-block;
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.03em;
		padding: 0.1rem 0.4rem;
		border-radius: 9999px;
	}
	.badge-green  { background: #dcfce7; color: #166534; }
	.badge-yellow { background: #fef9c3; color: #713f12; }
	.badge-blue   { background: #dbeafe; color: #1e40af; }
	.badge-orange { background: #ffedd5; color: #9a3412; }
	.badge-gray   { background: #f3f4f6; color: #374151; }
	.empty { color: #6b7280; font-size: 0.875rem; }
	.add-copy-form { margin-top: 0.5rem; }
	.add-copy-row { display: flex; gap: 0.75rem; align-items: flex-end; }
	.add-copy-row label { flex: 1; }
</style>
