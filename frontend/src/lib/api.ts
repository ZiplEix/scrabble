import { supabase } from "./supabase";
import { user } from "./stores/user";
import { get } from "svelte/store";
import type { GameSummary } from "./types/game_summary";

// Helpers to get current session details
const getUserId = (): number | null => {
    const u = get(user);
    return u?.id ? Number(u.id) : null;
};

const getUsername = (): string => {
    const u = get(user);
    return u?.username || "";
};

// Helper to throw a standardized error compatible with the frontend's expected error format
const throwApiError = (message: string, status = 500): never => {
    const err: any = new Error(message);
    err.response = {
        status,
        data: {
            message,
            error: message
        }
    };
    throw err;
};

function normalizeWordsPlayed(words: any): any[] {
    if (!Array.isArray(words)) return [];
    return words.map((w: any) => ({
        word: w.word ?? w.Word ?? "",
        position: w.position ?? w.Position ?? "",
        direction: w.direction ?? w.Direction ?? "",
        score: w.score ?? w.Score ?? 0
    }));
}

// ==========================================
// 1. PARTIES / GAMES
// ==========================================

export async function getGames(): Promise<GameSummary[]> {
    const currentUserId = getUserId();
    if (!currentUserId) return [];

    // Select player's participations
    const { data: gp, error: gpError } = await supabase
        .from("game_players")
        .select("game_id")
        .eq("player_id", currentUserId);

    if (gpError || !gp || gp.length === 0) return [];

    const gameIds = gp.map(g => g.game_id);

    // Select games details
    const { data: dbGames, error: gError } = await supabase
        .from("games")
        .select(`
      id, name, status, current_turn, winner_username, created_at, created_by,
      current_turn_user:users!games_current_turn_fkey ( username )
    `)
        .in("id", gameIds)
        .order("created_at", { ascending: false });

    if (gError || !dbGames) return [];

    // Fetch last move times
    const { data: moves } = await supabase
        .from("game_moves")
        .select("game_id, created_at")
        .in("game_id", gameIds)
        .order("created_at", { ascending: false });

    const lastMoveMap: Record<string, string> = {};
    moves?.forEach(m => {
        if (!lastMoveMap[m.game_id]) {
            lastMoveMap[m.game_id] = m.created_at;
        }
    });

    return dbGames.map(g => {
        const turnUser = g.current_turn_user as any;
        return {
            id: g.id,
            name: g.name,
            status: g.status,
            current_turn_user_id: g.current_turn,
            current_turn_username: turnUser?.username || "Ordinateur",
            winner_username: g.winner_username || "",
            last_play_time: lastMoveMap[g.id] || g.created_at,
            is_your_game: g.created_by === currentUserId
        };
    });
}

export async function getGame(gameId: string): Promise<any> {
    const currentUserId = getUserId();
    const { data: game, error: gError } = await supabase
        .from("games")
        .select("*")
        .eq("id", gameId)
        .single();

    if (gError || !game) throwApiError("Game not found", 404);

    const { data: players, error: pError } = await supabase
        .from("game_players")
        .select("player_id, score, position, rack, users(username, is_bot)")
        .eq("game_id", gameId)
        .order("position", { ascending: true });

    if (pError || !players) throwApiError("Players not found", 404);

    const { data: moves } = await supabase
        .from("game_moves")
        .select("player_id, move, created_at")
        .eq("game_id", gameId)
        .order("created_at", { ascending: true });

    // Find current user rack
    const me = (players || []).find(p => p.player_id === currentUserId);
    const yourRack = me?.rack || "";

    // Blank tiles (jokers) position tracking
    const blankTiles: { x: number; y: number }[] = [];
    moves?.forEach(m => {
        const mv = m.move as any;
        if (mv && Array.isArray(mv.letters)) {
            mv.letters.forEach((l: any) => {
                if (l.blank) blankTiles.push({ x: l.x, y: l.y });
            });
        }
    });

    const currentTurnPlayer = (players || []).find(p => p.player_id === game.current_turn);

    return {
        id: game.id,
        name: game.name,
        status: game.status,
        current_turn: game.current_turn,
        current_turn_name: currentTurnPlayer?.users ? (currentTurnPlayer.users as any).username : "Ordinateur",
        board: game.board,
        remaining_letters: game.available_letters ? game.available_letters.length : 0,
        your_rack: yourRack,
        is_your_game: game.created_by === currentUserId,
        players: (players || []).map(p => ({
            id: p.player_id,
            username: (p.users as any)?.username || "Joueur",
            score: p.score,
            position: p.position,
            is_bot: (p.users as any)?.is_bot || false,
            rack: p.player_id === currentUserId ? p.rack : ""
        })),
        moves: (moves || []).map(m => ({
            player_id: m.player_id,
            move: m.move,
            played_at: m.created_at
        })),
        blank_tiles: blankTiles,
        winner_username: game.winner_username || "",
        ended_at: game.ended_at,
        pass_count: game.pass_count
    };
}

export async function createGame(name: string, players: string[], difficulty = "hard"): Promise<string> {
    const { data, error } = await supabase.rpc("rpc_create_game", {
        p_name: name,
        p_opponent_usernames: players,
        p_difficulty: difficulty
    });

    if (error) throwApiError(error.message);
    return data as string;
}

export async function deleteGame(gameId: string): Promise<void> {
    const { error } = await supabase.from("games").delete().eq("id", gameId);
    if (error) throwApiError(error.message);
}

export async function renameGame(gameId: string, newName: string): Promise<void> {
    const { error } = await supabase
        .from("games")
        .update({ name: newName })
        .eq("id", gameId);

    if (error) throwApiError(error.message);
}

export async function playWord(gameId: string, playData: { letters: any[]; word?: string; x?: number; y?: number; dir?: string }): Promise<any> {
    const { data, error } = await supabase.functions.invoke("game-action", {
        body: {
            action: "play",
            game_id: gameId,
            letters: playData.letters,
            word: playData.word,
            x: playData.x,
            y: playData.y,
            dir: playData.dir
        }
    });

    if (error || data?.error) throwApiError(data?.error || error?.message || "Move failed");
    return data;
}

export async function simulateScore(gameId: string, letters: any[]): Promise<any> {
    const { data, error } = await supabase.functions.invoke("game-action", {
        body: {
            action: "simulate_score",
            game_id: gameId,
            letters
        }
    });

    if (error || data?.error) throwApiError(data?.error || error?.message || "Simulation failed");
    return data;
}

export async function passTurn(gameId: string): Promise<void> {
    const currentUserId = getUserId();
    const { error } = await supabase.rpc("rpc_pass_turn", {
        p_game_id: gameId,
        p_user_id: currentUserId
    });

    if (error) throwApiError(error.message);
}

export async function getNewRack(gameId: string): Promise<string[]> {
    const currentUserId = getUserId();
    const { data, error } = await supabase.rpc("rpc_exchange_rack", {
        p_game_id: gameId,
        p_user_id: currentUserId
    });

    if (error) throwApiError(error.message);
    return (data as string).split("");
}

// ==========================================
// 2. MESSAGERIE / CHAT
// ==========================================

export async function getUnreadMessagesCount(gameId: string): Promise<number> {
    const currentUserId = getUserId();
    if (!currentUserId) return 0;

    const { data: gmr } = await supabase
        .from("game_message_reads")
        .select("last_read_message_id")
        .eq("game_id", gameId)
        .eq("user_id", currentUserId)
        .maybeSingle();

    const lastReadId = gmr?.last_read_message_id ? Number(gmr.last_read_message_id) : 0;

    const { count, error } = await supabase
        .from("messages")
        .select("*", { count: "exact", head: true })
        .eq("game_id", gameId)
        .gt("id", lastReadId);

    if (error) throwApiError(error.message);
    return count || 0;
}

export async function getMessages(gameId: string): Promise<any[]> {
    const { data, error } = await supabase
        .from("messages")
        .select("id, game_id, user_id, content, meta, created_at, users(username)")
        .eq("game_id", gameId)
        .order("created_at", { ascending: true });

    if (error) throwApiError(error.message);

    return (data || []).map(m => ({
        id: m.id,
        game_id: m.game_id,
        user_id: m.user_id,
        content: m.content,
        meta: m.meta,
        created_at: m.created_at,
        username: (m.users as any)?.username || "Joueur"
    }));
}

export async function sendMessage(gameId: string, content: string, meta = {}): Promise<any> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    const { data, error } = await supabase
        .from("messages")
        .insert({
            game_id: gameId,
            user_id: currentUserId,
            content,
            meta
        })
        .select()
        .single();

    if (error) throwApiError(error.message);
    return data;
}

export async function markMessagesAsRead(gameId: string, lastMessageId: number): Promise<void> {
    const currentUserId = getUserId();
    if (!currentUserId) return;

    const { error } = await supabase
        .from("game_message_reads")
        .upsert({
            user_id: currentUserId,
            game_id: gameId,
            last_read_message_id: lastMessageId,
            last_read_at: new Date().toISOString()
        });

    if (error) throwApiError(error.message);
}

export async function deleteChatMessage(gameId: string, messageId: number): Promise<void> {
    const { error } = await supabase
        .from("messages")
        .delete()
        .eq("id", messageId);

    if (error) throwApiError(error.message);
}

// ==========================================
// 3. PROFILS / STATISTIQUES / CLASSEMENT
// ==========================================

export async function getLeaderboard(limit = 50, offset = 0): Promise<any[]> {
    const { data, error } = await supabase
        .from("users")
        .select(`
      id, username, rating, is_bot,
      game_players(count)
    `)
        .eq("is_bot", false)
        .order("rating", { ascending: false })
        .range(offset, offset + limit - 1);

    if (error) throwApiError(error.message);
    return (data || []).map((u: any, index: number) => ({
        rank: offset + index + 1,
        user_id: u.id,
        username: u.username,
        rating: u.rating,
        games: u.game_players?.[0]?.count || 0
    }));
}

export async function getUserStats(): Promise<any> {
    const currentUserId = getUserId();
    const currentUsername = getUsername();
    if (!currentUserId) return {};

    const { data: userRow } = await supabase.from("users").select("rating").eq("id", currentUserId).single();
    const { data: playerGames } = await supabase.from("game_players").select("score, games(status, winner_username)").eq("player_id", currentUserId);

    let gamesPlayed = 0;
    let gamesWon = 0;
    let highestScore = 0;

    playerGames?.forEach(g => {
        const game = g.games as any;
        if (game?.status === "ended") {
            gamesPlayed++;
            if (game.winner_username === currentUsername) {
                gamesWon++;
            }
        }
        if (g.score > highestScore) highestScore = g.score;
    });

    return {
        games_played: gamesPlayed,
        games_won: gamesWon,
        highest_score: highestScore,
        rating: userRow?.rating || 1600
    };
}

export async function getUserAchievements(): Promise<any[]> {
    const currentUserId = getUserId();
    if (!currentUserId) return [];

    const { data, error } = await supabase
        .from("user_achievements")
        .select("achievement_id, unlocked_at, achievements(*)")
        .eq("user_id", currentUserId);

    if (error) throwApiError(error.message);

    return (data || []).map(ua => {
        const ach = ua.achievements as any;
        return {
            id: ach.id,
            title: ach.title,
            description: ach.description,
            badge_icon: ach.badge_icon,
            category: ach.category,
            unlocked_at: ua.unlocked_at
        };
    });
}

export async function getUserRatingHistory(userId: number, limit = 25): Promise<any[]> {
    const { data, error } = await supabase
        .from("user_rating_history")
        .select(`
      rating, created_at,
      games ( id, name, status, winner_username, ended_at )
    `)
        .eq("user_id", userId)
        .order("created_at", { ascending: false })
        .limit(limit);

    if (error) throwApiError(error.message);

    const { data: targetUser } = await supabase
        .from("users")
        .select("username")
        .eq("id", userId)
        .single();
    const targetUsername = targetUser?.username || "";

    const history = (data || []).map((rh: any) => {
        const game = rh.games;
        return {
            rating: rh.rating,
            created_at: rh.created_at,
            game_info: game ? {
                id: game.id,
                name: game.name,
                won: game.winner_username === targetUsername
            } : null
        };
    });

    history.reverse();
    return history;
}

export async function getCurrentUserProfile(): Promise<any> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    const { data: stats, error: statsErr } = await supabase.rpc("rpc_get_user_stats", { p_user_id: currentUserId });
    if (statsErr) throwApiError(statsErr.message);

    // Load achievements
    const { data: userAchs } = await supabase
        .from("user_achievements")
        .select("achievement_id, unlocked_at")
        .eq("user_id", currentUserId);

    const { data: allAchs } = await supabase
        .from("achievements")
        .select("id, title, description, badge_icon, category")
        .order("title", { ascending: true });

    const unlockedSet = new Map(userAchs?.map(ua => [ua.achievement_id, ua.unlocked_at]) || []);
    const achievements = (allAchs || []).map(a => ({
        id: a.id,
        title: a.title,
        description: a.description,
        badge_icon: a.badge_icon,
        category: a.category,
        unlocked: unlockedSet.has(a.id),
        unlocked_at: unlockedSet.get(a.id) || null
    }));

    return {
        ...stats,
        achievements
    };
}

export async function getUserProfile(profileId: number): Promise<any> {
    const currentUserId = getUserId();
    if (!profileId) throwApiError("Invalid user ID", 400);

    // Get stats
    const { data: stats, error: statsErr } = await supabase.rpc("rpc_get_user_stats", { p_user_id: profileId });
    if (statsErr) throwApiError(statsErr.message);

    // Check if friend
    let isFriend = false;
    if (currentUserId && currentUserId !== profileId) {
        const { data: friendRow } = await supabase
            .from("user_friends")
            .select("friend_id")
            .eq("user_id", currentUserId)
            .eq("friend_id", profileId)
            .maybeSingle();
        isFriend = !!friendRow;
    }

    // Get head-to-head
    let headToHead = null;
    if (currentUserId && currentUserId !== profileId) {
        const { data: h2hData, error: h2hErr } = await supabase.rpc("rpc_get_head_to_head", {
            p_user_id: profileId,
            p_viewer_id: currentUserId
        });
        if (!h2hErr) {
            headToHead = h2hData;
        }
    }

    // Load achievements
    const { data: userAchs } = await supabase
        .from("user_achievements")
        .select("achievement_id, unlocked_at")
        .eq("user_id", profileId);

    const { data: allAchs } = await supabase
        .from("achievements")
        .select("id, title, description, badge_icon, category")
        .order("title", { ascending: true });

    const unlockedSet = new Map(userAchs?.map(ua => [ua.achievement_id, ua.unlocked_at]) || []);
    const achievements = (allAchs || []).map(a => ({
        id: a.id,
        title: a.title,
        description: a.description,
        badge_icon: a.badge_icon,
        category: a.category,
        unlocked: unlockedSet.has(a.id),
        unlocked_at: unlockedSet.get(a.id) || null
    }));

    return {
        ...stats,
        is_friend: isFriend,
        head_to_head: headToHead,
        achievements
    };
}

export async function impersonateUser(username: string): Promise<{ token: string }> {
    if (!username) throwApiError("User parameter is required", 400);

    const { data, error } = await supabase.functions.invoke("migrate-user", {
        body: { action: "connect-as", username }
    });

    if (error || data?.error) {
        throwApiError(data?.error || error?.message || "Failed to impersonate", 400);
    }

    return { token: data.token };
}

export async function adminChangePassword(targetUsername: string, newPassword: string): Promise<void> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    // Check if caller is admin
    const { data: callerProfile, error: callerProfileError } = await supabase
        .from("users")
        .select("role")
        .eq("id", currentUserId)
        .single();

    if (callerProfileError || callerProfile?.role !== "admin") {
        throwApiError("Forbidden: Admins only", 403);
    }

    // Find target user
    const { data: targetUser, error: targetError } = await supabase
        .from("users")
        .select("uuid")
        .eq("username", targetUsername)
        .single();

    if (targetError || !targetUser || !targetUser.uuid) {
        throwApiError("Target user not found or not yet migrated to auth", 404);
    }

    // Update password via admin API
    const { error: updateError } = await supabase.auth.admin.updateUserById(targetUser!.uuid, {
        password: newPassword
    });

    if (updateError) {
        throwApiError(updateError.message, 400);
    }
}

export async function updateNotificationPrefs(prefs: { turn?: boolean; messages?: boolean }): Promise<void> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    // Fetch current preferences
    const { data: userRow } = await supabase
        .from("users")
        .select("notification_prefs")
        .eq("id", currentUserId)
        .single();

    const currentPrefs = userRow?.notification_prefs || { turn: true, messages: true };

    const newPrefs = {
        turn: prefs.turn !== undefined ? prefs.turn : currentPrefs.turn,
        messages: prefs.messages !== undefined ? prefs.messages : currentPrefs.messages
    };

    const { error } = await supabase
        .from("users")
        .update({ notification_prefs: newPrefs })
        .eq("id", currentUserId);

    if (error) throwApiError(error.message);
}

// ==========================================
// 4. RELATIONS AMICALES / FRIENDS
// ==========================================

export async function getFriends(): Promise<any[]> {
    const currentUserId = getUserId();
    if (!currentUserId) return [];

    const { data, error } = await supabase
        .from("user_friends")
        .select("friend_id, friend:users!user_friends_friend_id_fkey(id, username, rating, role)")
        .eq("user_id", currentUserId);

    if (error) throwApiError(error.message);

    const list = (data || []).map(d => d.friend).filter(Boolean);
    list.sort((a: any, b: any) => a.username.localeCompare(b.username));
    return list;
}

export async function addFriend(friendId: number): Promise<void> {
    const currentUserId = getUserId();
    const { error } = await supabase
        .from("user_friends")
        .insert({
            user_id: currentUserId,
            friend_id: friendId
        });

    if (error) throwApiError(error.message);
}

export async function removeFriend(friendId: number): Promise<void> {
    const currentUserId = getUserId();
    const { error } = await supabase
        .from("user_friends")
        .delete()
        .eq("user_id", currentUserId)
        .eq("friend_id", friendId);

    if (error) throwApiError(error.message);
}

export async function getRecentOpponents(): Promise<any[]> {
    const currentUserId = getUserId();
    if (!currentUserId) return [];

    const { data, error } = await supabase.rpc("rpc_get_recent_opponents");
    if (error) throwApiError(error.message);
    return data || [];
}

export async function suggestUsers(q: string): Promise<any[]> {
    const currentUserId = getUserId();
    if (q.length < 2) return [];

    const { data, error } = await supabase
        .from("users")
        .select("id, username")
        .neq("id", currentUserId)
        .ilike("username", `${q}%`)
        .limit(10);

    if (error) throwApiError(error.message);
    return data || [];
}

// ==========================================
// 5. ENTRAÎNEMENTS / PUZZLES
// ==========================================

export async function getPuzzles(limit = 20): Promise<any[]> {
    const currentUserId = getUserId();
    if (!currentUserId) return [];

    const { data: puzzles, error: pError } = await supabase
        .from("daily_puzzles")
        .select("*")
        .order("puzzle_date", { ascending: false })
        .limit(limit);

    if (pError || !puzzles) return [];

    const { data: attempts } = await supabase
        .from("puzzle_attempts")
        .select("*")
        .eq("player_id", currentUserId);

    const activeAttempts = (attempts || []).filter(a => puzzles.some(p => p.id === a.puzzle_id));

    const enrichedAttempts = await Promise.all(
        activeAttempts.map(async (attempt) => {
            const scoreVal = attempt.score ?? 0;
            const timeVal = attempt.time_used ?? 0;
            const { count, error } = await supabase
                .from("puzzle_attempts")
                .select("*", { count: "exact", head: true })
                .eq("puzzle_id", attempt.puzzle_id)
                .not("submitted_at", "is", null)
                .or(`score.gt.${scoreVal},and(score.eq.${scoreVal},time_used.lt.${timeVal})`);

            if (error) {
                console.error("getPuzzles count query error:", error);
            }
            const rank = (count ?? 0) + 1;

            return {
                ...attempt,
                rank_today: rank
            };
        })
    );

    const attemptMap = Object.fromEntries(enrichedAttempts.map(a => [a.puzzle_id, a]));

    return puzzles.map(p => {
        const attempt = attemptMap[p.id];
        return {
            id: p.id,
            puzzle_date: p.puzzle_date,
            level: p.level,
            has_attempted: !!(attempt && attempt.submitted_at),
            player_attempt: attempt ? {
                id: attempt.id,
                puzzle_id: attempt.puzzle_id,
                player_id: attempt.player_id,
                score: attempt.score,
                time_used_secs: attempt.time_used,
                submitted_at: attempt.submitted_at,
                words_played: normalizeWordsPlayed(attempt.words_played),
                rank_today: attempt.rank_today
            } : null
        };
    });
}

export async function getTodayPuzzle(): Promise<any> {
    const currentUserId = getUserId();
    const todayStr = new Date().toISOString().slice(0, 10);
    const { data, error } = await supabase
        .from("daily_puzzles")
        .select("*")
        .eq("puzzle_date", todayStr)
        .maybeSingle();

    if (error || !data) {
        throwApiError("Puzzle du jour indisponible, veuillez patienter.", 404);
    }

    // Check if player has already attempted today's puzzle
    let hasAttempted = false;
    if (currentUserId) {
        const { data: attempt } = await supabase
            .from("puzzle_attempts")
            .select("submitted_at")
            .eq("puzzle_id", data.id)
            .eq("player_id", currentUserId)
            .maybeSingle();
        hasAttempted = !!(attempt && attempt.submitted_at);
    }

    return {
        id: data.id,
        puzzle_date: data.puzzle_date,
        level: data.level,
        board: data.board,
        available_letters: data.available_letters,
        timeout_seconds: data.level === 3 ? 420 : data.level === 2 ? 300 : 180,
        has_player_attempted: hasAttempted,
        created_at: data.created_at
    };
}

export async function getPuzzleLeaderboard(puzzleId: string): Promise<any[]> {
    const { data, error } = await supabase
        .from("puzzle_attempts")
        .select("score, time_used, submitted_at, words_played, users(id, username)")
        .eq("puzzle_id", puzzleId)
        .not("submitted_at", "is", null)
        .order("score", { ascending: false })
        .order("time_used", { ascending: true })
        .limit(50);

    if (error) throwApiError(error.message);

    return (data || []).map((d, index) => ({
        rank: index + 1,
        player_id: (d.users as any).id,
        username: (d.users as any).username,
        score: d.score,
        time_used: d.time_used,
        submitted_at: d.submitted_at,
        words_played: normalizeWordsPlayed(d.words_played)
    }));
}

export async function getPuzzle(puzzleId: string): Promise<any> {
    const currentUserId = getUserId();
    const { data, error } = await supabase
        .from("daily_puzzles")
        .select("*")
        .eq("id", puzzleId)
        .single();

    if (error) throwApiError(error.message);

    let hasAttempted = false;
    if (currentUserId) {
        const { data: attempt } = await supabase
            .from("puzzle_attempts")
            .select("submitted_at")
            .eq("puzzle_id", puzzleId)
            .eq("player_id", currentUserId)
            .maybeSingle();
        hasAttempted = !!(attempt && attempt.submitted_at);
    }

    return {
        id: data.id,
        puzzle_date: data.puzzle_date,
        level: data.level,
        board: data.board,
        available_letters: data.available_letters,
        timeout_seconds: data.level === 3 ? 420 : data.level === 2 ? 300 : 180,
        has_player_attempted: hasAttempted,
        created_at: data.created_at
    };
}

export async function startPuzzleAttempt(puzzleId: string): Promise<any> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    const { data: existing } = await supabase
        .from("puzzle_attempts")
        .select("*")
        .eq("puzzle_id", puzzleId)
        .eq("player_id", currentUserId)
        .maybeSingle();

    if (existing) {
        if (existing.submitted_at) {
            throwApiError("Vous avez déjà soumis ce puzzle.", 400);
        }
        return {
            attempt_id: existing.id,
            started_at: existing.started_at,
            already_started: true
        };
    }

    const attemptId = crypto.randomUUID();
    const now = new Date().toISOString();

    const { error } = await supabase
        .from("puzzle_attempts")
        .insert({
            id: attemptId,
            puzzle_id: puzzleId,
            player_id: currentUserId,
            started_at: now,
            created_at: now
        });

    if (error) throwApiError(error.message);

    return {
        attempt_id: attemptId,
        started_at: now,
        already_started: false
    };
}

export async function submitPuzzleAttempt(puzzleId: string, letters?: any[]): Promise<any> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    let data: any;

    if (!letters || letters.length === 0) {
        // Direct submission of 0 score (timeout or skipped)
        const { data: attempt } = await supabase
            .from("puzzle_attempts")
            .select("id, started_at")
            .eq("puzzle_id", puzzleId)
            .eq("player_id", currentUserId)
            .maybeSingle();

        const now = new Date();
        const startedAt = attempt ? new Date(attempt.started_at) : now;
        const timeUsed = Math.floor((now.getTime() - startedAt.getTime()) / 1000);

        const attemptId = attempt?.id || crypto.randomUUID();
        const { error } = await supabase
            .from("puzzle_attempts")
            .upsert({
                id: attemptId,
                puzzle_id: puzzleId,
                player_id: currentUserId,
                score: 0,
                words_played: [],
                time_used: timeUsed,
                submitted_at: now.toISOString()
            });

        if (error) throwApiError(error.message);
        data = {
            id: attemptId,
            puzzle_id: puzzleId,
            player_id: currentUserId,
            score: 0,
            time_used_secs: timeUsed,
            submitted_at: now.toISOString(),
            words_played: []
        };
    } else {
        // Validate through the game-action edge function
        const { data: actionRes, error: actionErr } = await supabase.functions.invoke("game-action", {
            body: {
                action: "play",
                game_id: puzzleId,
                letters
            }
        });

        if (actionErr || actionRes?.error) {
            throwApiError(actionRes?.error || actionErr?.message || "Puzzle submission failed", 400);
        }
        data = actionRes;
    }

    const timeUsedSecs = data.time_used_secs ?? data.time_used ?? 0;

    // Calculate rank_today
    const { count, error } = await supabase
        .from("puzzle_attempts")
        .select("*", { count: "exact", head: true })
        .eq("puzzle_id", puzzleId)
        .not("submitted_at", "is", null)
        .or(`score.gt.${data.score},and(score.eq.${data.score},time_used.lt.${timeUsedSecs})`);

    if (error) {
        console.error("submitPuzzleAttempt count query error:", error);
    }
    const rank = (count ?? 0) + 1;

    return {
        id: data.id,
        puzzle_id: puzzleId,
        player_id: currentUserId,
        score: data.score,
        time_used_secs: timeUsedSecs,
        submitted_at: data.submitted_at,
        words_played: normalizeWordsPlayed(data.words_played),
        rank_today: rank
    };
}

// ==========================================
// 6. DICTIONNAIRE / DICTIONARY
// ==========================================

export async function getWordDefinition(word: string): Promise<any> {
    const wordUpper = word.toUpperCase();
    const { data, error } = await supabase
        .from("dictionary_definitions")
        .select("definitions")
        .eq("word", wordUpper)
        .maybeSingle();

    if (error) throwApiError(error.message);
    const defs = data?.definitions;
    if (!defs) throwApiError("Definition not found", 404);

    return defs;
}

export async function addWordDefinition(word: string, definitions: any): Promise<void> {
    const { error } = await supabase
        .from("dictionary_definitions")
        .upsert({
            word: word.toUpperCase(),
            definitions,
            created_at: new Date().toISOString()
        });

    if (error) throwApiError(error.message);
}

// ==========================================
// 7. SIGNALEMENTS / REPORTS
// ==========================================

export async function getMyReports(): Promise<any[]> {
    const currentUserId = getUserId();
    if (!currentUserId) return [];

    const { data, error } = await supabase
        .from("reports")
        .select("id, title, content, status, priority, type, created_at, updated_at, users(username)")
        .eq("user_id", currentUserId)
        .order("created_at", { ascending: false });

    if (error) throwApiError(error.message);

    return (data || []).map((r: any) => ({
        id: r.id,
        title: r.title,
        content: r.content,
        status: r.status,
        priority: r.priority,
        type: r.type,
        created_at: r.created_at,
        updated_at: r.updated_at,
        username: r.users?.username || "Moi"
    }));
}

export async function getAllReports(): Promise<any[]> {
    const { data, error } = await supabase
        .from("reports")
        .select("id, title, content, status, priority, type, created_at, updated_at, users(username)")
        .order("created_at", { ascending: false });

    if (error) throwApiError(error.message);

    return (data || []).map((r: any) => ({
        id: r.id,
        title: r.title,
        content: r.content,
        status: r.status,
        priority: r.priority,
        type: r.type,
        created_at: r.created_at,
        updated_at: r.updated_at,
        username: r.users?.username || "Utilisateur supprimé"
    }));
}

export async function getReport(reportId: number): Promise<any> {
    const { data, error } = await supabase
        .from("reports")
        .select("id, title, content, status, priority, type, created_at, updated_at, users(username)")
        .eq("id", reportId)
        .single();

    if (error || !data) throwApiError(error?.message || "Report not found", 404);

    const r = data as any;
    const userObj = r.users;
    const username = Array.isArray(userObj) ? userObj[0]?.username : userObj?.username;

    return {
        id: r.id,
        title: r.title,
        content: r.content,
        status: r.status,
        priority: r.priority,
        type: r.type,
        created_at: r.created_at,
        updated_at: r.updated_at,
        username: username || "Utilisateur"
    };
}

export async function createReport(title: string, content: string, type: string): Promise<void> {
    const currentUserId = getUserId();
    const { error } = await supabase
        .from("reports")
        .insert({
            user_id: currentUserId,
            title,
            content,
            type
        });

    if (error) throwApiError(error.message);
}

export async function updateReportStatus(reportId: number, payload: { title?: string; content?: string; status?: string; type?: string }): Promise<void> {
    const updates: any = {};
    if (payload.title !== undefined) updates.title = payload.title;
    if (payload.content !== undefined) updates.content = payload.content;
    if (payload.status !== undefined) updates.status = payload.status;
    if (payload.type !== undefined) updates.type = payload.type;
    updates.updated_at = new Date().toISOString();

    const { error } = await supabase
        .from("reports")
        .update(updates)
        .eq("id", reportId);

    if (error) throwApiError(error.message);
}

// ==========================================
// 8. PUSH NOTIFICATIONS
// ==========================================

export async function subscribeToPush(subscription: any): Promise<void> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    const { error } = await supabase
        .from("push_subscriptions")
        .upsert({
            user_id: currentUserId,
            subscription,
            created_at: new Date().toISOString()
        });

    if (error) throwApiError(error.message);
}

export async function unsubscribeFromPush(): Promise<void> {
    const currentUserId = getUserId();
    if (!currentUserId) throwApiError("Unauthorized", 401);

    const { error } = await supabase
        .from("push_subscriptions")
        .delete()
        .eq("user_id", currentUserId);

    if (error) throwApiError(error.message);
}
