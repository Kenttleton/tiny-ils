<script lang="ts">
	import { page } from '$app/state';
	import { isManager } from '$lib/auth';
	import { Button } from '$lib/components/ui/button';
	import * as Sheet from '$lib/components/ui/sheet';
	import { Separator } from '$lib/components/ui/separator';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import { faBars } from '@fortawesome/free-solid-svg-icons';
	import '../app.css';

	let { data, children } = $props();

	const user = $derived(data.user);
	const nodeId = $derived(data.nodeId);
	const manager = $derived(user ? isManager(user.claims, nodeId) : false);
	const isAuthPage = $derived(page.url.pathname.startsWith('/auth/'));

	let mobileMenuOpen = $state(false);
</script>

<svelte:head>
	<title>tiny-ils</title>
</svelte:head>

{#if !isAuthPage}
	<nav class="sticky top-0 z-50 border-b border-border bg-background">
		<div class="mx-auto flex max-w-screen-xl items-center gap-4 px-4 py-3">
			<!-- Brand -->
			<a href="/" class="text-base font-bold text-foreground no-underline">tiny-ils</a>

			<!-- Desktop links -->
			<div class="hidden flex-1 items-center gap-4 md:flex">
				<a href="/browse" class="text-sm text-muted-foreground hover:text-foreground transition-colors">Browse</a>
				{#if user}
					<a href="/profile/loans" class="text-sm text-muted-foreground hover:text-foreground transition-colors">Profile</a>
					{#if manager}
						<a href="/admin" class="text-sm text-muted-foreground hover:text-foreground transition-colors">Admin</a>
					{/if}
				{/if}
			</div>

			<!-- Desktop auth -->
			<div class="ml-auto hidden items-center md:flex">
				{#if user}
					<form method="POST" action="/auth/logout">
						<Button variant="outline" size="sm" type="submit">Sign out</Button>
					</form>
				{:else}
					<Button variant="outline" size="sm" href="/auth/login">Sign in</Button>
				{/if}
			</div>

			<!-- Mobile hamburger -->
			<div class="ml-auto md:hidden">
				<Sheet.Root bind:open={mobileMenuOpen}>
					<Sheet.Trigger>
						<Button variant="ghost" size="icon" aria-label="Open menu">
							<FontAwesomeIcon icon={faBars} class="h-4 w-4" />
						</Button>
					</Sheet.Trigger>
					<Sheet.Content side="right" class="w-64">
						<Sheet.Header>
							<Sheet.Title>tiny-ils</Sheet.Title>
						</Sheet.Header>
						<div class="mt-4 flex flex-col gap-1">
							<a
								href="/browse"
								onclick={() => (mobileMenuOpen = false)}
								class="rounded-md px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
							>Browse</a>
							{#if user}
								<a
									href="/profile/loans"
									onclick={() => (mobileMenuOpen = false)}
									class="rounded-md px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
								>Profile</a>
								{#if manager}
									<a
										href="/admin"
										onclick={() => (mobileMenuOpen = false)}
										class="rounded-md px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground transition-colors"
									>Admin</a>
								{/if}
							{/if}
							<Separator class="my-2" />
							{#if user}
								<form method="POST" action="/auth/logout">
									<Button variant="outline" class="w-full" type="submit" onclick={() => (mobileMenuOpen = false)}>
										Sign out
									</Button>
								</form>
							{:else}
								<Button variant="outline" class="w-full" href="/auth/login" onclick={() => (mobileMenuOpen = false)}>
									Sign in
								</Button>
							{/if}
						</div>
					</Sheet.Content>
				</Sheet.Root>
			</div>
		</div>
	</nav>
{/if}

<main class="mx-auto max-w-screen-xl px-4 py-8">
	{@render children()}
</main>
