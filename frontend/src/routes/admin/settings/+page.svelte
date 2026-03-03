<script lang="ts">
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let allowLocalhost = $state(form?.allowLocalhost ?? data.allowLocalhost);
</script>

<svelte:head>
	<title>Settings — Admin — tiny-ils</title>
</svelte:head>

<h1>Settings</h1>

{#if form?.error}
	<p class="error">{form.error}</p>
{/if}
{#if form?.success}
	<p class="success">Settings saved.</p>
{/if}

<form method="POST">
	<section>
		<h2>Network</h2>

		<div class="field">
			<label for="publicUrl">Public URL</label>
			<p class="desc">The URL users access this server from. Used for CSRF protection and link generation.</p>
			<input
				id="publicUrl"
				type="url"
				name="publicUrl"
				value={form?.publicUrl ?? data.publicUrl}
				required
				placeholder="https://ils.example.com"
			/>
		</div>

		<div class="field toggle-field">
			<div class="toggle-text">
				<span class="toggle-label">Allow localhost access</span>
				<p class="desc">
					When enabled, requests from any loopback address (<code>localhost</code>,
					<code>127.x.x.x</code>, <code>::1</code>, <code>0.0.0.0</code>) are allowed alongside
					the public URL. Disable in production environments that should only be accessible via the
					configured public URL.
				</p>
			</div>
			<button
				type="button"
				class="toggle"
				class:on={allowLocalhost}
				aria-pressed={allowLocalhost}
				onclick={() => (allowLocalhost = !allowLocalhost)}
			>
				<span class="toggle-knob"></span>
			</button>
			<input type="hidden" name="allowLocalhost" value={allowLocalhost ? 'true' : 'false'} />
		</div>
	</section>

	<button type="submit" class="btn-primary">Save settings</button>
</form>

<style>
	h1 { margin: 0 0 1.5rem; }
	h2 { font-size: 1rem; margin: 0 0 1rem; }
	section { margin-bottom: 2rem; }
	form { max-width: 520px; }
	.field { display: flex; flex-direction: column; gap: 0.35rem; margin-bottom: 1.25rem; }
	label { font-size: 0.875rem; font-weight: 600; }
	.desc { font-size: 0.8rem; color: #6b7280; margin: 0; }
	input[type='url'] {
		padding: 0.5rem 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 0.875rem;
		width: 100%;
		box-sizing: border-box;
	}
	/* Toggle row */
	.toggle-field { flex-direction: row; align-items: center; gap: 1rem; }
	.toggle-text { flex: 1; display: flex; flex-direction: column; gap: 0.25rem; }
	.toggle-label { font-size: 0.875rem; font-weight: 600; }
	.toggle {
		flex-shrink: 0;
		width: 44px;
		height: 24px;
		border-radius: 12px;
		background: #d1d5db;
		border: none;
		cursor: pointer;
		padding: 2px;
		transition: background 0.2s;
		position: relative;
	}
	.toggle.on { background: #111; }
	.toggle-knob {
		display: block;
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: #fff;
		transition: transform 0.2s;
	}
	.toggle.on .toggle-knob { transform: translateX(20px); }
	code { font-family: monospace; font-size: 0.85em; }
	.btn-primary {
		padding: 0.5rem 1.25rem;
		background: #111;
		color: #fff;
		border: none;
		border-radius: 4px;
		font-size: 0.875rem;
		cursor: pointer;
	}
	.error { color: #dc2626; font-size: 0.875rem; margin: 0 0 1rem; }
	.success { color: #16a34a; font-size: 0.875rem; margin: 0 0 1rem; }
</style>
