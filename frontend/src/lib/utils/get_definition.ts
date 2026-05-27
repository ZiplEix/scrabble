import { api } from "$lib/api";

export type WiktionaryDefinition = {
    title: string;
    extract: string; // texte brut
    url: string; // lien vers la page
    is_parsed?: boolean;
    def?: DefinitionGroup[];
} | null;

export async function getDefinition(
    word: string
): Promise<WiktionaryDefinition> {
    const cleanWord = word.toUpperCase().trim();
    if (!cleanWord) return null;

    // 1) Tentative d'interrogation du cache de l'API Go
    try {
        const cacheRes = await api.get(`/dictionary/${encodeURIComponent(cleanWord)}`);
        if (cacheRes.status === 200 && cacheRes.data) {
            // Cache Hit !
            return {
                title: cleanWord,
                extract: "",
                url: cacheRes.data.url || "",
                is_parsed: true,
                def: cacheRes.data.def || []
            };
        }
    } catch (err: any) {
        // En cas d'erreur 404, on continue vers le scraping.
        if (err?.response?.status !== 404) {
            console.warn(`Go API dictionary cache lookup failed for ${cleanWord}`, err);
        }
    }

    // 2) Cache Miss : Scraping & parsing client-side

    // Étape A : Wiktionnaire (Direct client-side CORS)
    try {
        const wiktUrl = `https://fr.wiktionary.org/w/api.php?action=query&format=json&prop=extracts&titles=${encodeURIComponent(cleanWord)}&origin=*`;
        const res = await fetch(wiktUrl);
        if (res.ok) {
            const data = await res.json();
            const pages = data?.query?.pages;
            if (pages) {
                const pageId = Object.keys(pages)[0];
                if (pageId !== "-1" && pages[pageId] && pages[pageId].extract) {
                    const extract = pages[pageId].extract;
                    const defGroups = extractDefinitions(extract);
                    if (defGroups && defGroups.length > 0) {
                        const url = `https://fr.wiktionary.org/wiki/${encodeURIComponent(cleanWord)}`;
                        
                        // Sauvegarde asynchrone dans le cache DB
                        try {
                            await api.post('/dictionary', {
                                word: cleanWord,
                                definitions: { url, def: defGroups }
                            });
                        } catch (saveErr) {
                            console.error(`Failed to save definition cache for ${cleanWord} from Wiktionnaire`, saveErr);
                        }

                        return {
                            title: cleanWord,
                            extract: "",
                            url,
                            is_parsed: true,
                            def: defGroups
                        };
                    }
                }
            }
        }
    } catch (wiktErr) {
        console.warn(`Direct Wiktionary fetch failed for ${cleanWord}, falling back to Larousse proxy`, wiktErr);
    }

    // Étape B : Fallback Larousse (via proxy car pas de CORS)
    try {
        const res = await fetch(`/api/larousse?word=${encodeURIComponent(cleanWord)}`);
        if (res.ok) {
            const data = await res.json();
            if (data && data.html) {
                const defGroups = extractDefinitions(data.html);
                const url = data.url || `https://www.larousse.fr/dictionnaires/francais/${encodeURIComponent(cleanWord.toLowerCase())}`;
                
                // Sauvegarde asynchrone dans le cache DB
                try {
                    await api.post('/dictionary', {
                        word: cleanWord,
                        definitions: { url, def: defGroups }
                    });
                } catch (saveErr) {
                    console.error(`Failed to save definition cache for ${cleanWord} from Larousse`, saveErr);
                }

                return {
                    title: cleanWord,
                    extract: "",
                    url,
                    is_parsed: true,
                    def: defGroups
                };
            }
        }
    } catch (larousseErr) {
        console.error(`Larousse proxy fallback failed for ${cleanWord}`, larousseErr);
    }

    // Étape C : Fallback 1mot.net (via proxy car pas de CORS)
    try {
        const res = await fetch(`/api/1mot?word=${encodeURIComponent(cleanWord)}`);
        if (res.ok) {
            const data = await res.json();
            if (data && data.html) {
                const defGroups = extractDefinitions(data.html);
                if (defGroups && defGroups.length > 0) {
                    const url = data.url || `https://1mot.net/${encodeURIComponent(cleanWord.toLowerCase())}`;
                    
                    // Sauvegarde asynchrone dans le cache DB
                    try {
                        await api.post('/dictionary', {
                            word: cleanWord,
                            definitions: { url, def: defGroups }
                        });
                    } catch (saveErr) {
                        console.error(`Failed to save definition cache for ${cleanWord} from 1mot.net`, saveErr);
                    }

                    return {
                        title: cleanWord,
                        extract: "",
                        url,
                        is_parsed: true,
                        def: defGroups
                    };
                }
            }
        }
    } catch (oneMotErr) {
        console.error(`1mot.net proxy fallback failed for ${cleanWord}`, oneMotErr);
    }

    // Étape D : Aucun dictionnaire ne connaît le mot (Mot invalide).
    // On met en cache un tableau vide pour éviter de répéter ces requêtes lentes à l'avenir !
    try {
        await api.post('/dictionary', {
            word: cleanWord,
            definitions: { url: "", def: [] }
        });
    } catch (saveErr) {
        console.error(`Failed to save empty/invalid definition cache for ${cleanWord}`, saveErr);
    }

    return {
        title: cleanWord,
        extract: "",
        url: "",
        is_parsed: true,
        def: []
    };
}

export type DefinitionGroup = {
    type: string;
    definitions: string[];
};

export function extractDefinitions(html: string, langId: string = "fr"): DefinitionGroup[] {
    const doc = parseHTML(html);

    // 1) Tente Larousse
    const fromLarousse = extractFromLarousse(doc);
    if (fromLarousse.length) return fromLarousse;

    // 2) Fallback : Wiktionnaire (ancienne API)
    const fromWikt = extractFromWiktionary(doc, langId);
    if (fromWikt.length) return fromWikt;

    // 3) Fallback : 1mot.net
    return extractFrom1Mot(doc);
}

/* ===================== LAROUSSE ===================== */

function extractFromLarousse(doc: Document): DefinitionGroup[] {
    const out: DefinitionGroup[] = [];
    // Plusieurs entrées possibles — on boucle
    const articles = Array.from(doc.querySelectorAll('#definition article.BlocDefinition'));
    if (!articles.length) return out;

    for (const art of articles) {
        // Type / catégorie grammaticale (ex. "nom masculin", "verbe", "adjectif")
        const cat = art.querySelector('.CatgramDefinition');
        const type = cleanText(cat?.textContent || "");
        if (!type) continue;

        // Définitions
        const defsUL = art.querySelector('ul.Definitions');
        if (!defsUL) continue;

        const defs: string[] = [];
        defsUL.querySelectorAll(':scope > li').forEach((li) => {
            const liClone = li.cloneNode(true) as HTMLElement;

            // Supprimer tout ce qui ressemble à des exemples/citations/illustrations
            liClone.querySelectorAll(
                '.Exemple, .Exemples, .ExempleDefinition, .exemple, .citation, .source, figure, blockquote, q, cite'
            ).forEach(n => n.remove());

            // Certains items contiennent des sous-listes (souvent des exemples) : on les retire par prudence
            liClone.querySelectorAll(':scope ul, :scope ol').forEach(n => n.remove());

            const raw = cleanText(liClone.textContent || "");

            const t1 = removeColonAfterLeadingTags(raw);
            const t2 = stripLeadingIndex(t1);
            const final = replaceTrailingColonWithDot(t2);

            if (final) defs.push(final);
        });

        if (defs.length) {
            out.push({ type, definitions: defs });
        }
    }
    return out;
}

/* ===================== WIKTIONNAIRE (fallback) ===================== */

function extractFromWiktionary(doc: Document, langId: string): DefinitionGroup[] {
    const frH2 = findLanguageH2(doc, langId);
    if (!frH2) return [];
    const groups: DefinitionGroup[] = [];
    const POS_RE = /^(Nom commun|Nom propre|Verbe|Adjectif|Adverbe|Pronom|Conjonction|Interjection|Préposition|Préfixe|Suffixe|Article|Déterminant|Participe|Locution[\s\-][\p{L}\- ]*)/iu;

    for (let el = frH2.nextElementSibling; el && el.tagName !== "H2"; el = el.nextElementSibling) {
        if (el.tagName !== "H3") continue;

        const rawHeading = (el.textContent ?? "").trim().replace(/\s*\d+\s*$/u, "");
        const posMatch = rawHeading.match(POS_RE);
        if (!posMatch) continue;

        const pos = posMatch[0];
        const p = nextElementOfTag(el, "P");
        const gender = extractGenderFromP(p);

        let cur = p ? p.nextElementSibling : el.nextElementSibling;
        let ol: HTMLOListElement | null = null;
        for (let i = 0; i < 5 && cur && cur.tagName !== "H3" && cur.tagName !== "H2"; i++, cur = cur.nextElementSibling) {
            if (cur.tagName === "OL") { ol = cur as HTMLOListElement; break; }
        }
        if (!ol) continue;

        const defs: string[] = [];
        ol.querySelectorAll(':scope > li').forEach((li) => {
            const liClone = li.cloneNode(true) as HTMLElement;
            liClone.querySelectorAll('ul, ol ul, q, blockquote, cite').forEach(n => n.remove());
            const text = cleanText(liClone.textContent || "");
            if (text) defs.push(text);
        });

        if (defs.length) groups.push({ type: gender ? `${pos} ${gender}` : pos, definitions: defs });
    }
    return groups;
}

/* ===================== 1MOT.NET (fallback) ===================== */

function extractFrom1Mot(doc: Document): DefinitionGroup[] {
    const excludes = [
        "mots valides", "mots invalides", "sous-mots", "cousins", 
        "lipogrammes", "épenthèse", "anagrammes", "milieu", 
        "préfixe", "pointage", "langues", "catégories", "sites web"
    ];
    const out: DefinitionGroup[] = [];
    const headings = Array.from(doc.querySelectorAll('h4'));
    for (const h4 of headings) {
        const text = cleanText(h4.textContent || "");
        const lower = text.toLowerCase();
        if (excludes.some(ex => lower.includes(ex))) {
            continue;
        }
        
        // C'est une section de définition
        const defs: string[] = [];
        let cur = h4.nextElementSibling;
        while (cur && cur.tagName !== "H4") {
            if (cur.tagName === "UL") {
                cur.querySelectorAll(':scope > li').forEach(li => {
                    const liText = cleanText(li.textContent || "");
                    if (liText) {
                        defs.push(liText);
                    }
                });
            }
            cur = cur.nextElementSibling;
        }
        if (defs.length) {
            out.push({ type: text, definitions: defs });
        }
    }
    return out;
}

/* ===================== Helpers ===================== */

/** Supprime un indice en tête du style "1.", "12.", "1 ." ... */
export function stripLeadingIndex(text: string): string {
  return text.replace(/^\s*(?:\d+|[ivxlcdm]+)\s*[.)]\s*/iu, "");
}

/** Remplace un ":" final (éventuels espaces après) par "." */
export function replaceTrailingColonWithDot(text: string): string {
  return text.replace(/:\s*$/u, ".");
}

/** Supprime le ":" placé juste après les étiquettes initiales entre parenthèses. */
export function removeColonAfterLeadingTags(text: string): string {
  return text.replace(/^((?:\([^)]*\)\s*)+):\s*/u, "$1");
}

function parseHTML(html: string): Document {
    if (typeof window !== "undefined" && typeof (window as any).DOMParser !== "undefined") {
        return new DOMParser().parseFromString(html, "text/html");
    }
    try {
        const { JSDOM } = require("jsdom");
        return new JSDOM(html).window.document;
    } catch {
        throw new Error("Aucun parser DOM disponible (DOMParser ou jsdom requis).");
    }
}

function findLanguageH2(doc: Document, langId: string): HTMLHeadingElement | null {
    const esc = (globalThis as any).CSS?.escape ?? ((s: string) => s.replace(/([^\w-])/g, '\\$1'));
    for (const h2 of Array.from(doc.querySelectorAll('h2'))) {
        if (h2.querySelector(`span#${esc(langId)}`)) return h2 as HTMLHeadingElement;
    }
    return null;
}

function nextElementOfTag(start: Element, tagName: string): Element | null {
    const upper = tagName.toUpperCase();
    for (let el = start.nextElementSibling; el; el = el.nextElementSibling) {
        if (el.tagName === upper) return el;
        if (el.tagName === "H2" || el.tagName === "H3") break;
    }
    return null;
}

function extractGenderFromP(p: Element | null): string | null {
    if (!p) return null;
    const candidates = Array.from(p.querySelectorAll(':scope i')).map(i => cleanText(i.textContent || ""));
    const known = candidates.find(t => /\b(masculin|féminin|invariable|épicène|pluriel|singulier|masculin et féminin|m\.\s*et\s*f\.)\b/i.test(t));
    return known || null;
}

function cleanText(s: string): string {
    return s
        .replace(/\s+/g, ' ')
        .replace(/\s*[:;,.]\s*/g, m => m.trim() + ' ')
        .replace(/\s+$/g, '')
        .trim();
}
