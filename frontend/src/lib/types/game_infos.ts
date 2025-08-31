type PlayerInfo = {
    id: number;
    username: string;
    score: number;
    position: number;
};

type MoveData = {
    dir: 'H' | 'V';
    letters: {
        char: string;
        x: number;
        y: number;
    blank?: boolean;
    }[];
    score: number;
    word: string;
    x: number;
    y: number;
    type?: '' | 'pass';
}

type MoveInfo = {
    player_id: number;
    move: MoveData;
    played_at: string;
};

type GameInfo = {
    id: string;
    name: string;
    board: string[][];
    your_rack: string;
    players: PlayerInfo[];
    moves: MoveInfo[];
    current_turn: number;
    current_turn_username: string;
    status: string;
    remaining_letters: number;
    winner_username?: string;
    ended_at?: string;
    is_your_game: boolean;
    blank_tiles?: { x: number; y: number }[];
};
