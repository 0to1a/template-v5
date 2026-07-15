<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { completeLogin, getPendingEmail } from '$lib/login';

	const email = getPendingEmail();

	// Arriving here without an email (e.g. a hard refresh) restarts the flow.
	$effect(() => {
		if (!email) void goto(resolve('/login'));
	});

	let code = $state('');
	let busy = $state(false);
	let error = $state('');

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		error = '';
		try {
			await completeLogin(email, code, (path) => goto(resolve(path as '/')));
		} catch {
			// Invalid email and invalid code are deliberately the same error.
			error = 'Invalid email or code.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>Enter your code · Template v5</title>
</svelte:head>

<main class="mx-auto flex min-h-screen max-w-md flex-col justify-center gap-4 p-8">
	<h1 class="text-2xl font-semibold">Enter your code</h1>
	<p class="text-neutral-600">We sent a one-time code to {email}. It is valid for 5 minutes.</p>

	<form class="flex flex-col gap-3" onsubmit={submit}>
		<label class="flex flex-col gap-1">
			<span class="text-sm">One-time code</span>
			<input
				class="rounded border border-neutral-300 px-3 py-2 tracking-widest"
				type="text"
				name="code"
				inputmode="numeric"
				autocomplete="one-time-code"
				required
				bind:value={code}
			/>
		</label>

		{#if error}
			<p class="text-sm text-red-600">{error}</p>
		{/if}

		<button
			class="rounded bg-neutral-900 px-3 py-2 text-white disabled:opacity-50"
			type="submit"
			disabled={busy}
		>
			{busy ? 'Signing in…' : 'Sign in'}
		</button>
	</form>
</main>
