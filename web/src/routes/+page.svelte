<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { isAuthenticated } from '$lib/auth';
	import { performLogout } from '$lib/logout';

	// Public home page. It only reflects auth state; a protected application
	// shell is a future PRD, not part of the initial version.
	const authed = isAuthenticated();
	let busy = $state(false);

	async function signOut() {
		busy = true;
		try {
			await performLogout((path) => goto(resolve(path as '/')));
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>Template v5</title>
</svelte:head>

<main class="mx-auto flex min-h-screen max-w-md flex-col items-center justify-center gap-4 p-8">
	<h1 class="text-2xl font-semibold">Template v5</h1>

	{#if authed}
		<p class="text-neutral-600">You are signed in.</p>
		<button
			class="rounded bg-neutral-900 px-3 py-2 text-white disabled:opacity-50"
			onclick={signOut}
			disabled={busy}
		>
			{busy ? 'Signing out…' : 'Sign out'}
		</button>
	{:else}
		<p class="text-neutral-600">You are not signed in.</p>
		<a class="underline" href={resolve('/login')}>Sign in</a>
	{/if}
</main>
