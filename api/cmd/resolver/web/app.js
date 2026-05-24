// Special Cells Mapping matching the backend/game
const SPECIAL_CELLS = {
    "0,0": "TW", "0,7": "TW", "0,14": "TW", "7,0": "TW", "7,14": "TW", "14,0": "TW", "14,7": "TW", "14,14": "TW",
    "1,1": "DW", "2,2": "DW", "3,3": "DW", "4,4": "DW", "10,10": "DW", "11,11": "DW", "12,12": "DW", "13,13": "DW",
    "1,13": "DW", "2,12": "DW", "3,11": "DW", "4,10": "DW", "10,4": "DW", "11,3": "DW", "12,2": "DW", "13,1": "DW",
    "7,7": "★",
    "1,5": "TL", "1,9": "TL", "5,1": "TL", "5,5": "TL", "5,9": "TL", "5,13": "TL", "9,1": "TL", "9,5": "TL", "9,9": "TL", "9,13": "TL", "13,5": "TL", "13,9": "TL",
    "0,3": "DL", "0,11": "DL", "2,6": "DL", "2,8": "DL", "3,0": "DL", "3,7": "DL", "3,14": "DL", "6,2": "DL", "6,6": "DL", "6,8": "DL", "6,12": "DL", "7,3": "DL", "7,11": "DL", "8,2": "DL", "8,6": "DL", "8,8": "DL", "8,12": "DL", "11,0": "DL", "11,7": "DL", "11,14": "DL", "12,6": "DL", "12,8": "DL", "14,3": "DL", "14,11": "DL"
};

const LETTER_VALUES = {
    'A': 1, 'B': 3, 'C': 3, 'D': 2, 'E': 1, 'F': 4, 'G': 2, 'H': 4, 'I': 1, 'J': 8, 'K': 10, 'L': 1, 'M': 2,
    'N': 1, 'O': 1, 'P': 3, 'Q': 8, 'R': 1, 'S': 1, 'T': 1, 'U': 1, 'V': 4, 'W': 10, 'X': 10, 'Y': 10, 'Z': 10
};

// Global state
let currentBoard = Array.from({ length: 15 }, () => Array(15).fill(""));
let activeDirection = "H"; // "H" for horizontal typing, "V" for vertical
let lastMoveResult = null;

// DOM Elements
const boardEl = document.getElementById("board");
const rackInput = document.getElementById("rack-input");
const jsonInput = document.getElementById("json-input");
const btnSolve = document.getElementById("btn-solve");
const btnClear = document.getElementById("btn-clear");
const btnExample = document.getElementById("btn-example");
const btnApplyMove = document.getElementById("btn-apply-move");
const resultCard = document.getElementById("result-card");
const noResultCard = document.getElementById("no-result-card");

// Initialize grid
function initBoard() {
    boardEl.innerHTML = "";
    for (let y = 0; y < 15; y++) {
        for (let x = 0; x < 15; x++) {
            const cell = document.createElement("div");
            cell.classList.add("scrabble-cell");
            
            const key = `${y},${x}`;
            const special = SPECIAL_CELLS[key];
            
            if (special) {
                cell.classList.add(`cell-${special.toLowerCase()}`);
                cell.setAttribute("data-special", special === "★" ? "" : special);
            } else {
                cell.classList.add("cell-normal");
            }
            
            cell.setAttribute("tabindex", "0");
            cell.setAttribute("data-x", x);
            cell.setAttribute("data-y", y);
            
            // Events
            cell.addEventListener("click", () => focusCell(x, y));
            cell.addEventListener("keydown", (e) => handleCellKeydown(e, x, y));
            
            boardEl.appendChild(cell);
        }
    }
    updateBoardUI();
}

function updateBoardUI() {
    for (let y = 0; y < 15; y++) {
        for (let x = 0; x < 15; x++) {
            const cell = boardEl.querySelector(`[data-x="${x}"][data-y="${y}"]`);
            const letter = currentBoard[y][x];
            
            // Clear existing tiles
            const existingTile = cell.querySelector(".scrabble-tile");
            if (existingTile) existingTile.remove();
            cell.classList.remove("has-letter");
            
            if (letter && letter.trim() !== "") {
                cell.classList.add("has-letter");
                const tile = document.createElement("div");
                tile.classList.add("scrabble-tile");
                tile.textContent = letter.toUpperCase();
                
                const score = LETTER_VALUES[letter.toUpperCase()];
                if (score !== undefined) {
                    const scoreEl = document.createElement("span");
                    scoreEl.classList.add("letter-score");
                    scoreEl.textContent = score;
                    tile.appendChild(scoreEl);
                }
                cell.appendChild(tile);
            }
        }
    }
    
    // Sync to JSON textarea
    jsonInput.value = JSON.stringify({ board: currentBoard }, null, 4);
}

function focusCell(x, y) {
    const cell = boardEl.querySelector(`[data-x="${x}"][data-y="${y}"]`);
    if (cell) {
        // Remove active class from previous
        boardEl.querySelectorAll(".active-input").forEach(c => c.classList.remove("active-input"));
        cell.classList.add("active-input");
        cell.focus();
    }
}

function handleCellKeydown(e, x, y) {
    const key = e.key;
    
    // Allow navigation
    if (key === "ArrowRight") {
        focusCell(Math.min(14, x + 1), y);
        activeDirection = "H";
        e.preventDefault();
        return;
    }
    if (key === "ArrowLeft") {
        focusCell(Math.max(0, x - 1), y);
        activeDirection = "H";
        e.preventDefault();
        return;
    }
    if (key === "ArrowDown") {
        focusCell(x, Math.min(14, y + 1));
        activeDirection = "V";
        e.preventDefault();
        return;
    }
    if (key === "ArrowUp") {
        focusCell(x, Math.max(0, y - 1));
        activeDirection = "V";
        e.preventDefault();
        return;
    }
    
    // Deleting character
    if (key === "Backspace" || key === "Delete") {
        currentBoard[y][x] = "";
        updateBoardUI();
        focusCell(x, y);
        e.preventDefault();
        return;
    }
    
    // Typing single character (A-Z)
    if (/^[a-zA-Z]$/.test(key)) {
        currentBoard[y][x] = key.toUpperCase();
        updateBoardUI();
        
        // Move to next cell
        if (activeDirection === "H") {
            focusCell(Math.min(14, x + 1), y);
        } else {
            focusCell(x, Math.min(14, y + 1));
        }
        e.preventDefault();
    }
}

// Clear board
btnClear.addEventListener("click", () => {
    currentBoard = Array.from({ length: 15 }, () => Array(15).fill(""));
    updateBoardUI();
    clearBestMoveOverlay();
    resultCard.classList.add("hidden");
    noResultCard.classList.add("hidden");
});

// Load Example
btnExample.addEventListener("click", () => {
    // Exact board structure provided in prompt
    currentBoard = [
        ["", "", "", "", "", "", "", "S", "", "", "", "", "", "", ""],
        ["", "", "", "V", "I", "T", "R", "A", "", "", "", "", "", "", ""],
        ["", "", "", "", "", "", "", "C", "", "", "", "", "", "", ""],
        ["", "", "", "", "", "", "", "H", "", "E", "W", "E", "", "", ""],
        ["", "", "", "", "P", "I", "M", "E", "N", "T", "E", "", "", "", ""],
        ["", "", "", "", "", "", "", "R", "", "", "B", "", "", "", ""],
        ["", "", "", "", "", "", "", "I", "", "", "S", "", "", "", ""],
        ["", "", "", "Y", "O", "D", "L", "E", "R", "", "", "", "", "", ""],
        ["", "", "", "", "", "E", "", "", "", "", "", "", "", "", ""],
        ["V", "", "", "", "", "C", "", "", "", "", "", "", "", "", ""],
        ["I", "", "", "", "", "A", "", "", "", "", "", "", "", "", ""],
        ["D", "A", "U", "P", "H", "I", "N", "", "", "", "", "", "", "", ""],
        ["I", "", "", "", "", "S", "", "", "", "", "", "", "", "", ""],
        ["M", "", "", "", "", "S", "", "", "", "", "", "", "", "", ""],
        ["A", "", "", "", "B", "A", "N", "N", "E", "R", "", "", "", "", ""]
    ];
    rackInput.value = "AEIMNPT";
    updateBoardUI();
    clearBestMoveOverlay();
    resultCard.classList.add("hidden");
    noResultCard.classList.add("hidden");
});

// Parse Textarea JSON
jsonInput.addEventListener("input", () => {
    try {
        const parsed = JSON.parse(jsonInput.value);
        if (parsed && Array.isArray(parsed.board)) {
            // Validate it is 15x15
            if (parsed.board.length === 15 && parsed.board.every(row => Array.isArray(row) && row.length === 15)) {
                currentBoard = parsed.board.map(row => row.map(cell => typeof cell === "string" ? cell : ""));
                updateBoardUI();
                clearBestMoveOverlay();
            }
        }
    } catch (e) {
        // Ignore parsing errors while typing
    }
});

// Clear glowing overlay of result
function clearBestMoveOverlay() {
    boardEl.querySelectorAll(".scrabble-tile.new-placed").forEach(t => {
        t.remove();
    });
    boardEl.querySelectorAll(".scrabble-cell.has-letter").forEach(c => {
        const tile = c.querySelector(".scrabble-tile");
        if (tile) {
            tile.classList.remove("new-placed");
        }
    });
}

// Draw best move on board
function drawBestMoveOverlay(move) {
    clearBestMoveOverlay();
    if (!move || !Array.isArray(move.letters)) return;
    
    move.letters.forEach(letter => {
        const cell = boardEl.querySelector(`[data-x="${letter.x}"][data-y="${letter.y}"]`);
        if (cell) {
            // If already has a letter tile, just add class
            const existingTile = cell.querySelector(".scrabble-tile");
            if (existingTile) {
                existingTile.classList.add("new-placed");
            } else {
                // Else create a virtual preview tile
                const tile = document.createElement("div");
                tile.classList.add("scrabble-tile", "new-placed");
                tile.textContent = letter.char.toUpperCase();
                
                const score = LETTER_VALUES[letter.char.toUpperCase()];
                if (score !== undefined) {
                    const scoreEl = document.createElement("span");
                    scoreEl.classList.add("letter-score");
                    scoreEl.textContent = score;
                    tile.appendChild(scoreEl);
                }
                cell.appendChild(tile);
            }
        }
    });
}

// Solve click
btnSolve.addEventListener("click", async () => {
    const rack = rackInput.value.toUpperCase().replace(/[^A-Z?]/g, "");
    rackInput.value = rack;
    
    if (!rack) {
        alert("Veuillez saisir au moins une lettre sur votre rack.");
        return;
    }
    
    // Show loading
    btnSolve.disabled = true;
    btnSolve.querySelector(".btn-text").classList.add("hidden");
    btnSolve.querySelector(".loader").classList.remove("hidden");
    
    resultCard.classList.add("hidden");
    noResultCard.classList.add("hidden");
    clearBestMoveOverlay();
    
    try {
        const response = await fetch("/solve", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                board: currentBoard,
                rack: rack
            })
        });
        
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || "Erreur serveur lors de la résolution.");
        }
        
        const data = await response.json();
        if (data && data.word) {
            lastMoveResult = data;
            
            // Populate results
            document.getElementById("res-score").textContent = data.score;
            document.getElementById("res-word").textContent = data.word;
            document.getElementById("res-pos").textContent = `${data.startX}, ${data.startY}`;
            document.getElementById("res-dir").textContent = data.direction === "H" ? "Horizontal" : "Vertical";
            
            // List placed letters
            const listEl = document.getElementById("res-placed-letters");
            listEl.innerHTML = "";
            data.letters.forEach(l => {
                const badge = document.createElement("span");
                badge.classList.add("placed-letter-badge");
                badge.innerHTML = `<strong>${l.char.toUpperCase()}</strong>${l.blank ? " (Joker)" : ""} <span class="coord">(${l.x},${l.y})</span>`;
                listEl.appendChild(badge);
            });
            
            resultCard.classList.remove("hidden");
            drawBestMoveOverlay(data);
            
            // Scroll to result on small screens
            if (window.innerWidth <= 1024) {
                resultCard.scrollIntoView({ behavior: "smooth" });
            }
        } else {
            noResultCard.classList.remove("hidden");
        }
    } catch (e) {
        console.error(e);
        alert(e.message || "Erreur réseau.");
    } finally {
        btnSolve.disabled = false;
        btnSolve.querySelector(".btn-text").classList.remove("hidden");
        btnSolve.querySelector(".loader").classList.add("hidden");
    }
});

// Apply move click
btnApplyMove.addEventListener("click", () => {
    if (!lastMoveResult || !Array.isArray(lastMoveResult.letters)) return;
    
    // Place them permanently in currentBoard
    lastMoveResult.letters.forEach(letter => {
        currentBoard[letter.y][letter.x] = letter.char.toUpperCase();
    });
    
    // Clear rack letters
    let currentRack = rackInput.value.toUpperCase();
    lastMoveResult.letters.forEach(letter => {
        const charToRemove = letter.blank ? "?" : letter.char.toUpperCase();
        const index = currentRack.indexOf(charToRemove);
        if (index !== -1) {
            currentRack = currentRack.substring(0, index) + currentRack.substring(index + 1);
        }
    });
    rackInput.value = currentRack;
    
    updateBoardUI();
    clearBestMoveOverlay();
    
    lastMoveResult = null;
    resultCard.classList.add("hidden");
    
    alert("Le coup a été appliqué avec succès sur le plateau !");
});

// Run
initBoard();
focusCell(7, 7);
activeDirection = "H";
