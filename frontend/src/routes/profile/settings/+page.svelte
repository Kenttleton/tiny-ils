<script lang="ts">
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	const profile = $derived(form?.profile ?? data.profile);
	const isSSOOnly = $derived(!!profile.sso_provider && !profile.has_password);
</script>

<svelte:head>
	<title>Settings — Profile — tiny-ils</title>
</svelte:head>

<h1>Settings</h1>

{#if form?.error}
	<p class="msg error">{form.error}</p>
{/if}
{#if data.linkError}
	<p class="msg error">Could not link Google account — it may already be linked to another user.</p>
{/if}
{#if form?.success && !form?.error}
	<p class="msg success">{form.unlinked ? 'SSO account unlinked.' : 'Settings saved.'}</p>
{/if}
{#if data.linked}
	<p class="msg success">{data.linked} account linked successfully.</p>
{/if}

<form method="POST" action="?/update">
	<section>
		<h2>Profile</h2>

		<div class="field">
			<label for="displayName">Display name</label>
			<input
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
		<section>
			<h2>Email &amp; Password</h2>

			<div class="field">
				<label for="email">Email</label>
				<input
					id="email"
					name="email"
					type="email"
					value={profile.email}
					placeholder="you@example.com"
				/>
			</div>

			{#if profile.has_password}
				<div class="field">
					<label for="currentPassword">Current password</label>
					<p class="desc">Required to change your password.</p>
					<input id="currentPassword" name="currentPassword" type="password" autocomplete="current-password" />
				</div>
			{/if}

			<div class="field">
				<label for="newPassword">{profile.has_password ? 'New password' : 'Set a password'}</label>
				{#if !profile.has_password}
					<p class="desc">Adding a password lets you log in with email and unlocks SSO unlinking.</p>
				{/if}
				<input id="newPassword" name="newPassword" type="password" autocomplete="new-password" minlength="8" />
			</div>
		</section>
	{:else}
		<section>
			<h2>Email &amp; Password</h2>
			<p class="desc">
				Your account is currently sign-in only via <strong>{profile.sso_provider}</strong>.
				Set a password below to enable email/password login and unlock the ability to unlink your
				{profile.sso_provider} account.
			</p>

			<div class="field">
				<label for="newPassword">Set a password</label>
				<input id="newPassword" name="newPassword" type="password" autocomplete="new-password" minlength="8" />
			</div>
		</section>
	{/if}

	<button type="submit" class="btn-primary">Save changes</button>
</form>

<section class="linked-section">
	<h2>Linked accounts</h2>
	{#if profile.sso_provider}
		<div class="linked-row">
			<span class="provider-badge">{profile.sso_provider}</span>
			<span class="linked-label">Linked</span>
			{#if profile.has_password}
				<form method="POST" action="?/unlinkSso" style="margin-left:auto">
					<button type="submit" class="btn-danger">Unlink</button>
				</form>
			{:else}
				<span class="unlink-hint">Set a password above to enable unlinking.</span>
			{/if}
		</div>
	{:else}
		<p class="desc">No SSO account linked.</p>
	{/if}

	{#if data.googleConfigured && !profile.sso_provider}
		<a href="/auth/google?link=true" class="btn-link-sso">Link Google account</a>
	{:else if data.googleConfigured && profile.sso_provider !== 'google'}
		<a href="/auth/google?link=true" class="btn-link-sso">Link a different Google account</a>
	{:else if !data.googleConfigured && data.isManager}
		<p class="desc admin-hint">Google SSO is not enabled on this node. Set <code>GOOGLE_CLIENT_ID</code> and <code>GOOGLE_CLIENT_SECRET</code> environment variables to enable it.</p>
	{/if}
</section>

<style>
	h1 { margin: 0 0 1.5rem; }
	h2 { font-size: 1rem; margin: 0 0 1rem; }
	section { margin-bottom: 2rem; max-width: 480px; }
	.field { display: flex; flex-direction: column; gap: 0.35rem; margin-bottom: 1.25rem; }
	label { font-size: 0.875rem; font-weight: 600; }
	.desc { font-size: 0.8rem; color: #6b7280; margin: 0; }
	input[type='text'],
	input[type='email'],
	input[type='password'] {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
		width: 100%;
		box-sizing: border-box;
	}
	.btn-primary {
		padding: 0.5rem 1.25rem;
		background: #111;
		color: #fff;
		border: none;
		border-radius: 4px;
		font-size: 0.875rem;
		cursor: pointer;
	}
	.btn-danger {
		padding: 0.4rem 1rem;
		background: #fff;
		color: #dc2626;
		border: 1px solid #fca5a5;
		border-radius: 4px;
		font-size: 0.875rem;
		cursor: pointer;
	}
	.btn-danger:hover { background: #fef2f2; }
	.linked-section { max-width: 480px; padding-top: 1.5rem; border-top: 1px solid #e5e7eb; margin-top: 0.5rem; }
	.linked-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.6rem 0.75rem;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		margin-bottom: 0.75rem;
		font-size: 0.875rem;
	}
	.provider-badge {
		text-transform: capitalize;
		font-weight: 600;
		background: #f3f4f6;
		padding: 0.15rem 0.5rem;
		border-radius: 4px;
		font-size: 0.8125rem;
	}
	.linked-label { color: #16a34a; font-size: 0.8125rem; }
	.unlink-hint { color: #9ca3af; font-size: 0.8rem; margin-left: auto; }
	.btn-link-sso {
		display: inline-block;
		padding: 0.4rem 1rem;
		background: #fff;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
		text-decoration: none;
	}
	.btn-link-sso:hover { background: #f9fafb; }
	.admin-hint { margin-top: 0.5rem; }
	.admin-hint code { font-size: 0.8rem; background: #f3f4f6; padding: 0.1rem 0.3rem; border-radius: 3px; }
	.msg { font-size: 0.875rem; margin: 0 0 1.25rem; }
	.error { color: #dc2626; }
	.success { color: #16a34a; }
</style>
