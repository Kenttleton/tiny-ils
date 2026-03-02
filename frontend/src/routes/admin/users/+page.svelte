<script lang="ts">
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
	let newUserId = $state('');
</script>

<svelte:head>
	<title>Users — Admin — tiny-ils</title>
</svelte:head>

<h1>Users &amp; Claims</h1>

{#if form?.error}
	<p class="error">{form.error}</p>
{/if}
{#if form?.success}
	<p class="success">Done.</p>
{/if}

<section>
	<h2>Current manager claims</h2>
	{#if data.claims.length === 0}
		<p class="empty">No claims found.</p>
	{:else}
		<table>
			<thead>
				<tr>
					<th>User ID</th>
					<th>Node</th>
					<th>Role</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				{#each data.claims as claim}
					<tr>
						<td class="mono">{claim.userId}</td>
						<td class="mono">{claim.nodeId}</td>
						<td>{claim.role}</td>
						<td>
							<form method="POST" action="?/revoke" class="inline">
								<input type="hidden" name="userId" value={claim.userId} />
								<button
									type="submit"
									class="btn-danger"
									onclick={(e) => {
										if (!confirm('Revoke this claim?')) e.preventDefault();
									}}
								>
									Revoke
								</button>
							</form>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</section>

<section class="grant-section">
	<h2>Grant manager claim</h2>
	<form method="POST" action="?/grant" class="grant-form">
		<label>
			User ID
			<input type="text" name="userId" bind:value={newUserId} placeholder="UUID" required />
		</label>
		<button type="submit" class="btn-primary">Grant MANAGER</button>
	</form>
</section>

<style>
	h1 { margin: 0 0 1.5rem; }
	h2 { font-size: 1rem; margin: 0 0 0.75rem; }
	section { margin-bottom: 2rem; }
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	th { text-align: left; padding: 0.5rem; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-weight: 500; }
	td { padding: 0.5rem; border-bottom: 1px solid #f3f4f6; }
	.mono { font-family: monospace; font-size: 0.8rem; }
	.inline { display: inline; }
	.btn-danger {
		padding: 0.2rem 0.5rem;
		background: none;
		border: 1px solid #fca5a5;
		border-radius: 4px;
		color: #dc2626;
		cursor: pointer;
		font-size: 0.75rem;
	}
	.grant-form { display: flex; flex-direction: column; gap: 0.75rem; max-width: 400px; }
	label { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.875rem; font-weight: 500; }
	input { padding: 0.5rem 0.75rem; border: 1px solid #d1d5db; border-radius: 4px; font-size: 0.875rem; }
	.btn-primary {
		padding: 0.5rem 1rem;
		background: #111;
		color: #fff;
		border: none;
		border-radius: 4px;
		font-size: 0.875rem;
		cursor: pointer;
		align-self: flex-start;
	}
	.error { color: #dc2626; font-size: 0.875rem; }
	.success { color: #16a34a; font-size: 0.875rem; }
	.empty { color: #6b7280; font-size: 0.875rem; }
</style>
