<div class="flex flex-col items-center justify-center min-h-[85vh] px-4 py-8">
    <div class="w-full max-w-md glass-card rounded-3xl p-8 text-center border border-white/60 shadow-xl relative overflow-hidden">
        
        <!-- Background accent glows -->
        <div class="absolute -top-12 -left-12 w-24 h-24 rounded-full bg-brand-gold/10 blur-xl"></div>
        <div class="absolute -bottom-12 -right-12 w-24 h-24 rounded-full bg-brand-emerald/10 blur-xl"></div>
        
        <!-- Styled Scrabble Title Tiles -->
        <div class="flex justify-center gap-1.5 mb-6">
            {#each ['S', 'C', 'R', 'A', 'B', 'B', 'L', 'E'] as letter}
                <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-amber-50 to-brand-gold-light text-gray-900 border border-brand-gold/30 shadow flex items-center justify-center font-bold text-lg relative select-none">
                    {letter}
                    <span class="absolute bottom-0.5 right-1 text-[8px] font-normal text-stone-500">
                        {letter === 'E' || letter === 'A' || letter === 'I' || letter === 'L' || letter === 'N' || letter === 'O' || letter === 'R' || letter === 'S' || letter === 'T' || letter === 'U' ? 1 : 
                         letter === 'D' || letter === 'G' || letter === 'M' ? 2 : 
                         letter === 'B' || letter === 'C' || letter === 'P' ? 3 : 
                         letter === 'F' || letter === 'H' || letter === 'V' ? 4 : 8}
                    </span>
                </div>
            {/each}
        </div>

        <h1 class="text-3xl font-extrabold mb-3 text-stone-800 tracking-tight">Scrabble en ligne</h1>
        <p class="text-stone-600 text-sm leading-relaxed mb-8 max-w-sm mx-auto">
            Jouez en ligne avec vos amis ou d'autres passionnés, gratuitement et sans publicité. Une expérience premium entièrement optimisée pour smartphone.
        </p>

        <div class="flex flex-col gap-3">
            <a 
                href="/login" 
                class="w-full inline-flex items-center justify-center bg-brand-emerald hover:bg-brand-emerald-hover text-white py-3.5 px-6 rounded-2xl font-bold shadow-lg shadow-brand-emerald/10 active:scale-[0.98] transition-all"
            >
                Se connecter
            </a>
            <a 
                href="/register" 
                class="w-full inline-flex items-center justify-center bg-white hover:bg-brand-cream-dark text-stone-800 py-3.5 px-6 rounded-2xl font-bold border border-stone-200 shadow-sm active:scale-[0.98] transition-all"
            >
                Créer un compte
            </a>
        </div>

        <div class="mt-8 pt-6 border-t border-stone-100">
            <a 
                href="/leaderboard" 
                class="inline-flex items-center gap-2 text-sm font-semibold text-brand-emerald hover:text-brand-emerald-hover transition-colors"
            >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"/>
                </svg>
                Voir le classement global
            </a>
        </div>
    </div>
</div>