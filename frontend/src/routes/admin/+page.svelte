<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Admin — tiny-ils</title>
</svelte:head>

<h1>Dashboard</h1>

<div class="stats">
	<div class="stat">
		<span class="num">{data.total}</span>
		<span class="label">Curios in catalog</span>
	</div>
</div>

<section>
	<h2>Recent curios</h2>
	{#if data.recentCurios.length === 0}
		<p class="empty">No curios yet. <a href="/admin/curios/new">Add one</a>.</p>
	{:else}
		<table>
			<thead>
				<tr>
					<th>Title</th>
					<th>Type</th>
					<th>Format</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				{#each data.recentCurios as c (c.id)}
					<tr>
						<td>{c.title}</td>
						<td>{c.mediaType}</td>
						<td>{c.formatType}</td>
						<td><a href="/admin/curios/{c.id}/edit">Edit</a></td>
					</tr>
				{/each}
			</tbody>
		</table>
		<p class="more"><a href="/admin/curios">View all curios →</a></p>
	{/if}
</section>

<style>
	h1 { margin: 0 0 1.5rem; }
	.stats { display: flex; gap: 1rem; margin-bottom: 2rem; }
	.stat {
		display: flex;
		flex-direction: column;
		padding: 1rem 1.5rem;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		min-width: 140px;
	}
	.num { font-size: 2rem; font-weight: 700; }
	.label { font-size: 0.75rem; color: #6b7280; }
	h2 { font-size: 1rem; margin: 0 0 0.75rem; }
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	th { text-align: left; padding: 0.5rem; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-weight: 500; }
	td { padding: 0.5rem; border-bottom: 1px solid #f3f4f6; }
	td a { color: #374151; text-decoration: none; }
	td a:hover { color: #111; text-decoration: underline; }
	.empty { color: #6b7280; }
	.more { margin-top: 0.75rem; font-size: 0.875rem; }
</style>
