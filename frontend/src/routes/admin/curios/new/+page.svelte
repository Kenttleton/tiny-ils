<script lang="ts">
	import type { ActionData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faArrowLeft, faMagnifyingGlass, faPlus } from '@fortawesome/free-solid-svg-icons';

	let { form }: { form: ActionData } = $props();

	const MEDIA_TYPES = ['THING', 'BOOK', 'VIDEO', 'AUDIO', 'GAME'];
	const FORMAT_TYPES = ['PHYSICAL', 'DIGITAL', 'BOTH'];

	// Extract initial values from form so $state doesn't capture reactive props directly.
	const initialMediaType = form?.values?.media_type ?? 'BOOK';
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

<a href="/admin/curios" class="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"><FontAwesomeIcon icon={faArrowLeft} class="h-3.5 w-3.5" /> Back to curios</a>
<h1 class="mb-6 text-2xl font-bold">Add curio</h1>

<!-- Step 1: optional metadata enrichment -->
{#if mediaType !== 'THING'}
	<section class="mb-6 rounded-md border border-border bg-muted/40 p-4">
		<h2 class="mb-3 text-base font-semibold">Enrich metadata (optional)</h2>
		<form method="POST" action="?/enrich" class="flex gap-2">
			<input type="hidden" name="mediaType" value={mediaType} />
			<Input
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
				class="flex-1"
			/>
			<Button type="submit" variant="outline"><FontAwesomeIcon icon={faMagnifyingGlass} class="mr-1.5 h-3.5 w-3.5" />Look up</Button>
		</form>
		{#if form?.error}
			<p class="mt-2 text-sm text-destructive">{form.error}</p>
		{/if}
		{#if form?.enriched}
			<p class="mt-2 text-sm text-green-600">Metadata loaded — review and adjust below.</p>
		{/if}
	</section>
{/if}

<!-- Step 2: curio form -->
<form method="POST" action="?/create" class="flex max-w-[600px] flex-col gap-4">
	{#if form?.error && !form?.enriched}
		<Alert.Root variant="destructive">
			<Alert.Description>{form.error}</Alert.Description>
		</Alert.Root>
	{/if}

	<div class="flex flex-col gap-1.5">
		<Label for="mediaType">Media type</Label>
		<select
			id="mediaType"
			name="mediaType"
			bind:value={mediaType}
			class="rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
		>
			{#each MEDIA_TYPES as mt}
				<option value={mt}>{mt}</option>
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
				<option value={ft} selected={ft === (form?.values?.format_type ?? 'PHYSICAL')}>{ft}</option>
			{/each}
		</select>
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="title">Title *</Label>
		<Input id="title" type="text" name="title" bind:value={title} required />
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="description">Description</Label>
		<textarea
			id="description"
			name="description"
			rows="4"
			bind:value={description}
			class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
		></textarea>
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="tags">Tags (comma-separated)</Label>
		<Input id="tags" type="text" name="tags" bind:value={tags} placeholder="fiction, science, paperback" />
	</div>

	<div class="flex flex-col gap-1.5">
		<Label for="barcode">Barcode</Label>
		<Input id="barcode" type="text" name="barcode" value={form?.values?.barcode ?? ''} />
	</div>

	<div>
		<Button type="submit"><FontAwesomeIcon icon={faPlus} class="mr-1.5 h-3.5 w-3.5" />Create curio</Button>
	</div>
</form>
