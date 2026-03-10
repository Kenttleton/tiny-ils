<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import * as Alert from '$lib/components/ui/alert';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faRocket } from '@fortawesome/free-solid-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
	<title>Setup — tiny-ils</title>
</svelte:head>

<div class="mx-auto mt-16 w-full max-w-lg px-4">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-2xl">Welcome to tiny-ils</Card.Title>
			<Card.Description>Create your admin account to get started.</Card.Description>
		</Card.Header>
		<Card.Content>
			{#if form?.error}
				<Alert.Root variant="destructive" class="mb-4">
					<Alert.Description>{form.error}</Alert.Description>
				</Alert.Root>
			{/if}

			<form method="POST" class="flex flex-col gap-4">
				<div class="flex flex-col gap-1.5">
					<Label for="publicUrl">Public URL</Label>
					<Input
						id="publicUrl"
						type="url"
						name="publicUrl"
						value={form?.publicUrl ?? data.detectedPublicUrl}
						required
						placeholder="https://ils.example.com"
					/>
					<p class="text-xs text-muted-foreground">The URL users access this server from. Used for CORS and link generation.</p>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="grpcAddress">Peer address</Label>
					<Input
						id="grpcAddress"
						type="text"
						name="grpcAddress"
						value={form?.grpcAddress ?? data.detectedGrpcAddress}
						placeholder="192.168.1.10:50153"
						class="font-mono"
					/>
					<p class="text-xs text-muted-foreground">
						The <code class="font-mono">host:port</code> other nodes use to reach this server's federation port.
						Auto-detected — override if behind NAT, a proxy, or VPN.
					</p>
				</div>

				<Separator class="my-1" />

				<div class="flex flex-col gap-1.5">
					<Label for="displayName">Display name <span class="text-muted-foreground font-normal">(optional)</span></Label>
					<Input id="displayName" type="text" name="displayName" value={form?.displayName ?? ''} autocomplete="name" />
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="email">Email</Label>
					<Input id="email" type="email" name="email" value={form?.email ?? ''} required autocomplete="email" />
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="password">Password</Label>
					<Input id="password" type="password" name="password" required autocomplete="new-password" minlength={8} />
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="confirm">Confirm password</Label>
					<Input id="confirm" type="password" name="confirm" required autocomplete="new-password" minlength={8} />
				</div>

				<Button type="submit" class="w-full mt-2"><FontAwesomeIcon icon={faRocket} class="mr-1.5 h-3.5 w-3.5" />Complete setup</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
