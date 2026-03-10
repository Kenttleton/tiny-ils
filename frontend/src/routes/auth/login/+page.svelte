<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import * as Alert from '$lib/components/ui/alert';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faArrowRightToBracket } from '@fortawesome/free-solid-svg-icons';
	import { faGoogle } from '@fortawesome/free-brands-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();
</script>

<svelte:head>
	<title>Sign in — tiny-ils</title>
</svelte:head>

<div class="mx-auto mt-16 w-full max-w-sm px-4">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-2xl">Sign in</Card.Title>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			{#if form?.error}
				<Alert.Root variant="destructive">
					<Alert.Description>{form.error}</Alert.Description>
				</Alert.Root>
			{/if}

			<form method="POST" class="flex flex-col gap-4">
				<input type="hidden" name="next" value={data.next} />

				<div class="flex flex-col gap-1.5">
					<Label for="email">Email</Label>
					<Input id="email" type="email" name="email" value={form?.email ?? ''} required autocomplete="email" />
				</div>

				<div class="flex flex-col gap-1.5">
					<Label for="password">Password</Label>
					<Input id="password" type="password" name="password" required autocomplete="current-password" />
				</div>

				<Button type="submit" class="w-full"><FontAwesomeIcon icon={faArrowRightToBracket} class="mr-1.5 h-3.5 w-3.5" />Sign in</Button>
			</form>

			<p class="text-center text-sm text-muted-foreground">
				Don't have an account? <a href="/auth/register" class="text-foreground underline-offset-4 hover:underline">Register</a>
			</p>

			<Separator />

			<a
				href="/auth/google"
				class="flex items-center justify-center rounded-md border border-border px-4 py-2 text-sm text-foreground transition-colors hover:bg-accent"
			><FontAwesomeIcon icon={faGoogle} class="mr-1.5 h-3.5 w-3.5" />Sign in with Google</a>

			<p class="text-center text-xs text-muted-foreground">
				Account at another library? <a href="/auth/cross-node" class="underline-offset-4 hover:underline">Partner library sign-in</a>
			</p>
		</Card.Content>
	</Card.Root>
</div>
