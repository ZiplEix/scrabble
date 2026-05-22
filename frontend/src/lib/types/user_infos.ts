export type UserInfos = {
    id: number;
    username: string;
    rating: number;
    role: string;
    created_at: string;
    games_count: number;
    games_count_top_percent?: number;
    best_score: number;
    best_score_top_percent?: number;
    victories: number;
    victories_top_percent?: number;
    puzzle_wins?: number;
    puzzle_wins_top_percent?: number;
    avg_score: number;
    avg_score_top_percent?: number;
    avg_points_per_move?: number;
    avg_points_per_move_top_percent?: number;
    best_move_score?: number;
    best_move_score_top_percent?: number;
    notifications_enabled: boolean;
    turn_notifications_enabled?: boolean;
    messages_notifications_enabled?: boolean;
    head_to_head?: HeadToHeadInfo;
    achievements?: AchievementResponse[];
    is_friend?: boolean;
}

export interface FriendInfo {
    id: number;
    username: string;
    rating: number;
    role: string;
}

export type AchievementResponse = {
    id: string;
    title: string;
    description: string;
    badge_icon: string;
    category: string;
    unlocked: boolean;
    unlocked_at?: string;
}

export type CommonGameSummary = {
    id: string;
    name: string;
    status: string;
    winner: string;
    user_score: number;
    opp_score: number;
    created_at: string;
}

export type HeadToHeadInfo = {
    games_played: number;
    user_wins: number;
    opponent_wins: number;
    user_avg_score: number;
    opp_avg_score: number;
    recent_games: CommonGameSummary[];
}

export let defaultUserInfos = {
    id: 0,
    username: '',
    rating: 1600,
    role: 'user',
    created_at: new Date().toISOString(),
    games_count: 0,
    best_score: 0,
    victories: 0,
    puzzle_wins: 0,
    avg_score: 0,
    avg_points_per_move: 0,
    best_move_score: 0,
    // top percent fields intentionally left undefined by default
    notifications_enabled: false,
    turn_notifications_enabled: true,
    messages_notifications_enabled: true
}

export interface RatingHistoryGameInfo {
    game_id: string;
    opponent_username: string;
    user_score: number;
    opponent_score: number;
    won: boolean;
    ended_at: string;
}

export interface RatingHistoryResponse {
    rating: number;
    created_at: string;
    game_info?: RatingHistoryGameInfo;
}
