<script lang="ts">
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
	<title>Peers — Admin — tiny-ils</title>
</svelte:head>

<h1>Peer Nodes</h1>

{#if form?.error}
	<p class="error">{form.error}</p>
{/if}
{#if form?.success}
	<p class="success">Peer registered.</p>
{/if}

<section>
	<h2>Connected peers ({data.peers.length})</h2>
	{#if data.peers.length === 0}
		<p class="empty">No peers connected yet.</p>
	{:else}
		<table>
			<thead>
				<tr>
					<th>Display name</th>
					<th>Address</th>
					<th>Node ID</th>
				</tr>
			</thead>
			<tbody>
				{#each data.peers as peer}
					<tr>
						<td>{peer.displayName || '—'}</td>
						<td class="mono">{peer.address}</td>
						<td class="mono">{peer.nodeId}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</section>

<section class="register-section">
	<h2>Register a peer</h2>
	<form method="POST" action="?/register" class="peer-form">
		<label>
			Display name
			<input type="text" name="displayName" value={form?.values?.displayName ?? ''} />
		</label>
		<label>
			Address (host:port) *
			<input
				type="text"
				name="address"
				value={form?.values?.address ?? ''}
				placeholder="192.168.1.10:50053"
				required
			/>
		</label>
		<label>
			Node ID (public key fingerprint) *
			<input type="text" name="nodeId" value={form?.values?.nodeId ?? ''} required />
		</label>
		<label>
			Public key (base64) *
			<textarea name="publicKey" rows="3" required>{form?.values?.publicKey ?? ''}</textarea>
		</label>
		<button type="submit" class="btn-primary">Register peer</button>
	</form>
</section>

<style>
	h1 { margin: 0 0 1.5rem; }
	h2 { font-size: 1rem; margin: 0 0 0.75rem; }
	section { margin-bottom: 2rem; }
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	th { text-align: left; padding: 0.5rem; border-bottom: 2px solid #e5e7eb; color: #6b7280; font-weight: 500; }
	td { padding: 0.5rem; border-bottom: 1px solid #f3f4f6; }
	.mono { font-family: monospace; font-size: 0.8rem; word-break: break-all; }
	.peer-form { display: flex; flex-direction: column; gap: 0.75rem; max-width: 480px; }
	label { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.875rem; font-weight: 500; }
	input, textarea {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
		font-family: inherit;
	}
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
