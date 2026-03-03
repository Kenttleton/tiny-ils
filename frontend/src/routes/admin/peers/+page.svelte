<script lang="ts">
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let copiedId = $state(false);
	let copiedKey = $state(false);

	function copy(text: string, which: 'id' | 'key') {
		navigator.clipboard.writeText(text).then(() => {
			if (which === 'id') {
				copiedId = true;
				setTimeout(() => (copiedId = false), 1500);
			} else {
				copiedKey = true;
				setTimeout(() => (copiedKey = false), 1500);
			}
		});
	}

	const capLabel: Record<string, string> = { curios: 'Catalog', users: 'Users', ui: 'UI' };
	const capabilityLabel = (c: string) => capLabel[c] ?? c;
</script>

<svelte:head>
	<title>Network — Admin — tiny-ils</title>
</svelte:head>

<h1>Network</h1>

{#if form?.error}
	<p class="msg error">{form.error}</p>
{/if}
{#if form?.success}
	<p class="msg success">Done.</p>
{/if}

<section>
	<h2>This library</h2>
	<div class="identity-card">
		<div class="identity-row">
			<div class="identity-field">
				<span class="field-label">Library ID</span>
				<code class="field-value">{data.nodeId || '—'}</code>
			</div>
			{#if data.nodeId}
				<button class="btn-copy" onclick={() => copy(data.nodeId, 'id')}>
					{copiedId ? 'Copied!' : 'Copy'}
				</button>
			{/if}
		</div>
		<div class="identity-row">
			<div class="identity-field">
				<span class="field-label">Public key</span>
				<code class="field-value key-value">{data.publicKey || '—'}</code>
			</div>
			{#if data.publicKey}
				<button class="btn-copy" onclick={() => copy(data.publicKey, 'key')}>
					{copiedKey ? 'Copied!' : 'Copy'}
				</button>
			{/if}
		</div>
		{#if data.capabilities.length > 0}
			<div class="identity-row caps-row">
				<div class="identity-field">
					<span class="field-label">Capabilities</span>
					<div class="cap-pills">
						{#each data.capabilities as cap}
							<span class="cap-pill cap-{cap}">{capabilityLabel(cap)}</span>
						{/each}
					</div>
				</div>
			</div>
		{/if}
	</div>
</section>

<section>
	<h2>Partner libraries ({data.peers.length})</h2>
	{#if data.peers.length === 0}
		<p class="empty">No partner libraries connected yet.</p>
	{:else}
		<table>
			<thead>
				<tr>
					<th>Name</th>
					<th>Address</th>
					<th>Library ID</th>
					<th>Capabilities</th>
					<th>Status</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				{#each data.peers as peer}
					<tr>
						<td>{peer.displayName || '—'}</td>
						<td class="mono">{peer.address}</td>
						<td class="mono">{peer.nodeId}</td>
						<td>
							<div class="cap-pills">
								{#each (peer.capabilities ?? []) as cap}
									<span class="cap-pill cap-{cap}">{capabilityLabel(cap)}</span>
								{:else}
									<span class="cap-unknown">—</span>
								{/each}
							</div>
						</td>
						<td>
							<span class="status-badge status-{(peer.status ?? 'pending').toLowerCase()}">
								{peer.status ?? 'PENDING'}
							</span>
						</td>
						<td>
							{#if peer.status === 'PENDING'}
								<form method="POST" action="?/approve" style="display:inline">
									<input type="hidden" name="nodeId" value={peer.nodeId} />
									<button type="submit" class="btn-approve">Approve</button>
								</form>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</section>

<section class="connect-section">
	<h2>Connect a partner library</h2>
	<p class="desc">
		Share your Library ID and public key with the other library's administrator, and ask them to do
		the same. Enter their details below to establish the connection.
	</p>
	<form method="POST" action="?/connect" class="peer-form">
		<label>
			Display name
			<input type="text" name="displayName" value={form?.values?.displayName ?? ''} />
		</label>
		<label>
			Address (host:port) <span class="required">*</span>
			<input
				type="text"
				name="address"
				value={form?.values?.address ?? ''}
				placeholder="192.168.1.10:50153"
				required
			/>
		</label>
		<label>
			Library ID <span class="required">*</span>
			<input type="text" name="nodeId" value={form?.values?.nodeId ?? ''} required />
		</label>
		<label>
			Public key (base64) <span class="required">*</span>
			<textarea name="publicKey" rows="3" required>{form?.values?.publicKey ?? ''}</textarea>
		</label>
		<button type="submit" class="btn-primary">Connect library</button>
	</form>
</section>

<style>
	h1 { margin: 0 0 1.5rem; }
	h2 { font-size: 1rem; margin: 0 0 0.75rem; }
	section { margin-bottom: 2rem; }
	.desc { font-size: 0.8125rem; color: #6b7280; margin: 0 0 1rem; }
	.identity-card {
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 0.75rem 1rem;
		background: #f9fafb;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 640px;
	}
	.identity-row {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}
	.identity-field {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}
	.field-label {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #6b7280;
		font-weight: 600;
	}
	.field-value {
		font-family: monospace;
		font-size: 0.8125rem;
		word-break: break-all;
		color: #111;
	}
	.key-value { font-size: 0.75rem; }
	.btn-copy {
		flex-shrink: 0;
		padding: 0.25rem 0.6rem;
		background: #fff;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.75rem;
		cursor: pointer;
		color: #374151;
		margin-top: 1.1rem;
		min-width: 60px;
	}
	.btn-copy:hover { background: #f3f4f6; }
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	th {
		text-align: left;
		padding: 0.5rem;
		border-bottom: 2px solid #e5e7eb;
		color: #6b7280;
		font-weight: 500;
	}
	td { padding: 0.5rem; border-bottom: 1px solid #f3f4f6; vertical-align: middle; }
	.mono { font-family: monospace; font-size: 0.8rem; word-break: break-all; }
	.status-badge {
		display: inline-block;
		font-size: 0.7rem;
		font-weight: 600;
		letter-spacing: 0.04em;
		padding: 0.15rem 0.45rem;
		border-radius: 9999px;
	}
	.status-connected { background: #dcfce7; color: #166534; }
	.status-pending   { background: #fef9c3; color: #713f12; }
	.btn-approve {
		padding: 0.2rem 0.6rem;
		background: #fff;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.75rem;
		cursor: pointer;
		color: #374151;
	}
	.btn-approve:hover { background: #f3f4f6; }
	.peer-form { display: flex; flex-direction: column; gap: 0.75rem; max-width: 480px; }
	label { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.875rem; font-weight: 500; }
	.required { color: #dc2626; font-weight: 400; }
	input,
	textarea {
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
	.msg { font-size: 0.875rem; margin: 0 0 1.25rem; }
	.error { color: #dc2626; }
	.success { color: #16a34a; }
	.empty { color: #6b7280; font-size: 0.875rem; }
	.connect-section { padding-top: 1.5rem; border-top: 1px solid #e5e7eb; }
	.caps-row { margin-top: 0.25rem; }
	.cap-pills { display: flex; flex-wrap: wrap; gap: 0.3rem; }
	.cap-pill {
		font-size: 0.68rem;
		font-weight: 600;
		letter-spacing: 0.03em;
		padding: 0.1rem 0.45rem;
		border-radius: 9999px;
		background: #f3f4f6;
		color: #374151;
	}
	.cap-curios { background: #eff6ff; color: #1d4ed8; }
	.cap-users  { background: #f0fdf4; color: #166534; }
	.cap-ui     { background: #fdf4ff; color: #7e22ce; }
	.cap-unknown { font-size: 0.75rem; color: #9ca3af; }
</style>
