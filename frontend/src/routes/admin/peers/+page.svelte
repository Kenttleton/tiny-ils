<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';
	import * as Table from '$lib/components/ui/table';
	import { Separator } from '$lib/components/ui/separator';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faCopy, faLink } from '@fortawesome/free-solid-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let copiedId = $state(false);
	let copiedKey = $state(false);
	let copiedAddr = $state(false);

	function copy(text: string, which: 'id' | 'key' | 'addr') {
		navigator.clipboard.writeText(text).then(() => {
			if (which === 'id') {
				copiedId = true;
				setTimeout(() => (copiedId = false), 1500);
			} else if (which === 'key') {
				copiedKey = true;
				setTimeout(() => (copiedKey = false), 1500);
			} else {
				copiedAddr = true;
				setTimeout(() => (copiedAddr = false), 1500);
			}
		});
	}

	const capLabel: Record<string, string> = { curios: 'Catalog', users: 'Users', ui: 'UI' };
	const capabilityLabel = (c: string) => capLabel[c] ?? c;

	function capClass(cap: string): string {
		const map: Record<string, string> = {
			curios: 'bg-blue-50 text-blue-800',
			users: 'bg-green-50 text-green-800',
			ui: 'bg-purple-50 text-purple-800',
		};
		return map[cap] ?? 'bg-zinc-100 text-zinc-700';
	}
</script>

<svelte:head>
	<title>Network — Admin — tiny-ils</title>
</svelte:head>

<h1 class="mb-6 text-2xl font-bold">Network</h1>

{#if form?.error}
	<Alert.Root variant="destructive" class="mb-4">
		<Alert.Description>{form.error}</Alert.Description>
	</Alert.Root>
{/if}
{#if form?.success}
	<p class="mb-5 text-sm text-green-600">Done.</p>
{/if}

<section class="mb-8">
	<h2 class="mb-3 text-base font-semibold">This library</h2>
	<div class="flex max-w-[640px] flex-col gap-3 rounded-md border border-border bg-muted/40 px-4 py-3">
		<div class="flex items-start gap-3">
			<div class="flex min-w-0 flex-1 flex-col gap-0.5">
				<span class="text-[0.7rem] font-semibold uppercase tracking-wide text-muted-foreground">Library ID</span>
				<code class="break-all font-mono text-[0.8125rem]">{data.nodeId || '—'}</code>
			</div>
			{#if data.nodeId}
				<Button variant="outline" size="sm" class="mt-4 shrink-0" onclick={() => copy(data.nodeId, 'id')}>
					<FontAwesomeIcon icon={faCopy} class="mr-1.5 h-3.5 w-3.5" />{copiedId ? 'Copied!' : 'Copy'}
				</Button>
			{/if}
		</div>
		<div class="flex items-start gap-3">
			<div class="flex min-w-0 flex-1 flex-col gap-0.5">
				<span class="text-[0.7rem] font-semibold uppercase tracking-wide text-muted-foreground">Public key</span>
				<code class="break-all font-mono text-xs">{data.publicKey || '—'}</code>
			</div>
			{#if data.publicKey}
				<Button variant="outline" size="sm" class="mt-4 shrink-0" onclick={() => copy(data.publicKey, 'key')}>
					<FontAwesomeIcon icon={faCopy} class="mr-1.5 h-3.5 w-3.5" />{copiedKey ? 'Copied!' : 'Copy'}
				</Button>
			{/if}
		</div>
		<div class="flex items-start gap-3">
			<div class="flex min-w-0 flex-1 flex-col gap-0.5">
				<span class="text-[0.7rem] font-semibold uppercase tracking-wide text-muted-foreground">Peer address</span>
				<code class="break-all font-mono text-[0.8125rem]">{data.grpcAddress || '—'}</code>
			</div>
			{#if data.grpcAddress}
				<Button variant="outline" size="sm" class="mt-4 shrink-0" onclick={() => copy(data.grpcAddress, 'addr')}>
					<FontAwesomeIcon icon={faCopy} class="mr-1.5 h-3.5 w-3.5" />{copiedAddr ? 'Copied!' : 'Copy'}
				</Button>
			{/if}
		</div>
		{#if data.capabilities.length > 0}
			<div class="flex flex-col gap-0.5">
				<span class="text-[0.7rem] font-semibold uppercase tracking-wide text-muted-foreground">Capabilities</span>
				<div class="flex flex-wrap gap-1.5">
					{#each data.capabilities as cap}
						<span class="rounded-full px-2 py-0.5 text-[0.68rem] font-semibold {capClass(cap)}">{capabilityLabel(cap)}</span>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</section>

<section class="mb-8">
	<h2 class="mb-3 text-base font-semibold">Partner libraries ({data.peers.length})</h2>
	{#if data.peers.length === 0}
		<p class="text-muted-foreground text-sm">No partner libraries connected yet.</p>
	{:else}
		<div class="overflow-x-auto">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Name</Table.Head>
						<Table.Head>Address</Table.Head>
						<Table.Head>Library ID</Table.Head>
						<Table.Head>Capabilities</Table.Head>
						<Table.Head>Status</Table.Head>
						<Table.Head></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.peers as peer}
						<Table.Row>
							<Table.Cell>{peer.displayName || '—'}</Table.Cell>
							<Table.Cell class="font-mono text-xs break-all">{peer.address}</Table.Cell>
							<Table.Cell class="font-mono text-xs break-all">{peer.nodeId}</Table.Cell>
							<Table.Cell>
								<div class="flex flex-wrap gap-1">
									{#each (peer.capabilities ?? []) as cap}
										<span class="rounded-full px-2 py-0.5 text-[0.68rem] font-semibold {capClass(cap)}">{capabilityLabel(cap)}</span>
									{:else}
										<span class="text-xs text-muted-foreground">—</span>
									{/each}
								</div>
							</Table.Cell>
							<Table.Cell>
								<span class="rounded-full px-2 py-0.5 text-xs font-semibold {(peer.status ?? 'PENDING') === 'CONNECTED' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}">
									{peer.status ?? 'PENDING'}
								</span>
							</Table.Cell>
							<Table.Cell>
								{#if peer.status === 'PENDING'}
									<form method="POST" action="?/approve" style="display:inline">
										<input type="hidden" name="nodeId" value={peer.nodeId} />
										<button type="submit" class="rounded border border-border px-2 py-0.5 text-xs text-foreground hover:bg-muted">Approve</button>
									</form>
								{/if}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	{/if}
</section>

<Separator class="mb-6" />

<section class="mb-8">
	<h2 class="mb-2 text-base font-semibold">Connect a partner library</h2>
	<p class="mb-4 text-xs text-muted-foreground">
		Share your Library ID and public key with the other library's administrator, and ask them to do
		the same. Enter their details below to establish the connection.
	</p>
	<form method="POST" action="?/connect" class="flex max-w-[480px] flex-col gap-3">
		<div class="flex flex-col gap-1.5">
			<Label for="displayName">Display name</Label>
			<Input id="displayName" type="text" name="displayName" value={form?.values?.displayName ?? ''} />
		</div>
		<div class="flex flex-col gap-1.5">
			<Label for="address">Address (host:port) <span class="text-destructive font-normal">*</span></Label>
			<Input
				id="address"
				type="text"
				name="address"
				value={form?.values?.address ?? ''}
				placeholder="192.168.1.10:50153"
				required
			/>
		</div>
		<div class="flex flex-col gap-1.5">
			<Label for="nodeId">Library ID <span class="text-destructive font-normal">*</span></Label>
			<Input id="nodeId" type="text" name="nodeId" value={form?.values?.nodeId ?? ''} required />
		</div>
		<div class="flex flex-col gap-1.5">
			<Label for="publicKey">Public key (base64) <span class="text-destructive font-normal">*</span></Label>
			<textarea
				id="publicKey"
				name="publicKey"
				rows="3"
				required
				class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
			>{form?.values?.publicKey ?? ''}</textarea>
		</div>
		<div>
			<Button type="submit"><FontAwesomeIcon icon={faLink} class="mr-1.5 h-3.5 w-3.5" />Connect library</Button>
		</div>
	</form>
</section>
