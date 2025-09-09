<script lang='ts'>
    import { api } from "$lib/api";
    import { onMount } from "svelte";
    import LogsTable from "./LogsTable.svelte";

    type LogEntry = { level: 'info' | 'warn' | 'error'; date: string; route: string; message: string };
    let logs: LogEntry[] = [];
    let loading: boolean = true;

    onMount(async () => {
        try {
            const res = await api.get('/admin/logs/resume');
            if (res && res.data) {
                // API returns array of entries with at least level, date, route, message
                logs = res.data.map((e: any) => {
                    const level = (e.level as 'info' | 'warn' | 'error') || 'info';
                    const date = (e.date) || new Date().toISOString();
                    const route = e.route || (e.raw && (e.raw.route || e.raw.path)) || '';
                    const rawMsg = e.message || (e.raw && (e.raw.msg || e.raw.message));
                    const reason = e.raw && (e.raw.reason || e.raw.error || e.raw.cause || null);
                    const message = (level !== 'info' && reason) ? reason : (rawMsg || JSON.stringify(e.raw || {}));
                    return { level, date, route, message };
                });
            }
        } catch (err) {
            console.error('failed to fetch logs resume', err);
        } finally {
            loading = false;
        }
    });
</script>

<LogsTable {logs} {loading} />
