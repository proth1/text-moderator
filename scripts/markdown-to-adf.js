#!/usr/bin/env node
/**
 * Markdown to Atlassian Document Format (ADF) Converter
 * Converts markdown to ADF JSON for Jira REST API v3
 * Usage: node markdown-to-adf.js "markdown text"
 */
const input = process.argv[2] || '';

function convertMarkdownToADF(markdown) {
  const content = [];
  const lines = markdown.split('\n');
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Code blocks
    if (line.match(/^```(\w*)/)) {
      const lang = line.match(/^```(\w*)/)[1] || null;
      const codeLines = [];
      i++;
      while (i < lines.length && !lines[i].startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      content.push({
        type: 'codeBlock',
        attrs: lang ? { language: lang } : {},
        content: [{ type: 'text', text: codeLines.join('\n') }]
      });
      i++;
      continue;
    }

    // Headers
    const h3 = line.match(/^### (.+)$/);
    if (h3) { content.push({ type: 'heading', attrs: { level: 3 }, content: [{ type: 'text', text: h3[1] }] }); i++; continue; }

    const h2 = line.match(/^## (.+)$/);
    if (h2) { content.push({ type: 'heading', attrs: { level: 2 }, content: [{ type: 'text', text: h2[1] }] }); i++; continue; }

    const h1 = line.match(/^# (.+)$/);
    if (h1) { content.push({ type: 'heading', attrs: { level: 1 }, content: [{ type: 'text', text: h1[1] }] }); i++; continue; }

    // Bullet lists
    if (line.match(/^[-*] /)) {
      const items = [];
      while (i < lines.length && lines[i].match(/^[-*] /)) {
        items.push({ type: 'listItem', content: [{ type: 'paragraph', content: parseInline(lines[i].replace(/^[-*] /, '')) }] });
        i++;
      }
      content.push({ type: 'bulletList', content: items });
      continue;
    }

    // Numbered lists
    if (line.match(/^\d+\. /)) {
      const items = [];
      while (i < lines.length && lines[i].match(/^\d+\. /)) {
        items.push({ type: 'listItem', content: [{ type: 'paragraph', content: parseInline(lines[i].replace(/^\d+\. /, '')) }] });
        i++;
      }
      content.push({ type: 'orderedList', content: items });
      continue;
    }

    // Empty line - skip
    if (line.trim() === '') { i++; continue; }

    // Regular paragraph
    content.push({ type: 'paragraph', content: parseInline(line) });
    i++;
  }

  return { type: 'doc', version: 1, content };
}

function parseInline(text) {
  const result = [];
  let remaining = text;

  while (remaining.length > 0) {
    // Bold **text**
    const bold = remaining.match(/^(.*?)\*\*(.+?)\*\*(.*)/s);
    if (bold) {
      if (bold[1]) result.push(...parseInline(bold[1]));
      result.push({ type: 'text', text: bold[2], marks: [{ type: 'strong' }] });
      remaining = bold[3];
      continue;
    }

    // Inline code `text`
    const code = remaining.match(/^(.*?)`([^`]+)`(.*)/s);
    if (code) {
      if (code[1]) result.push(...parseInline(code[1]));
      result.push({ type: 'text', text: code[2], marks: [{ type: 'code' }] });
      remaining = code[3];
      continue;
    }

    // Links [text](url)
    const link = remaining.match(/^(.*?)\[([^\]]+)\]\(([^)]+)\)(.*)/s);
    if (link) {
      if (link[1]) result.push(...parseInline(link[1]));
      result.push({ type: 'text', text: link[2], marks: [{ type: 'link', attrs: { href: link[3] } }] });
      remaining = link[4];
      continue;
    }

    // Italic *text*
    const italic = remaining.match(/^(.*?)\*([^*]+)\*(.*)/s);
    if (italic) {
      if (italic[1]) result.push(...parseInline(italic[1]));
      result.push({ type: 'text', text: italic[2], marks: [{ type: 'em' }] });
      remaining = italic[3];
      continue;
    }

    result.push({ type: 'text', text: remaining });
    break;
  }

  return result.filter(r => r.text);
}

console.log(JSON.stringify(convertMarkdownToADF(input)));
