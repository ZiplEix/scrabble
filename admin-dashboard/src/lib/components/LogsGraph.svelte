<script lang='ts'>
    import { api } from "$lib/api";
    import { onMount } from "svelte";
    import MinimalLineChart from "./MinimalLineChart.svelte";

    // chart data coming from API: will be filled on mount
    let points: number[] = [];
    let labels: string[] | null = null;

    onMount(async () => {
        try {
            const res = await api.get('/admin/stats/logs');
            // API returns { labels: [ISO timestamps], data: [ints] }
            if (res && res.data) {
                const d = res.data;
                // d.labels are ISO timestamps (UTC) for each hour; convert to localized 'HHh'
                if (Array.isArray(d.labels) && Array.isArray(d.data)) {
                    points = d.data.map((v: any) => Number(v));
                    labels = d.labels.map((iso: string) => {
                        try {
                            const dt = new Date(iso);
                            return dt.getHours().toString().padStart(2, '0') + 'h';
                        } catch (e) {
                            return iso;
                        }
                    });
                }
            }
        } catch (err) {
            console.error('Failed to fetch dashboard data:', err);
        }
    });
</script>

<div class="bg-white/4 rounded-lg p-4">
    <div class="w-full">
        <MinimalLineChart height={200} data={points} labels={labels} xTickStep={6} color="#60a5fa" strokeWidth={1.6} />
    </div>
</div>
