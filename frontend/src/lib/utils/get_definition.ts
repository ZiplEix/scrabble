export type WiktionaryDefinition = {
    title: string;
    extract: string; // texte brut
    url: string; // lien vers la page
} | null;

const url = "https://www.larousse.fr/dictionnaires/francais/"
export async function getDefinition(
    word: string
): Promise<WiktionaryDefinition> {
    const res = await fetch(`${url}${encodeURIComponent(word)}`);
    if (!res.ok) return null;
    if (res.status === 404) return null;

    const HTML = await res.text();

    return {
        title: word,
        extract: HTML,
        url: res.url,
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
    return extractFromWiktionary(doc, langId);
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

            const text = cleanText(liClone.textContent || "");

            const cleanedText = stripLeadingNumberDot(text).trim();

            if (cleanedText) defs.push(cleanedText);
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

/* ===================== Helpers ===================== */

/** Supprime un indice en tête du style "1.", "12.", "1 ." ... */
function stripLeadingNumberDot(text: string): string {
  return text.replace(/^\s*\d+\s*\.\s*/u, "");
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
