<script lang="ts">
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();

	const MEDIA_TYPES = ['THING', 'BOOK', 'VIDEO', 'AUDIO', 'GAME'];
	const FORMAT_TYPES = ['PHYSICAL', 'DIGITAL', 'BOTH'];

	// Extract initial values from form so $state doesn't capture reactive props directly.
	const initialMediaType = form?.values?.mediaType ?? 'BOOK';
	const initialTitle = form?.enriched?.title ?? form?.values?.title ?? '';
	const initialDescription = form?.enriched?.description ?? form?.values?.description ?? '';
	const initialTags = form?.enriched?.tags?.join(', ') ?? form?.values?.tags?.join(', ') ?? '';

	let mediaType = $state(initialMediaType);
	let enrichIdentifier = $state('');
	let title = $state(initialTitle);
	let description = $state(initialDescription);
	let tags = $state(initialTags);
</script>

<svelte:head>
	<title>New Curio — Admin — tiny-ils</title>
</svelte:head>

<a href="/admin/curios" class="back">← Back to curios</a>
<h1>Add curio</h1>

<!-- Step 1: optional metadata enrichment -->
{#if mediaType !== 'THING'}
	<section class="enrich-section">
		<h2>Enrich metadata (optional)</h2>
		<form method="POST" action="?/enrich" class="enrich-form">
			<input type="hidden" name="mediaType" value={mediaType} />
			<input
				type="text"
				name="identifier"
				bind:value={enrichIdentifier}
				placeholder={mediaType === 'BOOK'
					? 'ISBN or title'
					: mediaType === 'VIDEO'
						? 'TMDB ID or title'
						: mediaType === 'AUDIO'
							? 'MusicBrainz ID or title'
							: 'IGDB ID or title'}
			/>
			<button type="submit">Look up</button>
		</form>
		{#if form?.error}
			<p class="error">{form.error}</p>
		{/if}
		{#if form?.enriched}
			<p class="enrich-ok">Metadata loaded — review and adjust below.</p>
		{/if}
	</section>
{/if}

<!-- Step 2: curio form -->
<form method="POST" action="?/create" class="curio-form">
	{#if form?.error && !form?.enriched}
		<p class="error">{form.error}</p>
	{/if}

	<label>
		Media type
		<select name="mediaType" bind:value={mediaType}>
			{#each MEDIA_TYPES as mt}
				<option value={mt}>{mt}</option>
			{/each}
		</select>
	</label>

	<label>
		Format type
		<select name="formatType">
			{#each FORMAT_TYPES as ft}
				<option value={ft} selected={ft === (form?.values?.formatType ?? 'PHYSICAL')}>{ft}</option>
			{/each}
		</select>
	</label>

	<label>
		Title *
		<input type="text" name="title" bind:value={title} required />
	</label>

	<label>
		Description
		<textarea name="description" rows="4" bind:value={description}></textarea>
	</label>

	<label>
		Tags (comma-separated)
		<input type="text" name="tags" bind:value={tags} placeholder="fiction, science, paperback" />
	</label>

	<label>
		Barcode
		<input type="text" name="barcode" value={form?.values?.barcode ?? ''} />
	</label>

	<button type="submit" class="btn-primary">Create curio</button>
</form>

<style>
	.back { display: inline-block; margin-bottom: 1rem; color: #6b7280; text-decoration: none; font-size: 0.875rem; }
	h1 { margin: 0 0 1.5rem; }
	h2 { font-size: 1rem; margin: 0 0 0.75rem; }
	.enrich-section {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		padding: 1rem;
		margin-bottom: 1.5rem;
	}
	.enrich-form { display: flex; gap: 0.5rem; }
	.enrich-form input { flex: 1; padding: 0.4rem 0.6rem; border: 1px solid #d1d5db; border-radius: 4px; font-size: 0.875rem; }
	.enrich-form button { padding: 0.4rem 0.75rem; border: 1px solid #d1d5db; border-radius: 4px; background: #fff; cursor: pointer; font-size: 0.875rem; }
	.enrich-ok { color: #16a34a; font-size: 0.875rem; margin: 0.5rem 0 0; }
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
