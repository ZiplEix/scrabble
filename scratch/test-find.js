const fs = require('fs');
const path = require('path');
const { createClient } = require('@supabase/supabase-js');

// Parse .env manually
const envPath = path.resolve(__dirname, '../frontend/.env');
const envContent = fs.readFileSync(envPath, 'utf8');
const env = {};
envContent.split('\n').forEach(line => {
    const parts = line.split('=');
    if (parts.length >= 2) {
        const key = parts[0].trim();
        const val = parts.slice(1).join('=').trim().replace(/^"(.*)"$/, '$1');
        env[key] = val;
    }
});

const supabaseUrl = env.PUBLIC_SUPABASE_URL || '';
const supabaseAnonKey = env.PUBLIC_SUPABASE_ANON_KEY || '';

const supabase = createClient(supabaseUrl, supabaseAnonKey);

function normalizeWordsPlayed(words) {
    if (!Array.isArray(words)) return [];
    return words.map((w) => ({
        word: w.word ?? w.Word ?? "",
        position: w.position ?? w.Position ?? "",
        direction: w.direction ?? w.Direction ?? "",
        score: w.score ?? w.Score ?? 0
    }));
}

async function getPuzzles(currentUserId, limit = 20) {
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
            const { count } = await supabase
                .from("puzzle_attempts")
                .select("*", { count: "exact", head: true })
                .eq("puzzle_id", attempt.puzzle_id)
                .not("submitted_at", "is", null)
                .or(`score.gt.${scoreVal},and(score.eq.${scoreVal},time_used.lt.${timeVal})`);

            return {
                ...attempt,
                rank_today: (count ?? 0) + 1
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

async function run() {
    const puzzleId = '6ac119bb-03e3-4a6e-b7b4-2cbc85f8bd57';
    const history = await getPuzzles(1, 50);
    const item = history.find((h) => h.id === puzzleId && h.has_attempted && h.player_attempt);
    console.log("Found item:", JSON.stringify(item, null, 2));
}

run();
