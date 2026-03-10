<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData, ActionData } from './$types';
	import * as Tabs from '$lib/components/ui/tabs';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faCheck, faXmark, faTruck, faCircleCheck, faPlus } from '@fortawesome/free-solid-svg-icons';

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

	function statusClass(s: string): string {
		const map: Record<string, string> = {
			PENDING: 'bg-yellow-100 text-yellow-800',
			APPROVED: 'bg-blue-100 text-blue-800',
			IN_TRANSIT: 'bg-purple-100 text-purple-800',
			RECEIVED: 'bg-green-100 text-green-800',
			REJECTED: 'bg-red-100 text-red-800',
			CANCELLED: 'bg-zinc-100 text-zinc-700',
		};
		return map[s] ?? 'bg-zinc-100 text-zinc-700';
	}

	function typeClass(t: string): string {
		const map: Record<string, string> = {
			ILL: 'bg-blue-100 text-blue-800',
			RETURN: 'bg-yellow-100 text-yellow-800',
			PERMANENT: 'bg-purple-100 text-purple-800',
		};
		return map[t] ?? 'bg-zinc-100 text-zinc-700';
	}
</script>

<h1 class="mb-4 text-2xl font-bold">Transfers</h1>

{#if form?.error}
	<Alert.Root variant="destructive" class="mb-4">
		<Alert.Description>{form.error}</Alert.Description>
	</Alert.Root>
{/if}

<div class="mb-6 flex flex-wrap items-center gap-2">
	<Tabs.Root bind:value={activeTab} class="w-full">
		<div class="flex flex-wrap items-center gap-2">
			<Tabs.List>
				<Tabs.Trigger value="incoming">Incoming ({data.incoming.length})</Tabs.Trigger>
				<Tabs.Trigger value="outgoing">Outgoing ({data.outgoing.length})</Tabs.Trigger>
				<Tabs.Trigger value="history">History ({data.history.length})</Tabs.Trigger>
			</Tabs.List>
			<Button variant="outline" size="sm" onclick={() => (showRequestForm = !showRequestForm)} class="ml-auto">
				{#if !showRequestForm}<FontAwesomeIcon icon={faPlus} class="mr-1.5 h-3.5 w-3.5" />{/if}{showRequestForm ? 'Hide form' : 'Request transfer'}
			</Button>
		</div>

		<!-- Request new transfer form -->
		{#if showRequestForm}
			<form method="POST" action="?/request" use:enhance class="my-4 rounded-md border border-border p-5">
				<h2 class="mb-4 text-base font-semibold">Request a new transfer</h2>
				<div class="mb-3 flex flex-wrap gap-4">
					<div class="flex min-w-[180px] flex-1 flex-col gap-1.5">
						<Label for="copyId">Copy ID</Label>
						<Input id="copyId" name="copyId" required placeholder="UUID of the physical copy" />
					</div>
					<div class="flex min-w-[180px] flex-1 flex-col gap-1.5">
						<Label for="transferType">Type</Label>
						<select
							id="transferType"
							name="transferType"
							class="rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
						>
							<option value="ILL">ILL (Inter-Library Loan)</option>
							<option value="RETURN">Return to home node</option>
							<option value="PERMANENT">Permanent transfer</option>
						</select>
					</div>
				</div>
				<div class="mb-3 flex flex-wrap gap-4">
					<div class="flex min-w-[180px] flex-1 flex-col gap-1.5">
						<Label for="sourceNode">Source node (holds the copy)</Label>
						<Input id="sourceNode" name="sourceNode" placeholder="Leave blank for this node" />
					</div>
					<div class="flex min-w-[180px] flex-1 flex-col gap-1.5">
						<Label for="destNode">Destination node</Label>
						<Input id="destNode" name="destNode" placeholder="Leave blank for this node" value={nodeId} />
					</div>
				</div>
				<div class="mb-4 flex flex-col gap-1.5">
					<Label for="notes">Notes</Label>
					<textarea
						id="notes"
						name="notes"
						rows="2"
						class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
					></textarea>
				</div>
				<Button type="submit">Submit request</Button>
			</form>
		{/if}

		<!-- ─── Incoming ──────────────────────────────────────────────────────────────── -->
		<Tabs.Content value="incoming">
			{#if data.incoming.length === 0}
				<p class="text-muted-foreground text-sm italic py-4">No incoming transfers.</p>
			{:else}
				{#each data.incoming as t (t.id)}
					<div class="mb-3 overflow-hidden rounded-md border border-border">
						<div
							class="flex cursor-pointer flex-wrap items-center gap-3 bg-muted/40 px-4 py-3 hover:bg-muted/60"
							role="button"
							tabindex="0"
							onclick={() => toggleExpand(t.id)}
							onkeydown={(e) => e.key === 'Enter' && toggleExpand(t.id)}
						>
							<span class="rounded-full px-2 py-0.5 text-xs font-semibold uppercase tracking-wide {typeClass(t.transferType)}">{t.transferType}</span>
							<span class="font-mono text-xs text-muted-foreground" title={t.copyId}>{t.copyId.slice(0, 8)}…</span>
							<span class="flex-1 text-sm">{t.sourceNode.slice(0, 8)}… → <strong>this node</strong></span>
							<span class="rounded-full px-2 py-0.5 text-xs font-semibold uppercase tracking-wide {statusClass(t.status)}">{t.status}</span>
							<span class="ml-auto whitespace-nowrap text-xs text-muted-foreground">{fmt(t.requestedAt)}</span>
						</div>

						{#if expandedIds.has(t.id)}
							<div class="border-t border-border p-4 text-sm">
								<p class="my-1"><strong>Copy:</strong> {t.copyId}</p>
								<p class="my-1"><strong>Initiated by:</strong> {t.initiatedBy}</p>
								{#if t.notes}<p class="my-1"><strong>Notes:</strong> {t.notes}</p>{/if}
								<div class="my-3 rounded bg-muted/40 p-3 font-mono text-xs leading-relaxed">
									<div>▸ PENDING → requested {fmt(t.requestedAt)}</div>
									{#if t.approvedAt}<div>▸ APPROVED → {fmt(t.approvedAt)} by {t.approvedBy}</div>{/if}
									{#if t.shippedAt}<div>▸ IN_TRANSIT → shipped {fmt(t.shippedAt)}</div>{/if}
									{#if t.receivedAt}<div>▸ RECEIVED → confirmed {fmt(t.receivedAt)}</div>{/if}
								</div>
								<div class="flex flex-wrap gap-2">
									{#if t.status === 'PENDING'}
										<form method="POST" action="?/approve" use:enhance>
											<input type="hidden" name="id" value={t.id} />
											<Button type="submit" size="sm"><FontAwesomeIcon icon={faCheck} class="mr-1.5 h-3.5 w-3.5" />Approve</Button>
										</form>
										<form method="POST" action="?/reject" use:enhance>
											<input type="hidden" name="id" value={t.id} />
											<Button type="submit" size="sm" variant="destructive"><FontAwesomeIcon icon={faXmark} class="mr-1.5 h-3.5 w-3.5" />Reject</Button>
										</form>
									{:else if t.status === 'IN_TRANSIT'}
										<form method="POST" action="?/receive" use:enhance>
											<input type="hidden" name="id" value={t.id} />
											<Button type="submit" size="sm"><FontAwesomeIcon icon={faCircleCheck} class="mr-1.5 h-3.5 w-3.5" />Confirm received</Button>
										</form>
									{:else if t.status === 'APPROVED'}
										<p class="m-0 text-sm italic text-muted-foreground">Waiting for source node to ship.</p>
									{/if}
								</div>
							</div>
						{/if}
					</div>
				{/each}
			{/if}
		</Tabs.Content>

		<!-- ─── Outgoing ──────────────────────────────────────────────────────────────── -->
		<Tabs.Content value="outgoing">
			{#if data.outgoing.length === 0}
				<p class="text-muted-foreground text-sm italic py-4">No outgoing transfers.</p>
			{:else}
				{#each data.outgoing as t (t.id)}
					<div class="mb-3 overflow-hidden rounded-md border border-border">
						<div
							class="flex cursor-pointer flex-wrap items-center gap-3 bg-muted/40 px-4 py-3 hover:bg-muted/60"
							role="button"
							tabindex="0"
							onclick={() => toggleExpand(t.id)}
							onkeydown={(e) => e.key === 'Enter' && toggleExpand(t.id)}
						>
							<span class="rounded-full px-2 py-0.5 text-xs font-semibold uppercase tracking-wide {typeClass(t.transferType)}">{t.transferType}</span>
							<span class="font-mono text-xs text-muted-foreground" title={t.copyId}>{t.copyId.slice(0, 8)}…</span>
							<span class="flex-1 text-sm"><strong>this node</strong> → {t.destNode.slice(0, 8)}…</span>
							<span class="rounded-full px-2 py-0.5 text-xs font-semibold uppercase tracking-wide {statusClass(t.status)}">{t.status}</span>
							<span class="ml-auto whitespace-nowrap text-xs text-muted-foreground">{fmt(t.requestedAt)}</span>
						</div>

						{#if expandedIds.has(t.id)}
							<div class="border-t border-border p-4 text-sm">
								<p class="my-1"><strong>Copy:</strong> {t.copyId}</p>
								<p class="my-1"><strong>Initiated by:</strong> {t.initiatedBy}</p>
								{#if t.notes}<p class="my-1"><strong>Notes:</strong> {t.notes}</p>{/if}
								<div class="my-3 rounded bg-muted/40 p-3 font-mono text-xs leading-relaxed">
									<div>▸ PENDING → requested {fmt(t.requestedAt)}</div>
									{#if t.approvedAt}<div>▸ APPROVED → {fmt(t.approvedAt)}</div>{/if}
									{#if t.shippedAt}<div>▸ IN_TRANSIT → shipped {fmt(t.shippedAt)}</div>{/if}
									{#if t.receivedAt}<div>▸ RECEIVED → {fmt(t.receivedAt)}</div>{/if}
								</div>
								<div class="flex flex-wrap gap-2">
									{#if t.status === 'APPROVED'}
										<form method="POST" action="?/ship" use:enhance>
											<input type="hidden" name="id" value={t.id} />
											<Button type="submit" size="sm"><FontAwesomeIcon icon={faTruck} class="mr-1.5 h-3.5 w-3.5" />Mark shipped</Button>
										</form>
									{:else if t.status === 'PENDING'}
										<form method="POST" action="?/cancel" use:enhance>
											<input type="hidden" name="id" value={t.id} />
											<Button type="submit" size="sm" variant="destructive">Cancel</Button>
										</form>
									{:else if t.status === 'IN_TRANSIT'}
										<p class="m-0 text-sm italic text-muted-foreground">Item in transit — waiting for destination to confirm.</p>
									{/if}
								</div>
							</div>
						{/if}
					</div>
				{/each}
			{/if}
		</Tabs.Content>

		<!-- ─── History ───────────────────────────────────────────────────────────────── -->
		<Tabs.Content value="history">
			{#if data.history.length === 0}
				<p class="text-muted-foreground text-sm italic py-4">No completed transfers.</p>
			{:else}
				{#each data.history as t (t.id)}
					<div class="mb-3 overflow-hidden rounded-md border border-border">
						<div
							class="flex cursor-pointer flex-wrap items-center gap-3 bg-muted/40 px-4 py-3 hover:bg-muted/60"
							role="button"
							tabindex="0"
							onclick={() => toggleExpand(t.id)}
							onkeydown={(e) => e.key === 'Enter' && toggleExpand(t.id)}
						>
							<span class="rounded-full px-2 py-0.5 text-xs font-semibold uppercase tracking-wide {typeClass(t.transferType)}">{t.transferType}</span>
							<span class="font-mono text-xs text-muted-foreground" title={t.copyId}>{t.copyId.slice(0, 8)}…</span>
							<span class="flex-1 text-sm">{t.sourceNode.slice(0, 8)}… → {t.destNode.slice(0, 8)}…</span>
							<span class="rounded-full px-2 py-0.5 text-xs font-semibold uppercase tracking-wide {statusClass(t.status)}">{t.status}</span>
							<span class="ml-auto whitespace-nowrap text-xs text-muted-foreground">{fmt(t.receivedAt ?? t.requestedAt)}</span>
						</div>

						{#if expandedIds.has(t.id)}
							<div class="border-t border-border p-4 font-mono text-xs leading-relaxed">
								<div>▸ PENDING → requested {fmt(t.requestedAt)} by {t.initiatedBy.slice(0, 8)}…</div>
								{#if t.approvedAt}<div>▸ APPROVED → {fmt(t.approvedAt)}</div>{/if}
								{#if t.shippedAt}<div>▸ IN_TRANSIT → shipped {fmt(t.shippedAt)}</div>{/if}
								{#if t.receivedAt}<div>▸ {t.status} → confirmed {fmt(t.receivedAt)}</div>{/if}
								{#if t.notes}<div>▸ Notes: {t.notes}</div>{/if}
							</div>
						{/if}
					</div>
				{/each}
			{/if}
		</Tabs.Content>
	</Tabs.Root>
</div>
