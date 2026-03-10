<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';
	import * as Table from '$lib/components/ui/table';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faArrowUp, faArrowDown, faTrash, faUserPlus } from '@fortawesome/free-solid-svg-icons';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let showCreate = $state(false);
</script>

<svelte:head>
	<title>Users — Admin — tiny-ils</title>
</svelte:head>

<div class="mb-6 flex items-center justify-between">
	<h1 class="text-2xl font-bold">Users</h1>
	<Button onclick={() => (showCreate = !showCreate)} variant={showCreate ? 'outline' : 'default'}>
		{#if !showCreate}<FontAwesomeIcon icon={faUserPlus} class="mr-1.5 h-3.5 w-3.5" />{/if}{showCreate ? 'Cancel' : 'Create user'}
	</Button>
</div>

{#if form?.error}
	<Alert.Root variant="destructive" class="mb-4">
		<Alert.Description>{form.error}</Alert.Description>
	</Alert.Root>
{/if}
{#if form?.success}
	<p class="mb-5 text-sm text-green-600">
		{#if form.action === 'created'}User created successfully.{/if}
		{#if form.action === 'promoted'}User promoted to Manager.{/if}
		{#if form.action === 'demoted'}User demoted to User.{/if}
		{#if form.action === 'deleted'}User deleted.{/if}
	</p>
{/if}

{#if showCreate}
	<section class="mb-6 max-w-[480px] rounded-md border border-border bg-muted/40 p-5">
		<h2 class="mb-4 text-base font-semibold">Create user account</h2>
		<form method="POST" action="?/create" class="flex flex-col gap-3">
			<div class="flex flex-col gap-1.5">
				<Label for="email">Email <span class="text-destructive">*</span></Label>
				<Input id="email" name="email" type="email" placeholder="user@example.com" required />
			</div>
			<div class="flex flex-col gap-1.5">
				<Label for="displayName">Display name</Label>
				<Input id="displayName" name="displayName" type="text" placeholder="Optional" />
			</div>
			<div class="flex flex-col gap-1.5">
				<Label for="password">Password <span class="text-destructive">*</span></Label>
				<Input id="password" name="password" type="password" minlength={8} required autocomplete="new-password" />
			</div>
			<div>
				<Button type="submit">Create account</Button>
			</div>
		</form>
	</section>
{/if}

<section class="mb-8">
	{#if data.users.length === 0}
		<p class="text-muted-foreground text-sm">No users found.</p>
	{:else}
		<div class="overflow-x-auto">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Name</Table.Head>
						<Table.Head>Email</Table.Head>
						<Table.Head>Auth</Table.Head>
						<Table.Head>Role</Table.Head>
						<Table.Head>Created</Table.Head>
						<Table.Head></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.users as user}
						<Table.Row>
							<Table.Cell>{user.display_name || '—'}</Table.Cell>
							<Table.Cell>{user.email || '—'}</Table.Cell>
							<Table.Cell>
								<div class="flex flex-wrap gap-1">
									{#if user.sso_provider}
										<span class="rounded bg-blue-100 px-1.5 py-0.5 text-xs font-medium capitalize text-blue-800">{user.sso_provider}</span>
									{/if}
									{#if user.has_password}
										<span class="rounded bg-zinc-100 px-1.5 py-0.5 text-xs font-medium text-zinc-700">password</span>
									{/if}
								</div>
							</Table.Cell>
							<Table.Cell>
								<span class="rounded px-2 py-0.5 text-xs font-semibold {user.role === 'MANAGER' ? 'bg-amber-100 text-amber-800' : 'bg-zinc-100 text-zinc-600'}">
									{user.role || 'USER'}
								</span>
							</Table.Cell>
							<Table.Cell class="whitespace-nowrap text-xs text-muted-foreground">
								{new Date(user.created_at * 1000).toLocaleDateString()}
							</Table.Cell>
							<Table.Cell class="whitespace-nowrap">
								{#if user.role === 'MANAGER'}
									<form method="POST" action="?/demote" class="inline">
										<input type="hidden" name="userId" value={user.id} />
										<button type="submit" class="mr-1 rounded border border-border px-2 py-0.5 text-xs text-foreground hover:bg-muted"><FontAwesomeIcon icon={faArrowDown} class="mr-1.5 h-3.5 w-3.5" />Demote</button>
									</form>
								{:else}
									<form method="POST" action="?/promote" class="inline">
										<input type="hidden" name="userId" value={user.id} />
										<button type="submit" class="mr-1 rounded border border-border px-2 py-0.5 text-xs text-foreground hover:bg-muted"><FontAwesomeIcon icon={faArrowUp} class="mr-1.5 h-3.5 w-3.5" />Promote</button>
									</form>
								{/if}
								<form method="POST" action="?/delete" class="inline">
									<input type="hidden" name="userId" value={user.id} />
									<button
										type="submit"
										class="rounded border border-red-300 px-2 py-0.5 text-xs text-red-600 hover:bg-red-50"
										onclick={(e) => {
											if (!confirm(`Delete ${user.email || user.display_name}? This cannot be undone.`)) {
												e.preventDefault();
											}
										}}
									>
										<FontAwesomeIcon icon={faTrash} class="mr-1.5 h-3.5 w-3.5" />Delete
									</button>
								</form>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		{#if data.total > data.limit}
			<div class="mt-4 flex items-center gap-4 text-sm">
				{#if data.page > 1}
					<a href="?page={data.page - 1}" class="text-foreground hover:underline">← Previous</a>
				{/if}
				<span class="text-muted-foreground">Page {data.page} of {Math.ceil(data.total / data.limit)}</span>
				{#if data.page * data.limit < data.total}
					<a href="?page={data.page + 1}" class="text-foreground hover:underline">Next →</a>
				{/if}
			</div>
		{/if}
	{/if}
</section>
