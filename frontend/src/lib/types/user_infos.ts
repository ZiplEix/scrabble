export type UserInfos = {
    id: number;
    username: string;
    role: string;
    created_at: string;
    games_count: number;
    best_score: number;
    victories: number;
    avg_score: number;
    avg_points_per_move?: number;
    best_move_score?: number;
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
    notifications_enabled: false,
    turn_notifications_enabled: true,
    messages_notifications_enabled: true
}
