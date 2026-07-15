<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authClient } from '$lib/client';
	import { setPendingEmail } from '$lib/login';

	let email = $state('');
	let busy = $state(false);
	let error = $state('');

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		error = '';
		try {
			// The response is always generic: it never reveals whether the
			// email belongs to an account.
			await authClient.requestLogin({ email });
			setPendingEmail(email);
			await goto(resolve('/login/otp'));
		} catch {
			error = 'Something went wrong. Please try again.';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>Sign in · Template v5</title>
</svelte:head>

<main class="mx-auto flex min-h-screen max-w-md flex-col justify-center gap-4 p-8">
	<h1 class="text-2xl font-semibold">Sign in</h1>
	<p class="text-neutral-600">Enter your email and we will send you a one-time code.</p>

	<form class="flex flex-col gap-3" onsubmit={submit}>
		<label class="flex flex-col gap-1">
			<span class="text-sm">Email</span>
			<input
				class="rounded border border-neutral-300 px-3 py-2"
				type="email"
				name="email"
				required
				bind:value={email}
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
			{busy ? 'Sending…' : 'Send code'}
		</button>
	</form>
</main>
