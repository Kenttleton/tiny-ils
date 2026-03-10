<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import * as Alert from '$lib/components/ui/alert';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faArrowRightToBracket } from '@fortawesome/free-solid-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
	<title>Partner library sign-in — tiny-ils</title>
</svelte:head>

<div class="mx-auto mt-16 w-full max-w-md px-4">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-2xl">Sign in from another library</Card.Title>
			<Card.Description>
				If your account is registered at a different partner library, enter that library's connection
				details and your user ID to sign in here.
			</Card.Description>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			{#if form?.error}
				<Alert.Root variant="destructive">
					<Alert.Description>{form.error}</Alert.Description>
				</Alert.Root>
			{/if}

			<form method="POST" class="flex flex-col gap-5">
				<input type="hidden" name="next" value={data.next} />

				<div class="flex flex-col gap-1.5">
					<Label for="home_node_address">Home library address</Label>
					<Input
						id="home_node_address"
						type="text"
						name="home_node_address"
						value={form?.homeNodeAddress ?? ''}
						required
						placeholder="e.g. library.example.org:50153"
						autocomplete="off"
						class="font-mono"
					/>
					<p class="text-xs text-muted-foreground">The gRPC address of the library where your account lives.</p>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="home_node_id">Home library ID</Label>
					<Input
						id="home_node_id"
						type="text"
						name="home_node_id"
						value={form?.homeNodeId ?? ''}
						required
						placeholder="e.g. abc123..."
						autocomplete="off"
						class="font-mono"
					/>
					<p class="text-xs text-muted-foreground">Found under Network on your home library's admin page.</p>
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="user_id">Your user ID</Label>
					<Input
						id="user_id"
						type="text"
						name="user_id"
						value={form?.userId ?? ''}
						required
						placeholder="e.g. 550e8400-..."
						autocomplete="off"
						class="font-mono"
					/>
					<p class="text-xs text-muted-foreground">Found on your profile page at your home library.</p>
				</div>

				<Button type="submit" class="w-full"><FontAwesomeIcon icon={faArrowRightToBracket} class="mr-1.5 h-3.5 w-3.5" />Sign in</Button>
			</form>

			<p class="text-center text-sm text-muted-foreground">
				Have an account here? <a href="/auth/login" class="text-foreground underline-offset-4 hover:underline">Local sign in</a>
			</p>
		</Card.Content>
	</Card.Root>
</div>
