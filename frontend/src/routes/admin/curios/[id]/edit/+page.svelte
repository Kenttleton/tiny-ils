<script lang="ts">
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	const { curio } = $derived(data);

	const MEDIA_TYPES = ['THING', 'BOOK', 'VIDEO', 'AUDIO', 'GAME'];
	const FORMAT_TYPES = ['PHYSICAL', 'DIGITAL', 'BOTH'];
</script>

<svelte:head>
	<title>Edit {curio.title} — Admin — tiny-ils</title>
</svelte:head>

<a href="/admin/curios" class="back">← Back to curios</a>
<h1>Edit curio</h1>

{#if form?.error}
	<p class="error">{form.error}</p>
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

<style>
	.back { display: inline-block; margin-bottom: 1rem; color: #6b7280; text-decoration: none; font-size: 0.875rem; }
	h1 { margin: 0 0 1.5rem; }
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
	.error { color: #dc2626; font-size: 0.875rem; margin: 0; }
</style>
