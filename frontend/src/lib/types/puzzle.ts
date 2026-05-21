export type PuzzleInfo = {
	id: string;
	puzzle_date: string;
	level: number;
	board: any;
	available_letters: string;
	timeout_seconds: number;
	has_player_attempted: boolean;
	created_at: string;
};

export type PuzzleAttempt = {
	id: string;
	puzzle_id: string;
	player_id: number;
	started_at: string;
	score: number;
	words_played: PuzzleWordRecord[];
	time_used_secs: number; // calculé côté serveur
	rank_today: number;
	submitted_at?: string;
	created_at: string;
};

export type PuzzleStarted = {
	attempt_id: string;
	started_at: string;
	timeout_seconds: number;
	already_started: boolean;
};

export type PuzzleWordRecord = {
	word: string;
	position: string;
	direction: string;
	score: number;
};

export type PuzzleDailyLeaderboard = {
	rank: number;
	player_id: number;
	username: string;
	score: number;
	time_used: number;
	attempts: number;
	words_played?: PuzzleWordRecord[];
	submitted_at: string;
};

export type PuzzleHistory = {
	id: string;
	puzzle_date: string;
	level: number;
	has_attempted: boolean;
	player_attempt?: PuzzleAttempt;
	day_leaderboard?: PuzzleDailyLeaderboard[];
};

export type PuzzleStats = {
	total_attempts: number;
	best_score: number;
	average_score: number;
	completed_puzzles: number;
};
