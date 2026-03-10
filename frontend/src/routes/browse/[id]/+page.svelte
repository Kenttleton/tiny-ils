<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import * as Alert from '$lib/components/ui/alert';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faArrowLeft, faBookOpen, faClockRotateLeft } from '@fortawesome/free-solid-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	const { curio, copies } = $derived(data);
	const availableCopies = $derived(copies.filter((c) => c.status === 'AVAILABLE'));
</script>

<svelte:head>
	<title>{curio.title} — tiny-ils</title>
</svelte:head>

<a href="/browse" class="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"><FontAwesomeIcon icon={faArrowLeft} class="h-3.5 w-3.5" /> Back to browse</a>

<div class="mb-8">
	<div class="mb-3 flex flex-wrap items-center gap-3">
		<h1 class="m-0 text-2xl font-bold">{curio.title}</h1>
		<span class="rounded-full bg-muted px-2 py-0.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{curio.mediaType}</span>
		<span class="rounded-full bg-muted px-2 py-0.5 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{curio.formatType}</span>
	</div>

	{#if curio.description}
		<p class="leading-relaxed text-foreground">{curio.description}</p>
	{/if}

	{#if curio.tags?.length}
		<p class="text-sm text-muted-foreground">{curio.tags.join(' · ')}</p>
	{/if}
</div>

{#if form?.error}
	<Alert.Root variant="destructive" class="mb-4">
		<Alert.Description>{form.error}</Alert.Description>
	</Alert.Root>
{/if}
{#if form?.success}
	<p class="mb-4 text-sm text-green-600">
		{form.action === 'checkout' ? 'Checked out successfully!' : 'Hold placed successfully!'}
	</p>
{/if}

<section>
	<h2 class="mb-3 text-base font-semibold">Physical copies ({copies.length})</h2>

	{#if copies.length === 0}
		<p class="text-muted-foreground text-sm">No physical copies registered.</p>
	{:else}
		<ul class="flex flex-col gap-2 p-0" style="list-style:none">
			{#each copies as copy (copy.id)}
				<li class="flex items-center justify-between rounded-md border border-border px-4 py-3 {copy.status === 'AVAILABLE' ? '' : 'bg-muted/40'}">
					<div>
						<strong>{copy.location || 'On shelf'}</strong>
						<span class="ml-2 text-xs text-muted-foreground">{copy.condition}</span>
					</div>
					<div>
						{#if copy.status === 'AVAILABLE'}
							<form method="POST" action="?/checkout">
								<input type="hidden" name="copyId" value={copy.id} />
								<Button type="submit" size="sm"><FontAwesomeIcon icon={faBookOpen} class="mr-1.5 h-3.5 w-3.5" />Check out</Button>
							</form>
						{:else}
							<span class="text-sm text-muted-foreground">Checked out</span>
						{/if}
					</div>
				</li>
			{/each}
		</ul>

		{#if availableCopies.length === 0}
			<form method="POST" action="?/hold" class="mt-4 flex items-center gap-4">
				<p class="m-0 text-sm text-muted-foreground">All copies are checked out.</p>
				<Button type="submit" variant="outline"><FontAwesomeIcon icon={faClockRotateLeft} class="mr-1.5 h-3.5 w-3.5" />Place hold</Button>
			</form>
		{/if}
	{/if}
</section>
