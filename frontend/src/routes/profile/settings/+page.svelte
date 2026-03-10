<script lang="ts">
  import type { ActionData, PageData } from "./$types";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import * as Alert from "$lib/components/ui/alert";
  import { Separator } from "$lib/components/ui/separator";
  import { FontAwesomeIcon } from '@fortawesome/svelte-fontawesome';
  import { faFloppyDisk, faLinkSlash } from '@fortawesome/free-solid-svg-icons';
  import { faGoogle } from '@fortawesome/free-brands-svg-icons';

  let { data, form }: { data: PageData; form: ActionData } = $props();

  const profile = $derived(form?.profile ?? data.profile);
  const isSSOOnly = $derived(!!profile.sso_provider && !profile.has_password);
</script>

<svelte:head>
  <title>Settings — Profile — tiny-ils</title>
</svelte:head>

<h1 class="mb-6 text-2xl font-bold">Settings</h1>

{#if form?.error}
  <Alert.Root variant="destructive" class="mb-4">
    <Alert.Description>{form.error}</Alert.Description>
  </Alert.Root>
{/if}
{#if data.linkError}
  <Alert.Root variant="destructive" class="mb-4">
    <Alert.Description
      >Could not link Google account — it may already be linked to another user.</Alert.Description
    >
  </Alert.Root>
{/if}
{#if form?.success && !form?.error}
  <p class="mb-5 text-sm text-green-600">
    {form.unlinked ? "SSO account unlinked." : "Settings saved."}
  </p>
{/if}
{#if data.linked}
  <p class="mb-5 text-sm text-green-600">
    {data.linked} account linked successfully.
  </p>
{/if}

<form method="POST" action="?/update" class="max-w-[480px]">
  <section class="mb-8">
    <h2 class="mb-4 text-base font-semibold">Profile</h2>

    <div class="mb-5 flex flex-col gap-1.5">
      <Label for="displayName">Display name</Label>
      <Input
        id="displayName"
        name="displayName"
        type="text"
        value={profile.display_name}
        required
        placeholder="Your name"
      />
    </div>
  </section>

  {#if !isSSOOnly}
    <section class="mb-8">
      <h2 class="mb-4 text-base font-semibold">Email &amp; Password</h2>

      <div class="mb-5 flex flex-col gap-1.5">
        <Label for="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          value={profile.email}
          placeholder="you@example.com"
        />
      </div>

      {#if profile.has_password}
        <div class="mb-5 flex flex-col gap-1.5">
          <Label for="currentPassword">Current password</Label>
          <p class="m-0 text-xs text-muted-foreground">
            Required to change your password.
          </p>
          <Input
            id="currentPassword"
            name="currentPassword"
            type="password"
            autocomplete="current-password"
          />
        </div>
      {/if}

      <div class="mb-5 flex flex-col gap-1.5">
        <Label for="newPassword"
          >{profile.has_password ? "New password" : "Set a password"}</Label
        >
        {#if !profile.has_password}
          <p class="m-0 text-xs text-muted-foreground">
            Adding a password lets you log in with email and unlocks SSO
            unlinking.
          </p>
        {/if}
        <Input
          id="newPassword"
          name="newPassword"
          type="password"
          autocomplete="new-password"
          minlength={8}
        />
      </div>
    </section>
  {:else}
    <section class="mb-8">
      <h2 class="mb-4 text-base font-semibold">Email &amp; Password</h2>
      <p class="mb-4 text-xs text-muted-foreground">
        Your account is currently sign-in only via <strong
          >{profile.sso_provider}</strong
        >. Set a password below to enable email/password login and unlock the
        ability to unlink your
        {profile.sso_provider} account.
      </p>

      <div class="mb-5 flex flex-col gap-1.5">
        <Label for="newPassword">Set a password</Label>
        <Input
          id="newPassword"
          name="newPassword"
          type="password"
          autocomplete="new-password"
          minlength={8}
        />
      </div>
    </section>
  {/if}

  <Button type="submit"><FontAwesomeIcon icon={faFloppyDisk} class="mr-1.5 h-3.5 w-3.5" />Save changes</Button>
</form>

<Separator class="my-6 max-w-[480px]" />

<section class="max-w-[480px]">
  <h2 class="mb-3 text-base font-semibold">Linked accounts</h2>
  {#if profile.sso_provider}
    <div
      class="mb-3 flex items-center gap-3 rounded-md border border-border px-3 py-2.5 text-sm"
    >
      <span
        class="rounded bg-zinc-100 px-2 py-0.5 text-[0.8125rem] font-semibold capitalize"
        >{profile.sso_provider}</span
      >
      <span class="text-[0.8125rem] text-green-600">Linked</span>
      {#if profile.has_password}
        <form method="POST" action="?/unlinkSso" class="ml-auto">
          <Button type="submit" variant="destructive" size="sm"><FontAwesomeIcon icon={faLinkSlash} class="mr-1.5 h-3.5 w-3.5" />Unlink</Button>
        </form>
      {:else}
        <span class="ml-auto text-xs text-muted-foreground"
          >Set a password above to enable unlinking.</span
        >
      {/if}
    </div>
  {:else}
    <p class="text-muted-foreground text-sm">No SSO account linked.</p>
  {/if}

  {#if data.googleConfigured && !profile.sso_provider}
    <a
      href="/auth/google?link=true"
      class="mt-2 inline-flex items-center rounded-md border border-border bg-background px-4 py-2 text-sm text-foreground no-underline hover:bg-muted"
      ><FontAwesomeIcon icon={faGoogle} class="mr-1.5 h-3.5 w-3.5" />Link Google account</a
    >
  {:else if data.googleConfigured && profile.sso_provider !== "google"}
    <a
      href="/auth/google?link=true"
      class="mt-2 inline-flex items-center rounded-md border border-border bg-background px-4 py-2 text-sm text-foreground no-underline hover:bg-muted"
      ><FontAwesomeIcon icon={faGoogle} class="mr-1.5 h-3.5 w-3.5" />Link a different Google account</a
    >
  {:else if !data.googleConfigured && data.isManager}
    <p class="mt-2 text-xs text-muted-foreground">
      Google SSO is not enabled on this node. Set <code
        class="rounded bg-muted px-1 py-0.5 font-mono text-[0.8rem]"
        >GOOGLE_CLIENT_ID</code
      >
      and
      <code class="rounded bg-muted px-1 py-0.5 font-mono text-[0.8rem]"
        >GOOGLE_CLIENT_SECRET</code
      > environment variables to enable it.
    </p>
  {/if}
</section>
