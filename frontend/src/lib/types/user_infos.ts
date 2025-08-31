export type UserInfos = {
    id: number;
    username: string;
    role: string;
    created_at: string;
    games_count: number;
    games_count_top_percent?: number;
    best_score: number;
    best_score_top_percent?: number;
    victories: number;
    victories_top_percent?: number;
    avg_score: number;
    avg_score_top_percent?: number;
    avg_points_per_move?: number;
    avg_points_per_move_top_percent?: number;
    best_move_score?: number;
    best_move_score_top_percent?: number;
    notifications_enabled: boolean;
    turn_notifications_enabled?: boolean;
    messages_notifications_enabled?: boolean;
}

export let defaultUserInfos = {
    id: 0,
    username: '',
    role: 'user',
    created_at: new Date().toISOString(),
    games_count: 0,
    best_score: 0,
    victories: 0,
    avg_score: 0,
    avg_points_per_move: 0,
    best_move_score: 0,
    // top percent fields intentionally left undefined by default
    notifications_enabled: false,
    turn_notifications_enabled: true,
    messages_notifications_enabled: true
}
