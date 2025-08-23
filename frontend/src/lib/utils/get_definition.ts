export type WiktionaryDefinition = {
    title: string;
    extract: string; // texte brut
    url: string; // lien vers la page
} | null;

// MediaWiki nécessite le paramètre origin=* pour autoriser le CORS depuis le navigateur.
// On demande aussi le texte brut (explaintext) et on suit les redirections.
const wiktionaryBase = "https://fr.wiktionary.org/w/api.php";
const localProxy = "/api/wiktionary";

export async function getDefinition(
    word: string
): Promise<WiktionaryDefinition> {
    const title = String(word).trim().toLowerCase();
    if (!title) return null;

    const params = new URLSearchParams({
        action: "query",
        format: "json",
        prop: "extracts",
        // explaintext: '1',
        // redirects: '1',
        titles: title,
        origin: "*",
    });

    const directUrl = `${wiktionaryBase}?${params.toString()}`;
    const proxiedUrl = `${localProxy}?title=${encodeURIComponent(title)}`;
    try {
        // On tente d'abord via le proxy local (pas de CORS depuis le navigateur)
        // let res = await fetch(proxiedUrl, { method: 'GET' });
        // if (!res.ok) {
        //     // fallback direct si le proxy n'est pas dispo
        //     res = await fetch(directUrl, { method: 'GET' });
        //     console.log('getDefinition: fetched direct', directUrl);
        // }

        let res = await fetch(directUrl, { method: "GET" });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        const pages = data?.query?.pages;
        if (!pages || typeof pages !== "object") return null;
        const firstKey = Object.keys(pages)[0];
        if (!firstKey) return null;
        const page = pages[firstKey];
        if (!page || page.missing === "" || !page.extract) return null;
        return {
            title: page.title || word,
            extract: page.extract as string,
            url: `https://fr.wiktionary.org/wiki/${encodeURIComponent(
                page.title || word
            )}`,
        };
    } catch (err) {
        console.error("getDefinition error", err);
        return null;
    }
}

export type DefinitionGroup = {
    type: string;
    definitions: string[];
};

/**
 * Extrait le type (POS + genre éventuel) et les définitions (sans exemples) d'une entrée Wiktionnaire.
 * @param html HTML complet de la page/section reçue (string)
 * @param langId id du span de langue (par défaut "fr")
 */
export function extractDefinitions(
    html: string,
    langId: string = "fr"
): DefinitionGroup[] {
    const doc = parseHTML(html);
    const frH2 = findLanguageH2(doc, langId);
    if (!frH2) return [];

    const groups: DefinitionGroup[] = [];
    const POS_RE =
        /^(Nom commun|Nom propre|Verbe|Adjectif|Adverbe|Pronom|Conjonction|Interjection|Préposition|Préfixe|Suffixe|Article|Déterminant|Participe|Locution[\s\-][\p{L}\- ]*)/iu;

    // On parcourt les éléments jusqu'au prochain H2 (changement de langue)
    for (
        let el = frH2.nextElementSibling;
        el && el.tagName !== "H2";
        el = el.nextElementSibling
    ) {
        if (el.tagName !== "H3") continue;

        // Ex: "Nom commun 1" → "Nom commun"
        const rawHeading = (el.textContent ?? "").trim().replace(/\s*\d+\s*$/u, "");
        const posMatch = rawHeading.match(POS_RE);
        if (!posMatch) continue;

        const pos = posMatch[0]; // ex. "Nom commun", "Verbe", "Locution nominale", ...

        // Cherche le <p> suivant pour récupérer un éventuel genre ("masculin", "féminin", "invariable", etc.)
        const p = nextElementOfTag(el, "P");
        const gender = extractGenderFromP(p);

        // Cherche la première liste de définitions <ol> après le <p>
        // (parfois il peut y avoir des noeuds intermédiaires, on balaie quelques frères)
        let cur = p ? p.nextElementSibling : el.nextElementSibling;
        let ol: HTMLOListElement | null = null;
        for (
            let i = 0;
            i < 5 && cur && cur.tagName !== "H3" && cur.tagName !== "H2";
            i++, cur = cur.nextElementSibling
        ) {
            if (cur.tagName === "OL") {
                ol = cur as HTMLOListElement;
                break;
            }
        }
        if (!ol) continue;

        // Récupère les <li> (définitions), en retirant les <ul> d'exemples
        const defs: string[] = [];
        ol.querySelectorAll(":scope > li").forEach((li) => {
            const liClone = li.cloneNode(true) as HTMLElement;
            // Supprime les UL d'exemples
            liClone.querySelectorAll("ul, ol ul").forEach((u) => u.remove());
            // Texte nettoyé
            const text = cleanText(liClone.textContent || "");
            if (text) defs.push(text);
        });

        if (defs.length) {
            groups.push({
                type: gender ? `${pos} ${gender}` : pos,
                definitions: defs,
            });
        }
    }

    return groups;
}

/* ------------------- Helpers ------------------- */

function parseHTML(html: string): Document {
    // Navigateur
    if (
        typeof window !== "undefined" &&
        typeof (window as any).DOMParser !== "undefined"
    ) {
        return new DOMParser().parseFromString(html, "text/html");
    }
    // Node.js (optionnel) via jsdom si dispo
    try {
        // eslint-disable-next-line @typescript-eslint/no-var-requires
        const { JSDOM } = require("jsdom");
        return new JSDOM(html).window.document;
    } catch {
        throw new Error("Aucun parser DOM disponible (DOMParser ou jsdom requis).");
    }
}

function findLanguageH2(
    doc: Document,
    langId: string
): HTMLHeadingElement | null {
    // Cherche un H2 contenant <span id="fr">Français</span> (id = langId)
    const h2s = Array.from(doc.querySelectorAll("h2"));
    for (const h2 of h2s) {
        if (h2.querySelector(`span#${CSS.escape(langId)}`))
            return h2 as HTMLHeadingElement;
    }
    return null;
}

function nextElementOfTag(start: Element, tagName: string): Element | null {
    const upper = tagName.toUpperCase();
    for (let el = start.nextElementSibling; el; el = el.nextElementSibling) {
        if (el.tagName === upper) return el;
        if (el.tagName === "H2" || el.tagName === "H3") break; // on ne traverse pas les sections suivantes
    }
    return null;
}

function extractGenderFromP(p: Element | null): string | null {
    if (!p) return null;
    // On prend les <i> du <p> uniquement (pour éviter les domaines en italique à l'intérieur des définitions)
    const candidates = Array.from(p.querySelectorAll(":scope i")).map((i) =>
        cleanText(i.textContent || "")
    );
    // On filtre sur quelques mots-clés usuels
    const known = candidates.find((t) =>
        /\b(masculin|féminin|invariable|épicène|pluriel|singulier|masculin et féminin|m\.\s*et\s*f\.)\b/i.test(
            t
        )
    );
    return known || null;
}

function cleanText(s: string): string {
    return s
        .replace(/\s+/g, " ") // espaces multiples → simple espace
        .replace(/\s*:\s*$/, "") // ":" final isolé
        .trim();
}
