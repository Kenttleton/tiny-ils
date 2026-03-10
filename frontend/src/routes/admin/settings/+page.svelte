<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let allowLocalhost = $state(form?.allowLocalhost ?? data.allowLocalhost);
</script>

<svelte:head>
	<title>Settings — Admin — tiny-ils</title>
</svelte:head>

<h1 class="mb-6 text-2xl font-bold">Settings</h1>

{#if form?.error}
	<Alert.Root variant="destructive" class="mb-4">
		<Alert.Description>{form.error}</Alert.Description>
	</Alert.Root>
{/if}
{#if form?.success}
	<p class="mb-4 text-sm text-green-600">Settings saved.</p>
{/if}

<form method="POST" class="max-w-[520px]">
	<section class="mb-8">
		<h2 class="mb-4 text-base font-semibold">Network</h2>

		<div class="mb-5 flex flex-col gap-1.5">
			<Label for="publicUrl">Public URL</Label>
			<p class="m-0 text-xs text-muted-foreground">
				The URL users access this server from. Used for CSRF protection and link generation.
			</p>
			<Input
				id="publicUrl"
				type="url"
				name="publicUrl"
				value={form?.publicUrl ?? data.publicUrl}
				required
				placeholder="https://ils.example.com"
			/>
		</div>

		<div class="mb-5 flex flex-col gap-1.5">
			<Label for="grpcAddress">Peer address</Label>
			<p class="m-0 text-xs text-muted-foreground">
				The <code class="font-mono text-[0.85em]">host:port</code> other nodes use to reach this server's federation
				port (default 50153). Override if behind NAT, a proxy, or VPN.
			</p>
			<Input
				id="grpcAddress"
				type="text"
				name="grpcAddress"
				value={form?.grpcAddress ?? data.grpcAddress}
				placeholder="192.168.1.10:50153"
			/>
		</div>

		<div class="mb-5 flex flex-row items-center gap-4">
			<div class="flex flex-1 flex-col gap-1">
				<span class="text-sm font-semibold">Allow localhost access</span>
				<p class="m-0 text-xs text-muted-foreground">
					When enabled, requests from any loopback address (<code class="font-mono text-[0.85em]">localhost</code>,
					<code class="font-mono text-[0.85em]">127.x.x.x</code>, <code class="font-mono text-[0.85em]">::1</code>, <code class="font-mono text-[0.85em]">0.0.0.0</code>) are
					allowed alongside the public URL. Disable in production environments
					that should only be accessible via the configured public URL.
				</p>
			</div>
			<button
				type="button"
				aria-label="Allow localhost access"
				class="relative h-6 w-11 shrink-0 rounded-full border-none p-0.5 transition-colors {allowLocalhost ? 'bg-foreground' : 'bg-muted-foreground/30'} cursor-pointer"
				aria-pressed={allowLocalhost}
				onclick={() => (allowLocalhost = !allowLocalhost)}
			>
				<span
					class="block h-5 w-5 rounded-full bg-white transition-transform {allowLocalhost ? 'translate-x-5' : 'translate-x-0'}"
				></span>
			</button>
			<input type="hidden" name="allowLocalhost" value={allowLocalhost ? 'true' : 'false'} />
		</div>
	</section>

	<Button type="submit">Save settings</Button>
</form>
