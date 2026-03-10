<script lang="ts">
	import type { ActionData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import * as Alert from '$lib/components/ui/alert';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faUserPlus } from '@fortawesome/free-solid-svg-icons';

	let { form }: { form: ActionData } = $props();
</script>

<svelte:head>
	<title>Register — tiny-ils</title>
</svelte:head>

<div class="mx-auto mt-16 w-full max-w-sm px-4">
	<Card.Root>
		<Card.Header>
			<Card.Title class="text-2xl">Create account</Card.Title>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			{#if form?.error}
				<Alert.Root variant="destructive">
					<Alert.Description>{form.error}</Alert.Description>
				</Alert.Root>
			{/if}

			<form method="POST" class="flex flex-col gap-4">
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

				<Button type="submit" class="w-full"><FontAwesomeIcon icon={faUserPlus} class="mr-1.5 h-3.5 w-3.5" />Create account</Button>
			</form>

			<p class="text-center text-sm text-muted-foreground">
				Already have an account? <a href="/auth/login" class="text-foreground underline-offset-4 hover:underline">Sign in</a>
			</p>
		</Card.Content>
	</Card.Root>
</div>
