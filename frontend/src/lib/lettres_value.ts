import { specialCells } from "./cells";

export const letterValues: Record<string, number> = {
	A: 1, B: 3, C: 3, D: 2, E: 1,
	F: 4, G: 2, H: 4, I: 1, J: 8,
	K: 10, L: 1, M: 2, N: 1, O: 1,
	P: 3, Q: 8, R: 1, S: 1, T: 1,
	U: 1, V: 4, W: 10, X: 10, Y: 10, Z: 10,
	'?': 0,
};

export function computeWordValue(letters: { x: number; y: number; letter: string }[]): number {
    let score = 0;
    let wordMultiplier = 1;

    for (const { x, y, letter } of letters) {
        const key = `${x},${y}`;
        const type = specialCells.get(key);
        const val = letterValues[letter.toUpperCase()] ?? 0;

        if (type === 'DL') {
			score += val * 2;
		} else if (type === 'TL') {
			score += val * 3;
		} else if (type === 'DW' || type === '★') {
			score += val;
			wordMultiplier *= 2;
		} else if (type === 'TW') {
			score += val;
			wordMultiplier *= 3;
		} else {
			score += val;
		}
    }

    if (letters.length === 7) {
        score += 50; // Bonus for using all 7 letters
    }

    return score * wordMultiplier;
}