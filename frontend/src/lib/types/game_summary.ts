export type GameSummary = {
    id: string;
    name: string;
    current_turn_username: string;
    last_play_time: string;
    is_your_game: boolean;
    status: string;
    winner_username: string | null;
}