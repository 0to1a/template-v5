<script lang="ts">
	import { resolve } from '$app/paths';
	import { isAuthenticated } from '$lib/auth';

	// Public home page. It only reflects auth state; a protected application
	// shell is a future PRD, not part of the initial version.
	const authed = isAuthenticated();
</script>

<main class="mx-auto flex min-h-screen max-w-md flex-col items-center justify-center gap-4 p-8">
	<h1 class="text-2xl font-semibold">Template v5</h1>

	{#if authed}
		<p class="text-neutral-600">You are signed in.</p>
	{:else}
		<p class="text-neutral-600">You are not signed in.</p>
		<a class="underline" href={resolve('/login')}>Sign in</a>
	{/if}
</main>
