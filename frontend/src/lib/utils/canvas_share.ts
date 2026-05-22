import type { GameInfo } from "../types/game_infos";
import { specialCells } from "../cells";
import { letterValues } from "../lettres_value";

export function downloadGameResultImage(game: GameInfo) {
    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Configurer la taille du Canvas (800 x 1050 pour un aspect poster premium)
    const width = 800;
    const height = 1050;
    canvas.width = width;
    canvas.height = height;

    // --- 1. FOND DE LA CARTE DE PARTAGE ---
    // Un dégradé sombre et premium (Style bleu nuit / ardoise profonde)
    const cardGrad = ctx.createLinearGradient(0, 0, 0, height);
    cardGrad.addColorStop(0, "#0f172a"); // slate-900
    cardGrad.addColorStop(0.5, "#1e1b4b"); // indigo-950
    cardGrad.addColorStop(1, "#020617"); // slate-950
    ctx.fillStyle = cardGrad;
    ctx.fillRect(0, 0, width, height);

    // --- 2. HEADER DE LA CARTE ---
    // Titre de l'application
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";

    // Titre principal
    ctx.fillStyle = "#fbbf24"; // Ambre doré
    ctx.font = "bold 32px Georgia, 'Times New Roman', serif";
    ctx.fillText("SCRABBLE CLUB", width / 2, 45);

    // Sous-titre décoratif
    ctx.fillStyle = "#94a3b8"; // slate-400
    ctx.font = "italic 16px 'Inter', sans-serif";
    ctx.fillText(game.name || "Partie de Scrabble", width / 2, 80);

    // --- 3. DESSIN DU PLATEAU (GRID) ---
    const boardSize = 540;
    const boardX = (width - boardSize) / 2;
    const boardY = 120;
    const cellSize = boardSize / 15;

    // Dessiner le fond du plateau
    ctx.fillStyle = "#1e293b"; // Fond ardoise sombre pour faire ressortir les cases
    ctx.beginPath();
    ctx.roundRect(boardX - 6, boardY - 6, boardSize + 12, boardSize + 12, 12);
    ctx.fill();

    // Bordure fine dorée autour du plateau
    ctx.strokeStyle = "rgba(251, 191, 36, 0.4)";
    ctx.lineWidth = 2;
    ctx.stroke();

    for (let row = 0; row < 15; row++) {
        for (let col = 0; col < 15; col++) {
            const cellX = boardX + col * cellSize;
            const cellY = boardY + row * cellSize;
            const letter = game.board[row]?.[col];

            if (letter && letter !== "") {
                // --- DESSINER UNE TUILE DE BOIS ---
                // Effet de relief avec ombre portée
                ctx.shadowColor = "rgba(0, 0, 0, 0.4)";
                ctx.shadowBlur = 6;
                ctx.shadowOffsetX = 2;
                ctx.shadowOffsetY = 2;

                // Fond de la tuile en dégradé de bois poli
                const tileGrad = ctx.createRadialGradient(
                    cellX + cellSize / 2, cellY + cellSize / 2, 2,
                    cellX + cellSize / 2, cellY + cellSize / 2, cellSize / 2 + 5
                );
                tileGrad.addColorStop(0, "#fde047"); // jaune bois clair
                tileGrad.addColorStop(0.6, "#eab308"); // ambre moyen
                tileGrad.addColorStop(1, "#ca8a04"); // bois sombre doré
                ctx.fillStyle = tileGrad;

                ctx.beginPath();
                ctx.roundRect(cellX + 1.5, cellY + 1.5, cellSize - 3, cellSize - 3, 5);
                ctx.fill();

                // Annuler l'ombre portée pour les dessins suivants
                ctx.shadowBlur = 0;
                ctx.shadowOffsetX = 0;
                ctx.shadowOffsetY = 0;

                // Bordure de la tuile
                ctx.strokeStyle = "#854d0e";
                ctx.lineWidth = 1;
                ctx.stroke();

                // Dessiner la Lettre principale
                ctx.fillStyle = "#1e1b4b"; // Encre indigo foncée
                ctx.font = "bold 20px Georgia, 'Times New Roman', serif";
                ctx.textAlign = "center";
                ctx.textBaseline = "middle";
                ctx.fillText(letter.toUpperCase(), cellX + cellSize / 2, cellY + cellSize / 2 - 2);

                // Dessiner la valeur des points (en bas à droite de la tuile)
                const pointVal = letterValues[letter.toUpperCase()] ?? 0;
                if (pointVal > 0) {
                    ctx.fillStyle = "#451a03";
                    ctx.font = "bold 9px 'Inter', sans-serif";
                    ctx.textAlign = "right";
                    ctx.textBaseline = "bottom";
                    ctx.fillText(pointVal.toString(), cellX + cellSize - 4, cellY + cellSize - 2);
                }
            } else {
                // --- DESSINER CASE SPECIALE OU STANDARD ---
                const key = `${col},${row}`;
                const special = specialCells.get(key);

                let cellBg = "#334155"; // Case vide standard (slate-700)
                let cellLabel = "";
                let textColor = "#f8fafc";

                if (special === "TW") {
                    cellBg = "#ef4444"; // Rouge vif
                    cellLabel = "MT"; // Mot Triple
                } else if (special === "DW") {
                    cellBg = "#f472b6"; // Rose / Saumon
                    cellLabel = "MD"; // Mot Double
                } else if (special === "TL") {
                    cellBg = "#3b82f6"; // Bleu
                    cellLabel = "LT"; // Lettre Triple
                } else if (special === "DL") {
                    cellBg = "#60a5fa"; // Bleu ciel
                    cellLabel = "LD"; // Lettre Double
                } else if (special === "★") {
                    cellBg = "#f59e0b"; // Ambre
                    cellLabel = "★";
                }

                ctx.fillStyle = cellBg;
                ctx.beginPath();
                ctx.roundRect(cellX + 1, cellY + 1, cellSize - 2, cellSize - 2, 3);
                ctx.fill();

                if (cellLabel !== "") {
                    ctx.fillStyle = textColor;
                    ctx.font = "bold 10px 'Inter', sans-serif";
                    ctx.textAlign = "center";
                    ctx.textBaseline = "middle";
                    ctx.fillText(cellLabel, cellX + cellSize / 2, cellY + cellSize / 2);
                }
            }
        }
    }

    // --- 4. SECTION DES SCORES ET DU CLASSEMENT ---
    const scoreY = boardY + boardSize + 40;

    // Trier les joueurs par score décroissant pour afficher le classement
    const sortedPlayers = [...game.players].sort((a, b) => b.score - a.score);

    // Dessiner un séparateur élégant
    ctx.strokeStyle = "rgba(148, 163, 184, 0.2)";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(80, scoreY - 15);
    ctx.lineTo(width - 80, scoreY - 15);
    ctx.stroke();

    // Affichage des cartes des scores côte à côte
    const cardWidth = 280;
    const cardHeight = 110;
    const spacing = 40;
    const startX = (width - (sortedPlayers.length * cardWidth + (sortedPlayers.length - 1) * spacing)) / 2;

    sortedPlayers.forEach((player, idx) => {
        const pX = startX + idx * (cardWidth + spacing);
        const pY = scoreY + 10;
        const isWinner = idx === 0;

        // Arrière-plan de la carte joueur
        const playerCardGrad = ctx.createLinearGradient(pX, pY, pX, pY + cardHeight);
        if (isWinner) {
            playerCardGrad.addColorStop(0, "rgba(251, 191, 36, 0.15)"); // Ambre léger
            playerCardGrad.addColorStop(1, "rgba(251, 191, 36, 0.05)");
        } else {
            playerCardGrad.addColorStop(0, "rgba(30, 41, 59, 0.6)"); // Slate léger
            playerCardGrad.addColorStop(1, "rgba(15, 23, 42, 0.8)");
        }

        ctx.fillStyle = playerCardGrad;
        ctx.beginPath();
        ctx.roundRect(pX, pY, cardWidth, cardHeight, 12);
        ctx.fill();

        // Bordure dorée pour le gagnant
        ctx.strokeStyle = isWinner ? "rgba(251, 191, 36, 0.8)" : "rgba(148, 163, 184, 0.2)";
        ctx.lineWidth = isWinner ? 2 : 1;
        ctx.stroke();

        // Couronne ou Médaille
        ctx.textAlign = "left";
        ctx.textBaseline = "middle";
        if (isWinner) {
            ctx.fillStyle = "#fbbf24";
            ctx.font = "bold 20px 'Inter', sans-serif";
            ctx.fillText("👑  GAGNANT", pX + 20, pY + 25);
        } else {
            ctx.fillStyle = "#94a3b8";
            ctx.font = "bold 16px 'Inter', sans-serif";
            ctx.fillText("🥈  PARTICIPANT", pX + 20, pY + 25);
        }

        // Nom du joueur
        ctx.fillStyle = "#f8fafc";
        ctx.font = "bold 18px 'Inter', sans-serif";
        ctx.fillText(player.username, pX + 20, pY + 60);

        // Score
        ctx.textAlign = "right";
        ctx.fillStyle = isWinner ? "#fbbf24" : "#f1f5f9";
        ctx.font = "bold 42px Georgia, 'Times New Roman', serif";
        ctx.fillText(player.score.toString(), pX + cardWidth - 20, pY + cardHeight / 2 + 5);

        // Sous-titre "pts" sous le score
        ctx.fillStyle = isWinner ? "rgba(251, 191, 36, 0.7)" : "#94a3b8";
        ctx.font = "12px 'Inter', sans-serif";
        ctx.fillText("points", pX + cardWidth - 20, pY + cardHeight - 15);
    });

    // --- 5. BAS DE PAGE / FOOTER DE LA MARQUE ---
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillStyle = "rgba(148, 163, 184, 0.4)";
    ctx.font = "12px 'Inter', sans-serif";
    ctx.fillText("Créé avec amour sur scrabble.baptiste.zip • Scrabble Club", width / 2, height - 40);

    // --- 6. EXPORT / TELECHARGEMENT ---
    const dataURL = canvas.toDataURL("image/png");
    const link = document.createElement("a");
    link.download = `scrabble_partie_${game.name.toLowerCase().replace(/[^a-z0-9]+/g, "_") || game.id}.png`;
    link.href = dataURL;
    link.click();
}
