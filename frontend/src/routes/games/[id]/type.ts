type PlayerInfo = {
    id: number;
    username: string;
    score: number;
    position: number;
};

type MoveInfo = {
    player_id: number;
    move: any;
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
};
