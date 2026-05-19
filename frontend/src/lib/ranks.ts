export type RankKey = 'fer' | 'bronze' | 'argent' | 'or' | 'platine';

export type RankInfo = {
    key: RankKey;
    label: string;
    minRating: number;
    accentClass: string;
    softClass: string;
    icon: string;
};

export const RANKS: RankInfo[] = [
    {
        key: 'fer',
        label: 'Fer',
        minRating: 0,
        accentClass: 'text-stone-700 ring-stone-300',
        softClass: 'bg-stone-100 text-stone-700',
        icon: '/rank_icon/iron.png'
    },
    {
        key: 'bronze',
        label: 'Bronze',
        minRating: 1500,
        accentClass: 'text-orange-700 ring-orange-300',
        softClass: 'bg-orange-100 text-orange-700',
        icon: '/rank_icon/bronze.png'
    },
    {
        key: 'argent',
        label: 'Argent',
        minRating: 1650,
        accentClass: 'text-slate-700 ring-slate-300',
        softClass: 'bg-slate-100 text-slate-700',
        icon: '/rank_icon/silver.png'
    },
    {
        key: 'or',
        label: 'Or',
        minRating: 1800,
        accentClass: 'text-amber-700 ring-amber-300',
        softClass: 'bg-amber-100 text-amber-700',
        icon: '/rank_icon/gold.png'
    },
    {
        key: 'platine',
        label: 'Platine',
        minRating: 1950,
        accentClass: 'text-cyan-700 ring-cyan-300',
        softClass: 'bg-cyan-100 text-cyan-700',
        icon: '/rank_icon/platinium.png'
    }
];

export function getRankInfo(rating: number): RankInfo {
    let current = RANKS[0];
    for (const rank of RANKS) {
        if (rating >= rank.minRating) {
            current = rank;
        }
    }
    return current;
}