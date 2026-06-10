import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "authorization, x-client-info, apikey, content-type",
};

const LETTER_VALUES: Record<string, number> = {
  "A": 1, "B": 3, "C": 3, "D": 2, "E": 1,
  "F": 4, "G": 2, "H": 4, "I": 1, "J": 8,
  "K": 10, "L": 1, "M": 2, "N": 1, "O": 1,
  "P": 3, "Q": 8, "R": 1, "S": 1, "T": 1,
  "U": 1, "V": 4, "W": 10, "X": 10, "Y": 10, "Z": 10
};

const SPECIAL_CELLS: Record<string, string> = {
  "0,0": "TW", "0,7": "TW", "0,14": "TW",
  "7,0": "TW", "7,14": "TW",
  "14,0": "TW", "14,7": "TW", "14,14": "TW",

  "1,1": "DW", "2,2": "DW", "3,3": "DW", "4,4": "DW",
  "10,10": "DW", "11,11": "DW", "12,12": "DW", "13,13": "DW",
  "1,13": "DW", "2,12": "DW", "3,11": "DW", "4,10": "DW",
  "10,4": "DW", "11,3": "DW", "12,2": "DW", "13,1": "DW",
  "7,7": "★", // center

  "1,5": "TL", "1,9": "TL",
  "5,1": "TL", "5,5": "TL", "5,9": "TL", "5,13": "TL",
  "9,1": "TL", "9,5": "TL", "9,9": "TL", "9,13": "TL",
  "13,5": "TL", "13,9": "TL",

  "0,3": "DL", "0,11": "DL",
  "2,6": "DL", "2,8": "DL",
  "3,0": "DL", "3,7": "DL", "3,14": "DL",
  "6,2": "DL", "6,6": "DL", "6,8": "DL", "6,12": "DL",
  "7,3": "DL", "7,11": "DL",
  "8,2": "DL", "8,6": "DL", "8,8": "DL", "8,12": "DL",
  "11,0": "DL", "11,7": "DL", "11,14": "DL",
  "12,6": "DL", "12,8": "DL",
  "14,3": "DL", "14,11": "DL"
};

interface PlacedLetter {
  x: number;
  y: number;
  char: string;
  blank: boolean;
}

interface FormedWord {
  word: string;
  startX: number;
  startY: number;
  dx: number;
  dy: number;
}

function extractFormedWords(board: string[][], placed: PlacedLetter[]): FormedWord[] {
  const letterMap = new Set(placed.map(l => `${l.x},${l.y}`));
  const visited = new Set<string>();
  const words: FormedWord[] = [];
  const dirs = [{ dx: 1, dy: 0 }, { dx: 0, dy: 1 }];

  for (const l of placed) {
    for (const dir of dirs) {
      let startX = l.x;
      let startY = l.y;
      while (true) {
        const nx = startX - dir.dx;
        const ny = startY - dir.dy;
        if (nx < 0 || nx >= 15 || ny < 0 || ny >= 15 || !board[ny][nx]) {
          break;
        }
        startX = nx;
        startY = ny;
      }

      let wordText = "";
      let touchesNewTile = false;
      let x = startX;
      let y = startY;
      while (x >= 0 && x < 15 && y >= 0 && y < 15) {
        const letter = board[y][x];
        if (!letter) break;
        wordText += letter;
        if (letterMap.has(`${x},${y}`)) {
          touchesNewTile = true;
        }
        x += dir.dx;
        y += dir.dy;
      }

      if (wordText.length > 1 && touchesNewTile) {
        const key = `${startX},${startY},${dir.dx},${dir.dy}`;
        if (!visited.has(key)) {
          visited.add(key);
          words.push({
            word: wordText,
            startX,
            startY,
            dx: dir.dx,
            dy: dir.dy
          });
        }
      }
    }
  }
  return words;
}

function computeWordScore(board: string[][], fw: FormedWord, isNew: Set<string>, isBlank: Set<string>): number {
  let wordMultiplier = 1;
  let wordScore = 0;
  let x = fw.startX;
  let y = fw.startY;

  while (x >= 0 && x < 15 && y >= 0 && y < 15) {
    const letter = board[y][x];
    if (!letter) break;

    let letterScore = 0;
    const posKey = `${x},${y}`;
    if (!isBlank.has(posKey)) {
      letterScore = LETTER_VALUES[letter] ?? 0;
    }

    if (isNew.has(posKey)) {
      const cellType = SPECIAL_CELLS[`${x},${y}`];
      if (cellType === "DL") {
        letterScore *= 2;
      } else if (cellType === "TL") {
        letterScore *= 3;
      } else if (cellType === "DW" || cellType === "★") {
        wordMultiplier *= 2;
      } else if (cellType === "TW") {
        wordMultiplier *= 3;
      }
    }

    wordScore += letterScore;
    x += fw.dx;
    y += fw.dy;
  }

  return wordScore * wordMultiplier;
}

function computeMoveScore(board: string[][], placed: PlacedLetter[], boardBlank: Set<string>): number {
  const isNew = new Set<string>();
  const isBlank = new Set<string>();
  for (const l of placed) {
    const posKey = `${l.x},${l.y}`;
    isNew.add(posKey);
    if (l.blank) {
      isBlank.add(posKey);
    }
  }
  // Merge history
  for (const p of boardBlank) {
    isBlank.add(p);
  }

  const formed = extractFormedWords(board, placed);
  let total = 0;
  for (const fw of formed) {
    total += computeWordScore(board, fw, isNew, isBlank);
  }
  if (placed.length === 7) {
    total += 50;
  }
  return total;
}

function resolveBlanks(rack: string, letters: PlacedLetter[]): { letters: PlacedLetter[], error?: string } {
  const rackCount: Record<string, number> = {};
  for (const r of rack) {
    rackCount[r] = (rackCount[r] ?? 0) + 1;
  }

  const resolved = JSON.parse(JSON.stringify(letters)) as PlacedLetter[];
  for (const l of resolved) {
    if (l.blank) {
      if (!rackCount['?']) {
        return { letters: [], error: "No blank tile in rack" };
      }
      rackCount['?']--;
    }
  }

  for (let i = 0; i < resolved.length; i++) {
    if (resolved[i].blank) continue;
    const char = resolved[i].char;
    if (rackCount[char] && rackCount[char] > 0) {
      rackCount[char]--;
    } else if (rackCount['?'] && rackCount['?'] > 0) {
      resolved[i].blank = true;
      rackCount['?']--;
    } else {
      return { letters: [], error: `Missing letter ${char} in rack` };
    }
  }

  return { letters: resolved };
}

function drawLetters(available: string[], count: number): { drawn: string[], remaining: string[] } {
  const drawn: string[] = [];
  const remaining = [...available];
  const drawCount = Math.min(remaining.length, count);
  for (let i = 0; i < drawCount; i++) {
    const idx = Math.floor(Math.random() * remaining.length);
    drawn.push(remaining[idx]);
    remaining.splice(idx, 1);
  }
  return { drawn, remaining };
}

serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response("ok", { headers: corsHeaders });
  }

  try {
    const supabaseUrl = Deno.env.get("SUPABASE_URL") ?? "";
    const supabaseServiceKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY") ?? "";
    const supabase = createClient(supabaseUrl, supabaseServiceKey);

    // Get Auth user from header
    const authHeader = req.headers.get("Authorization");
    if (!authHeader) {
      return new Response(JSON.stringify({ error: "Missing authorization header" }), { status: 401, headers: corsHeaders });
    }

    const { data: { user }, error: authError } = await supabase.auth.getUser(authHeader.replace("Bearer ", ""));
    if (authError || !user) {
      return new Response(JSON.stringify({ error: "Unauthorized" }), { status: 401, headers: corsHeaders });
    }

    // Get public integer user ID
    const { data: publicUser, error: publicUserError } = await supabase
      .from("users")
      .select("id, username")
      .eq("uuid", user.id)
      .single();

    if (publicUserError || !publicUser) {
      return new Response(JSON.stringify({ error: "User profile not found" }), { status: 404, headers: corsHeaders });
    }

    const url = new URL(req.url);
    const pathParts = url.pathname.split("/");
    
    const body = await req.json();
    const action = body.action || pathParts[pathParts.length - 1]; // "play" or "simulate_score"
    const gameId = body.game_id || pathParts[pathParts.length - 2];
    let playedLetters = body.letters as PlacedLetter[];

    if (!playedLetters || playedLetters.length === 0) {
      return new Response(JSON.stringify({ error: "No letters provided" }), { status: 400, headers: corsHeaders });
    }

    // Load game details
    let { data: game, error: gameError } = await supabase
      .from("games")
      .select("*")
      .eq("id", gameId)
      .maybeSingle();

    let isPuzzle = false;
    if (gameError || !game) {
      // Check if it's a daily puzzle
      const { data: puzzle, error: puzzleError } = await supabase
        .from("daily_puzzles")
        .select("*")
        .eq("id", gameId)
        .maybeSingle();

      if (puzzleError || !puzzle) {
        return new Response(JSON.stringify({ error: "Game or Puzzle not found" }), { status: 404, headers: corsHeaders });
      }

      isPuzzle = true;
      game = {
        id: puzzle.id,
        board: puzzle.board,
        available_letters: puzzle.available_letters,
        current_turn: publicUser.id,
        level: puzzle.level
      };
    }

    // Get player rack
    let playerRack = "";
    let player: any = null;

    if (isPuzzle) {
      playerRack = game.available_letters;
    } else {
      // Check player is in game
      const { data: dbPlayer, error: playerError } = await supabase
        .from("game_players")
        .select("*")
        .eq("game_id", gameId)
        .eq("player_id", publicUser.id)
        .single();

      if (playerError || !dbPlayer) {
        return new Response(JSON.stringify({ error: "Player not in game" }), { status: 403, headers: corsHeaders });
      }
      player = dbPlayer;
      playerRack = player.rack;
    }

    // Deduce/resolve blanks
    const resolvedResult = resolveBlanks(playerRack, playedLetters);
    if (resolvedResult.error) {
      return new Response(JSON.stringify({ error: resolvedResult.error }), { status: 400, headers: corsHeaders });
    }
    playedLetters = resolvedResult.letters;

    // Load historical blank tiles positions for correct scoring
    const boardBlanks = new Set<string>();
    if (!isPuzzle) {
      const { data: historicalMoves, error: movesError } = await supabase
        .from("game_moves")
        .select("move")
        .eq("game_id", gameId);

      if (!movesError && historicalMoves) {
        for (const m of historicalMoves) {
          const mv = m.move as any;
          if (mv && Array.isArray(mv.letters)) {
            for (const pl of mv.letters) {
              if (pl.blank) {
                boardBlanks.add(`${pl.x},${pl.y}`);
              }
            }
          }
        }
      }
    }

    // Load board and apply letters
    const board = game.board as string[][];

    // Check alignment
    const sameRow = playedLetters.every(l => l.y === playedLetters[0].y);
    const sameCol = playedLetters.every(l => l.x === playedLetters[0].x);
    if (!sameRow && !sameCol) {
      return new Response(JSON.stringify({ error: "Letters must be aligned in the same row or column" }), { status: 400, headers: corsHeaders });
    }

    // Apply letters temporarily
    const boardCopy = board.map(row => [...row]);
    for (const l of playedLetters) {
      if (boardCopy[l.y][l.x]) {
        return new Response(JSON.stringify({ error: `Cell at ${l.x},${l.y} already occupied` }), { status: 400, headers: corsHeaders });
      }
      boardCopy[l.y][l.x] = l.char;
    }

    // Verify first move covers center (7,7) or connects to existing tiles
    let isFirstMove = true;
    for (let y = 0; y < 15 && isFirstMove; y++) {
      for (let x = 0; x < 15; x++) {
        if (board[y][x]) {
          isFirstMove = false;
          break;
        }
      }
    }

    if (isFirstMove) {
      const coversCenter = playedLetters.some(l => l.x === 7 && l.y === 7);
      if (!coversCenter) {
        return new Response(JSON.stringify({ error: "First move must cover the center cell" }), { status: 400, headers: corsHeaders });
      }
    } else {
      let touchesExisting = false;
      const neighbors = [
        [-1, 0], [1, 0], [0, -1], [0, 1]
      ];
      for (const l of playedLetters) {
        for (const n of neighbors) {
          const nx = l.x + n[0];
          const ny = l.y + n[1];
          if (nx >= 0 && nx < 15 && ny >= 0 && ny < 15 && board[ny][nx]) {
            touchesExisting = true;
            break;
          }
        }
        if (touchesExisting) break;
      }
      if (!touchesExisting) {
        return new Response(JSON.stringify({ error: "Word must connect to existing letters" }), { status: 400, headers: corsHeaders });
      }
    }

    // Extract formed words
    const formed = extractFormedWords(boardCopy, playedLetters);

    // Verify words in dictionary
    const wordTexts = formed.map(fw => fw.word);
    const { data: dictWords, error: dictError } = await supabase
      .from("dictionary_words")
      .select("word")
      .in("word", wordTexts);

    if (dictError) {
      return new Response(JSON.stringify({ error: "Dictionary lookup failed" }), { status: 500, headers: corsHeaders });
    }

    const validWordSet = new Set(dictWords?.map(w => w.word));
    for (const w of wordTexts) {
      if (!validWordSet.has(w)) {
        return new Response(JSON.stringify({ error: `Invalid word played: ${w}` }), { status: 400, headers: corsHeaders });
      }
    }

    // Compute score
    const moveScore = computeMoveScore(boardCopy, playedLetters, boardBlanks);

    if (action === "simulate_score") {
      return new Response(JSON.stringify({ score: moveScore }), { status: 200, headers: { ...corsHeaders, "Content-Type": "application/json" } });
    }

    if (action === "play" && isPuzzle) {
      const { data: attempt, error: attemptErr } = await supabase
        .from("puzzle_attempts")
        .select("id, started_at, submitted_at")
        .eq("puzzle_id", gameId)
        .eq("player_id", publicUser.id)
        .maybeSingle();

      if (attemptErr || !attempt) {
        return new Response(JSON.stringify({ error: "Vous devez d'abord démarrer le puzzle" }), { status: 400, headers: corsHeaders });
      }

      if (attempt.submitted_at) {
        return new Response(JSON.stringify({ error: "Vous avez déjà soumis ce puzzle" }), { status: 400, headers: corsHeaders });
      }

      const startedAt = new Date(attempt.started_at);
      const now = new Date();
      const timeUsed = Math.floor((now.getTime() - startedAt.getTime()) / 1000);
      const expectedTimeout = game.level === 3 ? 420 : game.level === 2 ? 300 : 180;

      if (expectedTimeout > 0 && timeUsed > expectedTimeout + 10) {
        return new Response(JSON.stringify({ error: "Le temps imparti a été dépassé" }), { status: 400, headers: corsHeaders });
      }

      // Format words_played
      const isNew = new Set(playedLetters.map(l => `${l.x},${l.y}`));
      const isBlank = new Set(playedLetters.filter(l => l.blank).map(l => `${l.x},${l.y}`));
      const wordsPlayed = formed.map(fw => {
        const wordScore = computeWordScore(boardCopy, fw, isNew, isBlank);
        return {
          Word: fw.word,
          Position: `${fw.startX},${fw.startY}`,
          Direction: fw.dy !== 0 ? "vertical" : "horizontal",
          Score: wordScore
        };
      });

      const { error: updateErr } = await supabase
        .from("puzzle_attempts")
        .update({
          score: moveScore,
          words_played: wordsPlayed,
          time_used: timeUsed,
          submitted_at: now.toISOString()
        })
        .eq("id", attempt.id);

      if (updateErr) {
        return new Response(JSON.stringify({ error: `Failed to save attempt: ${updateErr.message}` }), { status: 500, headers: corsHeaders });
      }

      return new Response(JSON.stringify({
        id: attempt.id,
        puzzle_id: gameId,
        player_id: publicUser.id,
        score: moveScore,
        time_used_secs: timeUsed,
        submitted_at: now.toISOString(),
        words_played: wordsPlayed
      }), { status: 200, headers: { ...corsHeaders, "Content-Type": "application/json" } });
    }

    if (action === "play") {
      // 1. Verify turn
      if (game.current_turn !== publicUser.id) {
        return new Response(JSON.stringify({ error: "Not your turn" }), { status: 400, headers: corsHeaders });
      }

      // 2. Compute new rack (remove played letters)
      let rackStr = player.rack;
      for (const l of playedLetters) {
        const toRemove = l.blank ? "?" : l.char;
        const idx = rackStr.indexOf(toRemove);
        if (idx === -1) {
          return new Response(JSON.stringify({ error: `Letter ${toRemove} not in rack` }), { status: 400, headers: corsHeaders });
        }
        rackStr = rackStr.slice(0, idx) + rackStr.slice(idx + 1);
      }

      // 3. Draw new letters
      const bagArray = game.available_letters.split("");
      const drawCount = 7 - rackStr.length;
      const drawResult = drawLetters(bagArray, drawCount);
      const newRackStr = rackStr + drawResult.drawn.join("");
      const newBagStr = drawResult.remaining.join("");

      // 4. Calculate next turn player
      const { data: players, error: playersError } = await supabase
        .from("game_players")
        .select("player_id, position, rack")
        .eq("game_id", gameId)
        .order("position", { ascending: true });

      if (playersError || !players) {
        return new Response(JSON.stringify({ error: "Failed to load players" }), { status: 500, headers: corsHeaders });
      }

      const currentPlayerPos = player.position;
      const nextPlayer = players.find(p => p.position === (currentPlayerPos + 1) % players.length);
      const nextPlayerId = nextPlayer ? nextPlayer.player_id : publicUser.id;

      // 5. Check if game is finished (rack is empty AND bag is empty)
      let isEnded = false;
      let winnerUsername = "";
      const playerUpdates: any[] = [];

      // Add updates for the playing user
      playerUpdates.push({
        player_id: publicUser.id,
        rack: newRackStr,
        score_to_add: moveScore
      });

      if (newRackStr.length === 0 && newBagStr.length === 0) {
        isEnded = true;

        // Deduct remaining letters score from other players
        let totalDeductions = 0;
        for (const p of players) {
          if (p.player_id === publicUser.id) continue;
          let pts = 0;
          for (const c of p.rack) {
            if (c !== "?") {
              pts += LETTER_VALUES[c] ?? 0;
            }
          }
          totalDeductions += pts;
          playerUpdates.push({
            player_id: p.player_id,
            rack: p.rack,
            score_to_add: -pts
          });
        }

        // Add bonus to finisher
        playerUpdates[0].score_to_add += totalDeductions;

        // Determine winner
        // Compute final scores by taking existing scores + score_to_add
        const finalScores = players.map(p => {
          const update = playerUpdates.find(up => up.player_id === p.player_id);
          const scoreToAdd = update ? update.score_to_add : 0;
          // Note: for current user, select score from db + scoreToAdd
          const currentScore = p.player_id === publicUser.id ? player.score : (p as any).score ?? 0; // Wait, we can fetch all scores
          return { player_id: p.player_id, finalScore: currentScore + scoreToAdd };
        });

        // Let's query users table to get winner username
        // (For simplicity we can query the database or do it in rpc)
        // Since we are going to call the database function rpc_commit_move, we can let Deno calculate who has the highest score and fetch their username.
      }

      // To find the winner username and player scores:
      const { data: dbPlayersInfo, error: dbPlayersInfoError } = await supabase
        .from("game_players")
        .select("player_id, score, users(username)")
        .eq("game_id", gameId);

      if (!dbPlayersInfoError && dbPlayersInfo) {
        const finalScores = dbPlayersInfo.map(p => {
          const update = playerUpdates.find(up => up.player_id === p.player_id);
          const scoreToAdd = update ? update.score_to_add : 0;
          return {
            player_id: p.player_id,
            username: (p.users as any)?.username ?? "Joueur",
            finalScore: p.score + scoreToAdd
          };
        });

        finalScores.sort((a, b) => b.finalScore - a.finalScore);
        winnerUsername = finalScores[0]?.username ?? "";
      }

      // 6. Commit everything in a single transaction via our database RPC!
      const moveJSON = {
        word: body.word || wordTexts[0],
        startX: body.x ?? playedLetters[0].x,
        startY: body.y ?? playedLetters[0].y,
        direction: body.dir || (sameRow ? "H" : "V"),
        letters: playedLetters,
        score: moveScore
      };

      const { error: commitError } = await supabase.rpc("rpc_commit_move", {
        p_game_id: gameId,
        p_user_id: publicUser.id,
        p_move_json: moveJSON,
        p_new_board: boardCopy,
        p_new_bag: newBagStr,
        p_next_turn_id: nextPlayerId,
        p_is_ended: isEnded,
        p_winner_username: winnerUsername,
        p_player_updates: playerUpdates
      });

      if (commitError) {
        return new Response(JSON.stringify({ error: `Database commit failed: ${commitError.message}` }), { status: 500, headers: corsHeaders });
      }

      return new Response(JSON.stringify({ success: true, new_rack: newRackStr.split("") }), { status: 200, headers: { ...corsHeaders, "Content-Type": "application/json" } });
    }

    return new Response(JSON.stringify({ error: "Invalid action" }), { status: 400, headers: corsHeaders });

  } catch (err) {
    return new Response(
      JSON.stringify({ error: err.message }),
      { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
    );
  }
});
