<script lang="ts">
	import type { PageData, ActionData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let showCreate = $state(false);
</script>

<svelte:head>
	<title>Users — Admin — tiny-ils</title>
</svelte:head>

<div class="header-row">
	<h1>Users</h1>
	<button class="btn-primary" onclick={() => (showCreate = !showCreate)}>
		{showCreate ? 'Cancel' : '+ Create user'}
	</button>
</div>

{#if form?.error}
	<p class="msg error">{form.error}</p>
{/if}
{#if form?.success}
	<p class="msg success">
		{#if form.action === 'created'}User created successfully.{/if}
		{#if form.action === 'promoted'}User promoted to Manager.{/if}
		{#if form.action === 'demoted'}User demoted to User.{/if}
		{#if form.action === 'deleted'}User deleted.{/if}
	</p>
{/if}

{#if showCreate}
	<section class="create-section">
		<h2>Create user account</h2>
		<form method="POST" action="?/create" class="create-form">
			<div class="field">
				<label for="email">Email <span class="required">*</span></label>
				<input id="email" name="email" type="email" placeholder="user@example.com" required />
			</div>
			<div class="field">
				<label for="displayName">Display name</label>
				<input id="displayName" name="displayName" type="text" placeholder="Optional" />
			</div>
			<div class="field">
				<label for="password">Password <span class="required">*</span></label>
				<input id="password" name="password" type="password" minlength="8" required autocomplete="new-password" />
			</div>
			<button type="submit" class="btn-primary">Create account</button>
		</form>
	</section>
{/if}

<section>
	{#if data.users.length === 0}
		<p class="empty">No users found.</p>
	{:else}
		<table>
			<thead>
				<tr>
					<th>Name</th>
					<th>Email</th>
					<th>Auth</th>
					<th>Role</th>
					<th>Created</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				{#each data.users as user}
					<tr>
						<td>{user.display_name || '—'}</td>
						<td>{user.email || '—'}</td>
						<td class="auth-col">
							{#if user.sso_provider}
								<span class="badge sso">{user.sso_provider}</span>
							{/if}
							{#if user.has_password}
								<span class="badge pw">password</span>
							{/if}
						</td>
						<td>
							<span class="role-badge {user.role === 'MANAGER' ? 'manager' : 'user'}">
								{user.role || 'USER'}
							</span>
						</td>
						<td class="date-col">
							{new Date(user.created_at * 1000).toLocaleDateString()}
						</td>
						<td class="actions-col">
							{#if user.role === 'MANAGER'}
								<form method="POST" action="?/demote" class="inline">
									<input type="hidden" name="userId" value={user.id} />
									<button type="submit" class="btn-action">Demote</button>
								</form>
							{:else}
								<form method="POST" action="?/promote" class="inline">
									<input type="hidden" name="userId" value={user.id} />
									<button type="submit" class="btn-action">Promote</button>
								</form>
							{/if}
							<form method="POST" action="?/delete" class="inline">
								<input type="hidden" name="userId" value={user.id} />
								<button
									type="submit"
									class="btn-danger"
									onclick={(e) => {
										if (!confirm(`Delete ${user.email || user.display_name}? This cannot be undone.`)) {
											e.preventDefault();
										}
									}}
								>
									Delete
								</button>
							</form>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>

		{#if data.total > data.limit}
			<div class="pagination">
				{#if data.page > 1}
					<a href="?page={data.page - 1}" class="page-link">← Previous</a>
				{/if}
				<span class="page-info"
					>Page {data.page} of {Math.ceil(data.total / data.limit)}</span
				>
				{#if data.page * data.limit < data.total}
					<a href="?page={data.page + 1}" class="page-link">Next →</a>
				{/if}
			</div>
		{/if}
	{/if}
</section>

<style>
	.header-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1.5rem;
	}
	h1 { margin: 0; }
	h2 { font-size: 1rem; margin: 0 0 1rem; }
	section { margin-bottom: 2rem; }
	.create-section {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		padding: 1.25rem;
		margin-bottom: 1.5rem;
		max-width: 480px;
	}
	.create-form { display: flex; flex-direction: column; gap: 0.75rem; }
	.field { display: flex; flex-direction: column; gap: 0.3rem; }
	label { font-size: 0.875rem; font-weight: 500; }
	.required { color: #dc2626; }
	input[type='email'],
	input[type='text'],
	input[type='password'] {
		padding: 0.45rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
	}
	table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
	th {
		text-align: left;
		padding: 0.5rem 0.75rem;
		border-bottom: 2px solid #e5e7eb;
		color: #6b7280;
		font-weight: 500;
		white-space: nowrap;
	}
	td { padding: 0.5rem 0.75rem; border-bottom: 1px solid #f3f4f6; vertical-align: middle; }
	.auth-col { display: flex; gap: 0.35rem; flex-wrap: wrap; }
	.date-col { color: #9ca3af; font-size: 0.8rem; white-space: nowrap; }
	.actions-col { white-space: nowrap; }
	.badge {
		display: inline-block;
		padding: 0.1rem 0.45rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: capitalize;
	}
	.badge.sso { background: #dbeafe; color: #1e40af; }
	.badge.pw { background: #f3f4f6; color: #374151; }
	.role-badge {
		display: inline-block;
		padding: 0.15rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
	}
	.role-badge.manager { background: #fef3c7; color: #92400e; }
	.role-badge.user { background: #f3f4f6; color: #6b7280; }
	.inline { display: inline; }
	.btn-primary {
		padding: 0.45rem 1rem;
		background: #111;
		color: #fff;
		border: none;
		border-radius: 4px;
		font-size: 0.875rem;
		cursor: pointer;
		align-self: flex-start;
	}
	.btn-action {
		padding: 0.2rem 0.6rem;
		background: #fff;
		color: #374151;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.75rem;
		cursor: pointer;
		margin-right: 0.25rem;
	}
	.btn-action:hover { background: #f9fafb; }
	.btn-danger {
		padding: 0.2rem 0.6rem;
		background: #fff;
		color: #dc2626;
		border: 1px solid #fca5a5;
		border-radius: 4px;
		font-size: 0.75rem;
		cursor: pointer;
	}
	.btn-danger:hover { background: #fef2f2; }
	.msg { font-size: 0.875rem; margin: 0 0 1.25rem; }
	.error { color: #dc2626; }
	.success { color: #16a34a; }
	.empty { color: #6b7280; font-size: 0.875rem; }
	.pagination {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding-top: 1rem;
		font-size: 0.875rem;
	}
	.page-link { color: #111; text-decoration: none; }
	.page-link:hover { text-decoration: underline; }
	.page-info { color: #6b7280; }
</style>
