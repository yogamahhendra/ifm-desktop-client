import MarkdownIt from 'markdown-it'
import DOMPurify from 'dompurify'
import { createHighlighterCore } from 'shiki/core'
import { createJavaScriptRegexEngine } from 'shiki/engine/javascript'
import githubDark from 'shiki/themes/github-dark.mjs'

import langJs from 'shiki/langs/javascript.mjs'
import langTs from 'shiki/langs/typescript.mjs'
import langGo from 'shiki/langs/go.mjs'
import langPy from 'shiki/langs/python.mjs'
import langBash from 'shiki/langs/bash.mjs'
import langJson from 'shiki/langs/json.mjs'
import langSql from 'shiki/langs/sql.mjs'
import langHtml from 'shiki/langs/html.mjs'
import langCss from 'shiki/langs/css.mjs'
import langYaml from 'shiki/langs/yaml.mjs'
import langMd from 'shiki/langs/markdown.mjs'
import langRust from 'shiki/langs/rust.mjs'
import langCsharp from 'shiki/langs/csharp.mjs'
import langCpp from 'shiki/langs/cpp.mjs'
import langDockerfile from 'shiki/langs/dockerfile.mjs'

let highlighter: any = null

const LOADED_LANGS = [
  langJs, langTs, langGo, langPy, langBash,
  langJson, langSql, langHtml, langCss, langYaml,
  langMd, langRust, langCsharp, langCpp, langDockerfile
]

async function initHighlighter() {
  if (highlighter) return highlighter
  try {
    highlighter = await createHighlighterCore({
      themes: [githubDark],
      langs: LOADED_LANGS,
      engine: createJavaScriptRegexEngine()
    })
  } catch (e) {
    console.error('Failed to initialize Shiki highlighter:', e)
  }
  return highlighter
}

initHighlighter()

const LANG_ALIASES: Record<string, string> = {
  sqlicommandline: 'sql',
  sqlite: 'sql',
  sqlite3: 'sql',
  mysql: 'sql',
  pgsql: 'sql',
  postgres: 'sql',
  postgresql: 'sql',
  js: 'javascript',
  node: 'javascript',
  ts: 'typescript',
  py: 'python',
  sh: 'bash',
  shell: 'bash',
  zsh: 'bash',
  golang: 'go',
  yml: 'yaml',
  md: 'markdown',
  dockerfile: 'dockerfile',
  rust: 'rust',
  rs: 'rust',
  csharp: 'csharp',
  'c#': 'csharp',
  cpp: 'cpp',
  'c++': 'cpp',
}

const DISPLAY_NAMES: Record<string, string> = {
  sql: 'SQL',
  javascript: 'JavaScript',
  typescript: 'TypeScript',
  go: 'Go',
  python: 'Python',
  bash: 'Bash',
  json: 'JSON',
  html: 'HTML',
  css: 'CSS',
  yaml: 'YAML',
  markdown: 'Markdown',
  rust: 'Rust',
  csharp: 'C#',
  cpp: 'C++',
  dockerfile: 'Dockerfile',
  text: 'Text'
}

const md = new MarkdownIt({
  html: false,
  linkify: true,
  typographer: true,
  breaks: true,
  highlight: (code, lang) => {
    const rawLang = (lang || '').toLowerCase().trim()
    const normalizedLang = LANG_ALIASES[rawLang] || rawLang || 'text'
    const displayLang = DISPLAY_NAMES[normalizedLang] || (rawLang ? rawLang.toUpperCase() : 'CODE')
    const trimmedCode = code.trim()
    const encodedCode = encodeURIComponent(trimmedCode)

    if (highlighter && highlighter.getLoadedLanguages().includes(normalizedLang)) {
      try {
        const highlighted = highlighter.codeToHtml(trimmedCode, {
          lang: normalizedLang,
          theme: 'github-dark'
        })
        return `<div class="code-block-container my-3 rounded-lg overflow-hidden border border-[#27272a] bg-[#121214]"><div class="flex items-center justify-between px-3 py-1.5 bg-[#18181b] border-b border-[#27272a] text-xs text-[#a1a1aa] select-none"><span class="font-mono text-[11px] font-medium tracking-wide">${escapeHtml(displayLang)}</span><button class="copy-code-btn flex items-center gap-1 hover:text-[#fafafa] transition-colors py-0.5 px-2 rounded hover:bg-[#27272a]" data-code="${encodedCode}"><span>Copy</span></button></div><div class="code-content p-3.5 overflow-x-auto text-xs font-mono leading-relaxed">${highlighted}</div></div>`
      } catch (err) {
        console.warn('Shiki highlight failed:', err)
      }
    }

    return `<div class="code-block-container my-3 rounded-lg overflow-hidden border border-[#27272a] bg-[#121214]"><div class="flex items-center justify-between px-3 py-1.5 bg-[#18181b] border-b border-[#27272a] text-xs text-[#a1a1aa] select-none"><span class="font-mono text-[11px] font-medium tracking-wide">${escapeHtml(displayLang)}</span><button class="copy-code-btn flex items-center gap-1 hover:text-[#fafafa] transition-colors py-0.5 px-2 rounded hover:bg-[#27272a]" data-code="${encodedCode}"><span>Copy</span></button></div><div class="code-content p-3.5 overflow-x-auto text-xs font-mono leading-relaxed"><pre class="m-0 p-0 bg-transparent text-[#fafafa] border-0"><code class="p-0 bg-transparent border-0">${escapeHtml(trimmedCode)}</code></pre></div></div>`
  }
})

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

export function renderMarkdown(content: string): string {
  if (!content) return ''
  // Clean up excessive blank lines (more than 2 consecutive newlines)
  const cleanedContent = content.replace(/\n{3,}/g, '\n\n')
  const rawHtml = md.render(cleanedContent)
  return DOMPurify.sanitize(rawHtml, {
    ADD_ATTR: ['data-code', 'target'],
    ADD_TAGS: ['button']
  })
}
