<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';
	import type { CopyTransfer } from '$lib/api';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	const { nodeId } = data;
	let activeTab = $state(data.tab ?? 'incoming');
	let expandedIds = $state(new Set<string>());
	let showRequestForm = $state(false);

	function toggleExpand(id: string) {
		const next = new Set(expandedIds);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		expandedIds = next;
	}

	function fmt(unix: number | undefined) {
		if (!unix) return '—';
		return new Date(unix * 1000).toLocaleString();
	}

	function statusBadgeClass(s: string) {
		const map: Record<string, string> = {
			PENDING: 'badge-yellow',
			APPROVED: 'badge-blue',
			IN_TRANSIT: 'badge-purple',
			RECEIVED: 'badge-green',
			REJECTED: 'badge-red',
			CANCELLED: 'badge-gray',
		};
		return 'badge ' + (map[s] ?? 'badge-gray');
	}

	function typeBadgeClass(t: string) {
		const map: Record<string, string> = {
			ILL: 'badge-blue',
			RETURN: 'badge-yellow',
			PERMANENT: 'badge-purple',
		};
		return 'badge ' + (map[t] ?? 'badge-gray');
	}
</script>

<h1>Transfers</h1>

{#if form?.error}
	<p class="error">{form.error}</p>
{/if}

<div class="tabs">
	<button class:active={activeTab === 'incoming'} onclick={() => (activeTab = 'incoming')}>
		Incoming ({data.incoming.length})
	</button>
	<button class:active={activeTab === 'outgoing'} onclick={() => (activeTab = 'outgoing')}>
		Outgoing ({data.outgoing.length})
	</button>
	<button class:active={activeTab === 'history'} onclick={() => (activeTab = 'history')}>
		History ({data.history.length})
	</button>
	<button class="btn-secondary" onclick={() => (showRequestForm = !showRequestForm)}>
		{showRequestForm ? 'Hide form' : '+ Request transfer'}
	</button>
</div>

<!-- Request new transfer form -->
{#if showRequestForm}
	<form method="POST" action="?/request" use:enhance class="request-form">
		<h2>Request a new transfer</h2>
		<div class="form-row">
			<label>
				Copy ID
				<input name="copyId" required placeholder="UUID of the physical copy" />
			</label>
			<label>
				Type
				<select name="transferType">
					<option value="ILL">ILL (Inter-Library Loan)</option>
					<option value="RETURN">Return to home node</option>
					<option value="PERMANENT">Permanent transfer</option>
				</select>
			</label>
		</div>
		<div class="form-row">
			<label>
				Source node (holds the copy)
				<input name="sourceNode" placeholder="Leave blank for this node" />
			</label>
			<label>
				Destination node
				<input name="destNode" placeholder="Leave blank for this node" value={nodeId} />
			</label>
		</div>
		<label>
			Notes
			<textarea name="notes" rows="2"></textarea>
		</label>
		<button type="submit">Submit request</button>
	</form>
{/if}

<!-- ─── Incoming ──────────────────────────────────────────────────────────────── -->
{#if activeTab === 'incoming'}
	{#if data.incoming.length === 0}
		<p class="empty">No incoming transfers.</p>
	{:else}
		{#each data.incoming as t (t.id)}
			<div class="transfer-card">
				<div class="card-header" role="button" tabindex="0"
					onclick={() => toggleExpand(t.id)}
					onkeydown={(e) => e.key === 'Enter' && toggleExpand(t.id)}>
					<span class={typeBadgeClass(t.transferType)}>{t.transferType}</span>
					<span class="copy-id" title={t.copyId}>{t.copyId.slice(0, 8)}…</span>
					<span class="nodes">{t.sourceNode.slice(0, 8)}… → <strong>this node</strong></span>
					<span class={statusBadgeClass(t.status)}>{t.status}</span>
					<span class="ts">{fmt(t.requestedAt)}</span>
				</div>

				{#if expandedIds.has(t.id)}
					<div class="card-detail">
						<p><strong>Copy:</strong> {t.copyId}</p>
						<p><strong>Initiated by:</strong> {t.initiatedBy}</p>
						{#if t.notes}<p><strong>Notes:</strong> {t.notes}</p>{/if}
						<div class="audit">
							<div>PENDING → requested {fmt(t.requestedAt)}</div>
							{#if t.approvedAt}<div>APPROVED → {fmt(t.approvedAt)} by {t.approvedBy}</div>{/if}
							{#if t.shippedAt}<div>IN_TRANSIT → shipped {fmt(t.shippedAt)}</div>{/if}
							{#if t.receivedAt}<div>RECEIVED → confirmed {fmt(t.receivedAt)}</div>{/if}
						</div>
						<div class="actions">
							{#if t.status === 'PENDING'}
								<form method="POST" action="?/approve" use:enhance>
									<input type="hidden" name="id" value={t.id} />
									<button type="submit" class="btn-primary">Approve</button>
								</form>
								<form method="POST" action="?/reject" use:enhance>
									<input type="hidden" name="id" value={t.id} />
									<button type="submit" class="btn-danger">Reject</button>
								</form>
							{:else if t.status === 'IN_TRANSIT'}
								<form method="POST" action="?/receive" use:enhance>
									<input type="hidden" name="id" value={t.id} />
									<button type="submit" class="btn-primary">Confirm received</button>
								</form>
							{:else if t.status === 'APPROVED'}
								<p class="info">Waiting for source node to ship.</p>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		{/each}
	{/if}
{/if}

<!-- ─── Outgoing ──────────────────────────────────────────────────────────────── -->
{#if activeTab === 'outgoing'}
	{#if data.outgoing.length === 0}
		<p class="empty">No outgoing transfers.</p>
	{:else}
		{#each data.outgoing as t (t.id)}
			<div class="transfer-card">
				<div class="card-header" role="button" tabindex="0"
					onclick={() => toggleExpand(t.id)}
					onkeydown={(e) => e.key === 'Enter' && toggleExpand(t.id)}>
					<span class={typeBadgeClass(t.transferType)}>{t.transferType}</span>
					<span class="copy-id" title={t.copyId}>{t.copyId.slice(0, 8)}…</span>
					<span class="nodes"><strong>this node</strong> → {t.destNode.slice(0, 8)}…</span>
					<span class={statusBadgeClass(t.status)}>{t.status}</span>
					<span class="ts">{fmt(t.requestedAt)}</span>
				</div>

				{#if expandedIds.has(t.id)}
					<div class="card-detail">
						<p><strong>Copy:</strong> {t.copyId}</p>
						<p><strong>Initiated by:</strong> {t.initiatedBy}</p>
						{#if t.notes}<p><strong>Notes:</strong> {t.notes}</p>{/if}
						<div class="audit">
							<div>PENDING → requested {fmt(t.requestedAt)}</div>
							{#if t.approvedAt}<div>APPROVED → {fmt(t.approvedAt)}</div>{/if}
							{#if t.shippedAt}<div>IN_TRANSIT → shipped {fmt(t.shippedAt)}</div>{/if}
							{#if t.receivedAt}<div>RECEIVED → {fmt(t.receivedAt)}</div>{/if}
						</div>
						<div class="actions">
							{#if t.status === 'APPROVED'}
								<form method="POST" action="?/ship" use:enhance>
									<input type="hidden" name="id" value={t.id} />
									<button type="submit" class="btn-primary">Mark shipped</button>
								</form>
							{:else if t.status === 'PENDING'}
								<form method="POST" action="?/cancel" use:enhance>
									<input type="hidden" name="id" value={t.id} />
									<button type="submit" class="btn-danger">Cancel</button>
								</form>
							{:else if t.status === 'IN_TRANSIT'}
								<p class="info">Item in transit — waiting for destination to confirm.</p>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		{/each}
	{/if}
{/if}

<!-- ─── History ───────────────────────────────────────────────────────────────── -->
{#if activeTab === 'history'}
	{#if data.history.length === 0}
		<p class="empty">No completed transfers.</p>
	{:else}
		{#each data.history as t (t.id)}
			<div class="transfer-card">
				<div class="card-header" role="button" tabindex="0"
					onclick={() => toggleExpand(t.id)}
					onkeydown={(e) => e.key === 'Enter' && toggleExpand(t.id)}>
					<span class={typeBadgeClass(t.transferType)}>{t.transferType}</span>
					<span class="copy-id" title={t.copyId}>{t.copyId.slice(0, 8)}…</span>
					<span class="nodes">{t.sourceNode.slice(0, 8)}… → {t.destNode.slice(0, 8)}…</span>
					<span class={statusBadgeClass(t.status)}>{t.status}</span>
					<span class="ts">{fmt(t.receivedAt ?? t.requestedAt)}</span>
				</div>

				{#if expandedIds.has(t.id)}
					<div class="card-detail audit">
						<div>PENDING → requested {fmt(t.requestedAt)} by {t.initiatedBy.slice(0, 8)}…</div>
						{#if t.approvedAt}<div>APPROVED → {fmt(t.approvedAt)}</div>{/if}
						{#if t.shippedAt}<div>IN_TRANSIT → shipped {fmt(t.shippedAt)}</div>{/if}
						{#if t.receivedAt}<div>{t.status} → confirmed {fmt(t.receivedAt)}</div>{/if}
						{#if t.notes}<div>Notes: {t.notes}</div>{/if}
					</div>
				{/if}
			</div>
		{/each}
	{/if}
{/if}

<style>
	h1 { margin-bottom: 1rem; }
	.error { color: #dc2626; margin-bottom: 1rem; }
	.tabs {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
	}
	.tabs button {
		padding: 0.4rem 0.9rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		background: white;
		cursor: pointer;
		font-size: 0.875rem;
	}
	.tabs button.active {
		background: #1d4ed8;
		color: white;
		border-color: #1d4ed8;
	}
	.tabs .btn-secondary {
		margin-left: auto;
		background: #f9fafb;
	}
	.request-form {
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		padding: 1.25rem;
		margin-bottom: 1.5rem;
	}
	.request-form h2 { margin: 0 0 1rem; font-size: 1rem; }
	.form-row { display: flex; gap: 1rem; margin-bottom: 0.75rem; flex-wrap: wrap; }
	.form-row label { flex: 1; min-width: 180px; }
	label { display: flex; flex-direction: column; gap: 0.25rem; font-size: 0.875rem; margin-bottom: 0.75rem; }
	input, select, textarea {
		padding: 0.35rem 0.5rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
	}
	.empty { color: #6b7280; font-style: italic; padding: 1rem 0; }
	.transfer-card {
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		margin-bottom: 0.75rem;
		overflow: hidden;
	}
	.card-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		background: #f9fafb;
		cursor: pointer;
		flex-wrap: wrap;
	}
	.card-header:hover { background: #f3f4f6; }
	.copy-id { font-family: monospace; font-size: 0.8rem; color: #6b7280; }
	.nodes { font-size: 0.85rem; flex: 1; }
	.ts { font-size: 0.75rem; color: #9ca3af; margin-left: auto; white-space: nowrap; }
	.card-detail {
		padding: 1rem;
		border-top: 1px solid #e5e7eb;
		font-size: 0.875rem;
	}
	.card-detail p { margin: 0.25rem 0; }
	.audit {
		background: #f9fafb;
		border-radius: 4px;
		padding: 0.75rem;
		margin: 0.75rem 0;
		font-family: monospace;
		font-size: 0.8rem;
		line-height: 1.8;
	}
	.audit div::before { content: '▸ '; color: #9ca3af; }
	.actions { display: flex; gap: 0.5rem; flex-wrap: wrap; }
	.info { color: #6b7280; font-style: italic; margin: 0; }
	.badge {
		display: inline-block;
		padding: 0.15rem 0.5rem;
		border-radius: 999px;
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		white-space: nowrap;
	}
	.badge-yellow { background: #fef9c3; color: #92400e; }
	.badge-blue   { background: #dbeafe; color: #1e40af; }
	.badge-purple { background: #ede9fe; color: #5b21b6; }
	.badge-green  { background: #dcfce7; color: #14532d; }
	.badge-red    { background: #fee2e2; color: #991b1b; }
	.badge-gray   { background: #f3f4f6; color: #374151; }
	button[type="submit"], .btn-primary {
		padding: 0.35rem 0.9rem;
		border: none;
		border-radius: 4px;
		background: #1d4ed8;
		color: white;
		cursor: pointer;
		font-size: 0.875rem;
	}
	.btn-danger {
		padding: 0.35rem 0.9rem;
		border: none;
		border-radius: 4px;
		background: #dc2626;
		color: white;
		cursor: pointer;
		font-size: 0.875rem;
	}
</style>
