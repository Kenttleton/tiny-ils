<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import * as Sheet from '$lib/components/ui/sheet';
	import { Separator } from '$lib/components/ui/separator';
	import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
	import {
		faBars,
		faGauge,
		faBook,
		faListCheck,
		faArrowRightArrowLeft,
		faUsers,
		faNetworkWired,
		faGear
	} from '@fortawesome/free-solid-svg-icons';

	let { children } = $props();

	const navLinks = [
		{ href: '/admin', label: 'Dashboard', icon: faGauge },
		{ href: '/admin/curios', label: 'Curios', icon: faBook },
		{ href: '/admin/loans', label: 'Loans', icon: faListCheck },
		{ href: '/admin/transfers', label: 'Transfers', icon: faArrowRightArrowLeft },
		{ href: '/admin/users', label: 'Users', icon: faUsers },
		{ href: '/admin/peers', label: 'Network', icon: faNetworkWired },
		{ href: '/admin/settings', label: 'Settings', icon: faGear }
	];

	let mobileOpen = $state(false);
</script>

<div class="flex gap-6">
	<!-- Desktop sidebar -->
	<aside class="hidden w-44 shrink-0 md:block">
		<nav class="sticky top-20 flex flex-col gap-0.5">
			<p class="mb-2 px-2 text-[0.65rem] font-semibold uppercase tracking-widest text-muted-foreground">Admin</p>
			{#each navLinks as link}
				<a
					href={link.href}
					class="flex items-center rounded-md px-3 py-1.5 text-sm transition-colors hover:bg-accent hover:text-accent-foreground
						{page.url.pathname === link.href ? 'bg-accent text-accent-foreground font-medium' : 'text-muted-foreground'}"
				>
					<FontAwesomeIcon icon={link.icon} class="mr-2 h-3.5 w-3.5 shrink-0" />
					{link.label}
				</a>
			{/each}
		</nav>
	</aside>

	<!-- Mobile nav bar -->
	<div class="mb-4 flex items-center gap-2 md:hidden">
		<Sheet.Root bind:open={mobileOpen}>
			<Sheet.Trigger>
				<Button variant="outline" size="sm" aria-label="Open admin menu">
					<FontAwesomeIcon icon={faBars} class="mr-1.5 h-3.5 w-3.5" />
					Admin
				</Button>
			</Sheet.Trigger>
			<Sheet.Content side="left" class="w-56">
				<Sheet.Header>
					<Sheet.Title>Admin</Sheet.Title>
				</Sheet.Header>
				<nav class="mt-4 flex flex-col gap-0.5">
					{#each navLinks as link}
						<a
							href={link.href}
							onclick={() => (mobileOpen = false)}
							class="flex items-center rounded-md px-3 py-2 text-sm transition-colors hover:bg-accent hover:text-accent-foreground
								{page.url.pathname === link.href ? 'bg-accent text-accent-foreground font-medium' : 'text-muted-foreground'}"
						>
							<FontAwesomeIcon icon={link.icon} class="mr-2 h-3.5 w-3.5 shrink-0" />
							{link.label}
						</a>
					{/each}
				</nav>
			</Sheet.Content>
		</Sheet.Root>
	</div>

	<!-- Page content -->
	<div class="min-w-0 flex-1">
		{@render children()}
	</div>
</div>
