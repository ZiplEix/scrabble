<script lang="ts">
    import { user } from "$lib/stores/user";
    export let id: number | null | undefined;
    export let username: string = '';
</script>

{#if id && $user?.id !== id}
    <a href={'/user/' + id} class="ml-1 text-gray-800 font-semibold hover:underline">
        {username}
    </a>
{:else}
    <strong class="ml-1 text-gray-800">{username}</strong>
{/if}
