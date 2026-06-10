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

async function simulateSubmitPuzzleAttempt(puzzleId, currentUserId) {
    // 1. Fetch the actual attempt from DB to get the score and time_used
    const { data: attempt } = await supabase
        .from("puzzle_attempts")
        .select("*")
        .eq("puzzle_id", puzzleId)
        .eq("player_id", currentUserId)
        .maybeSingle();

    if (!attempt) {
        console.error("Attempt not found");
        return null;
    }

    const timeUsedSecs = attempt.time_used;
    const score = attempt.score;

    // Calculate rank_today
    const { count, error } = await supabase
        .from("puzzle_attempts")
        .select("*", { count: "exact", head: true })
        .eq("puzzle_id", puzzleId)
        .not("submitted_at", "is", null)
        .or(`score.gt.${score},and(score.eq.${score},time_used.lt.${timeUsedSecs})`);

    if (error) {
        console.error("Count query error:", error);
    }

    return {
        id: attempt.id,
        puzzle_id: puzzleId,
        player_id: currentUserId,
        score: score,
        time_used_secs: timeUsedSecs,
        submitted_at: attempt.submitted_at,
        rank_today: (count ?? 0) + 1
    };
}

async function run() {
    const puzzleId = '6ac119bb-03e3-4a6e-b7b4-2cbc85f8bd57';
    const res = await simulateSubmitPuzzleAttempt(puzzleId, 1);
    console.log("submitPuzzleAttempt simulated result:", JSON.stringify(res, null, 2));
}

run();
