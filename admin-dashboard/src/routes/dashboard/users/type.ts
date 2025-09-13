export type User = {
    id: number;
    username: string;
    role: string;
    created_at: string;
    notification_prefs?: Record<string, any>;
    messages_count: number;
    games_count: number;
    ongoing_games: number;
    finished_games: number;
    last_activity_at?: string;
    games: { id: string; name: string; status: string; created_at: string }[];
}