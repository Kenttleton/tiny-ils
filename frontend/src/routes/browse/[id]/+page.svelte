<script lang="ts">
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	const { curio, copies } = $derived(data);
	const availableCopies = $derived(copies.filter((c) => c.available));
</script>

<svelte:head>
	<title>{curio.title} — tiny-ils</title>
</svelte:head>

<a href="/browse" class="back">← Back to browse</a>

<div class="curio">
	<div class="header">
		<h1>{curio.title}</h1>
		<span class="badge">{curio.mediaType}</span>
		<span class="badge">{curio.formatType}</span>
	</div>

	{#if curio.description}
		<p class="desc">{curio.description}</p>
	{/if}

	{#if curio.tags?.length}
		<p class="tags">{curio.tags.join(' · ')}</p>
	{/if}
</div>

{#if form?.error}
	<p class="error">{form.error}</p>
{/if}
{#if form?.success}
	<p class="success">
		{form.action === 'checkout' ? 'Checked out successfully!' : 'Hold placed successfully!'}
	</p>
{/if}

<section class="copies">
	<h2>Physical copies ({copies.length})</h2>

	{#if copies.length === 0}
		<p class="empty">No physical copies registered.</p>
	{:else}
		<ul>
			{#each copies as copy (copy.id)}
				<li class="copy" class:unavailable={!copy.available}>
					<div>
						<strong>{copy.location || 'On shelf'}</strong>
						<span class="cond">{copy.condition}</span>
					</div>
					<div>
						{#if copy.available}
							<form method="POST" action="?/checkout">
								<input type="hidden" name="copyId" value={copy.id} />
								<button type="submit" class="btn-primary">Check out</button>
							</form>
						{:else}
							<span class="unavail-label">Checked out</span>
						{/if}
					</div>
				</li>
			{/each}
		</ul>

		{#if availableCopies.length === 0}
			<form method="POST" action="?/hold" class="hold-form">
				<p>All copies are checked out.</p>
				<button type="submit" class="btn-secondary">Place hold</button>
			</form>
		{/if}
	{/if}
</section>

<style>
	.back { display: inline-block; margin-bottom: 1.5rem; color: #6b7280; text-decoration: none; font-size: 0.875rem; }
	.back:hover { color: #111; }
	.curio { margin-bottom: 2rem; }
	.header { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 0.75rem; }
	h1 { margin: 0; font-size: 1.75rem; }
	.badge {
		font-size: 0.75rem;
		padding: 0.2rem 0.5rem;
		background: #f3f4f6;
		border-radius: 9999px;
		color: #374151;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.desc { color: #374151; line-height: 1.6; }
	.tags { color: #6b7280; font-size: 0.875rem; }
	.error { color: #dc2626; }
	.success { color: #16a34a; }
	h2 { font-size: 1.1rem; margin: 0 0 1rem; }
	ul { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.5rem; }
	.copy {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
	}
	.copy.unavailable { background: #f9fafb; }
	.cond { font-size: 0.75rem; color: #6b7280; margin-left: 0.5rem; }
	.unavail-label { font-size: 0.875rem; color: #9ca3af; }
	.btn-primary {
		padding: 0.35rem 0.75rem;
		background: #111;
		color: #fff;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.875rem;
	}
	.btn-secondary {
		padding: 0.35rem 0.75rem;
		background: #fff;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.875rem;
	}
	.hold-form { margin-top: 1rem; display: flex; align-items: center; gap: 1rem; }
	.hold-form p { margin: 0; color: #6b7280; font-size: 0.875rem; }
</style>
