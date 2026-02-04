"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.markdownToBlocks = markdownToBlocks;
/**
 * Converts markdown text to Notion block objects
 *
 * This is a simple, secure replacement for @tryfabric/martian's markdownToBlocks
 * to eliminate security vulnerabilities from the katex dependency chain.
 *
 * Supports:
 * - Headings (h1, h2, h3)
 * - Paragraphs
 * - Bulleted lists
 * - Numbered lists
 * - Code blocks
 * - Quotes
 * - Bold, italic, and inline code formatting
 *
 * @param markdown - The markdown string to convert
 * @returns Array of Notion block objects
 */
function markdownToBlocks(markdown) {
    const blocks = [];
    const lines = markdown.split('\n');
    let i = 0;
    while (i < lines.length) {
        const line = lines[i];
        const trimmedLine = line.trim();
        // Skip empty lines at the top level
        if (!trimmedLine) {
            i++;
            continue;
        }
        // Headings
        const headingMatch = trimmedLine.match(/^(#{1,3})\s+(.+)$/);
        if (headingMatch) {
            const level = headingMatch[1].length;
            const text = headingMatch[2];
            const headingType = level === 1 ? 'heading_1' : level === 2 ? 'heading_2' : 'heading_3';
            blocks.push({
                object: 'block',
                type: headingType,
                [headingType]: {
                    rich_text: parseRichText(text),
                },
            });
            i++;
            continue;
        }
        // Code blocks
        if (trimmedLine.startsWith('```')) {
            const language = trimmedLine.slice(3).trim() || 'plain text';
            const codeLines = [];
            i++;
            while (i < lines.length && !lines[i].trim().startsWith('```')) {
                codeLines.push(lines[i]);
                i++;
            }
            blocks.push({
                object: 'block',
                type: 'code',
                code: {
                    rich_text: [{
                            type: 'text',
                            text: { content: codeLines.join('\n') },
                        }],
                    language: language,
                },
            });
            i++; // Skip closing ```
            continue;
        }
        // Block quotes
        if (trimmedLine.startsWith('>')) {
            const quoteText = trimmedLine.slice(1).trim();
            blocks.push({
                object: 'block',
                type: 'quote',
                quote: {
                    rich_text: parseRichText(quoteText),
                },
            });
            i++;
            continue;
        }
        // Bulleted lists
        if (trimmedLine.match(/^[-*]\s+/)) {
            const text = trimmedLine.replace(/^[-*]\s+/, '');
            blocks.push({
                object: 'block',
                type: 'bulleted_list_item',
                bulleted_list_item: {
                    rich_text: parseRichText(text),
                },
            });
            i++;
            continue;
        }
        // Numbered lists
        if (trimmedLine.match(/^\d+\.\s+/)) {
            const text = trimmedLine.replace(/^\d+\.\s+/, '');
            blocks.push({
                object: 'block',
                type: 'numbered_list_item',
                numbered_list_item: {
                    rich_text: parseRichText(text),
                },
            });
            i++;
            continue;
        }
        // Horizontal rule
        if (trimmedLine.match(/^(-{3,}|\*{3,}|_{3,})$/)) {
            blocks.push({
                object: 'block',
                type: 'divider',
                divider: {},
            });
            i++;
            continue;
        }
        // Regular paragraph
        blocks.push({
            object: 'block',
            type: 'paragraph',
            paragraph: {
                rich_text: parseRichText(trimmedLine),
            },
        });
        i++;
    }
    return blocks;
}
/**
 * Parse markdown text into Notion rich text format
 * Supports: **bold**, *italic*, `code`, and [links](url)
 */
function parseRichText(text) {
    if (!text) {
        return [{ type: 'text', text: { content: '' } }];
    }
    const richText = [];
    let currentText = '';
    let i = 0;
    while (i < text.length) {
        // Bold: **text**
        if (text[i] === '*' && text[i + 1] === '*') {
            // Save any accumulated plain text
            if (currentText) {
                richText.push({ type: 'text', text: { content: currentText } });
                currentText = '';
            }
            // Find closing **
            i += 2;
            let boldText = '';
            while (i < text.length && !(text[i] === '*' && text[i + 1] === '*')) {
                boldText += text[i];
                i++;
            }
            richText.push({
                type: 'text',
                text: { content: boldText },
                annotations: { bold: true },
            });
            i += 2; // Skip closing **
            continue;
        }
        // Italic: *text* or _text_
        if ((text[i] === '*' || text[i] === '_') && text[i + 1] !== text[i]) {
            const marker = text[i];
            // Save any accumulated plain text
            if (currentText) {
                richText.push({ type: 'text', text: { content: currentText } });
                currentText = '';
            }
            // Find closing marker
            i++;
            let italicText = '';
            while (i < text.length && text[i] !== marker) {
                italicText += text[i];
                i++;
            }
            richText.push({
                type: 'text',
                text: { content: italicText },
                annotations: { italic: true },
            });
            i++; // Skip closing marker
            continue;
        }
        // Inline code: `text`
        if (text[i] === '`') {
            // Save any accumulated plain text
            if (currentText) {
                richText.push({ type: 'text', text: { content: currentText } });
                currentText = '';
            }
            // Find closing `
            i++;
            let codeText = '';
            while (i < text.length && text[i] !== '`') {
                codeText += text[i];
                i++;
            }
            richText.push({
                type: 'text',
                text: { content: codeText },
                annotations: { code: true },
            });
            i++; // Skip closing `
            continue;
        }
        // Links: [text](url)
        if (text[i] === '[') {
            const linkStart = i;
            let linkText = '';
            i++;
            // Find closing ]
            while (i < text.length && text[i] !== ']') {
                linkText += text[i];
                i++;
            }
            // Check if followed by (url)
            if (i < text.length && text[i] === ']' && text[i + 1] === '(') {
                i += 2; // Skip ](
                let url = '';
                while (i < text.length && text[i] !== ')') {
                    url += text[i];
                    i++;
                }
                // Save any accumulated plain text
                if (currentText) {
                    richText.push({ type: 'text', text: { content: currentText } });
                    currentText = '';
                }
                richText.push({
                    type: 'text',
                    text: { content: linkText, link: { url } },
                });
                i++; // Skip closing )
                continue;
            }
            else {
                // Not a link, treat as plain text
                currentText += text.slice(linkStart, i + 1);
                i++;
                continue;
            }
        }
        // Regular character
        currentText += text[i];
        i++;
    }
    // Add any remaining text
    if (currentText) {
        richText.push({ type: 'text', text: { content: currentText } });
    }
    return richText.length > 0 ? richText : [{ type: 'text', text: { content: '' } }];
}
