/**
 * Utility helpers for typography transformations.
 */

export type NbspType = 'normal' | 'narrow';

/**
 * Protect French punctuation by replacing the normal space before certain
 * punctuation marks with a non-breaking space.
 *
 * @param text - input string
 * @param nbspType - 'normal' uses U+00A0, 'narrow' uses U+202F
 */
export function protectFrenchPunctuation(text: string, nbspType: NbspType = 'normal'): string {
    if (!text) return text;
    const nbsp = nbspType === 'narrow' ? '\u202F' : '\u00A0';
    // replace a normal space before common French punctuation with the chosen NBSP
    return text.replace(/ (\?|!|:|;|»)/g, `${nbsp}$1`);
}

export default protectFrenchPunctuation;

/**
 * Escape HTML special chars to avoid XSS when rendering as HTML.
 */
function escapeHtml(str: string): string {
    return str.replace(/[&<>"']/g, function (c) {
        switch (c) {
            case '&': return '&amp;';
            case '<': return '&lt;';
            case '>': return '&gt;';
            case '"': return '&quot;';
            case "'": return '&#39;';
            default: return c;
        }
    });
}

/**
 * Format a message string applying simple lightweight markdown-like rules.
 * - **bold** -> <b>
 * - __underline__ -> <u>
 * - *italic* or _italic_ -> <em>
 * - ~~strike~~ -> <s>
 * - `inline code` -> <code>
 * It also protects French punctuation and escapes HTML.
 */
export function formatMessage(raw: string, nbspType: NbspType = 'normal'): string {
    if (!raw) return '';
    // remove leading/trailing whitespace before formatting
    raw = raw.trim();
    // escape first
    let s = escapeHtml(raw);
    // protect punctuation
    s = protectFrenchPunctuation(s, nbspType);

    // Protect inline code first by replacing them with placeholders so they
    // won't be altered by URL/link or other markup replacements.
    const codePlaceholders: string[] = [];
    s = s.replace(/`([^`]+)`/g, (_m, p1) => {
        const idx = codePlaceholders.length;
        codePlaceholders.push(p1); // already escaped
        return `@@CODE_${idx}@@`;
    });

    // Convert URLs (http(s):// or www.) to links. We trim trailing punctuation
    // commonly attached to sentences (.,:;!?) so the punctuation is not part of the URL.
    s = s.replace(/(https?:\/\/[^\s<]+)|(^|\s)(www\.[^\s<]+)/g, (m, g1, g2, g3, offset, str) => {
        let url = g1 || g3;
        const prefixSpace = g2 || '';
        if (!url) return m;

        // If url comes from the second capture (www...), ensure it has protocol
        const href = url.startsWith('www.') ? `https://${url}` : url;

        // Trim trailing punctuation characters . , : ; ! ? ) ]
        while (/[\.,:;!?)\]]$/.test(url)) {
            url = url.slice(0, -1);
        }

        // escape href for attribute safety
        const safeHref = escapeHtml(href);
        const safeText = escapeHtml(url);
        return `${prefixSpace}<a href="${safeHref}" target="_blank" rel="noopener noreferrer" class="text-blue-200 hover:underline">${safeText}</a>`;
    });

    // bold: **text**
    s = s.replace(/\*\*([^*]+)\*\*/g, '<b>$1</b>');
    // underline: __text__
    s = s.replace(/__([^_]+)__/g, '<u>$1</u>');
    // italic: *text* or _text_
    s = s.replace(/\*([^*]+)\*/g, '<em>$1</em>');
    s = s.replace(/_([^_]+)_/g, '<em>$1</em>');
    // strike: ~~text~~
    s = s.replace(/~~([^~]+)~~/g, '<s>$1</s>');

    // restore code placeholders as <code>
    s = s.replace(/@@CODE_(\d+)@@/g, (_m, id) => {
        const idx = parseInt(id, 10);
        const original = codePlaceholders[idx] ?? '';
        return `<code>${original}</code>`;
    });

    // preserve newlines as <br>
    s = s.replace(/\r?\n/g, '<br>');

    return s;
}

/**
 * Extract plain text from a message with HTML formatting.
 * @param raw - The raw message string with HTML
 * @returns The extracted plain text
 */
export function extractTextFromMessage(raw: string): string {
    if (!raw) return '';
    // remove leading/trailing whitespace before formatting
    raw = raw.trim();
    // escape first
    let s = escapeHtml(raw);
    // protect punctuation

    // inline code: `code`
    s = s.replace(/`([^`]+)`/g, '$1');
    // bold: **text**
    s = s.replace(/\*\*([^*]+)\*\*/g, '$1');
    // underline: __text__
    s = s.replace(/__([^_]+)__/g, '$1');
    // italic: *text* or _text_
    s = s.replace(/\*([^*]+)\*/g, '$1');
    s = s.replace(/_([^_]+)_/g, '$1');
    // strike: ~~text~~
    s = s.replace(/~~([^~]+)~~/g, '$1');

    // preserve newlines as <br>
    s = s.replace(/\r?\n/g, '\n');

    return s;
}
