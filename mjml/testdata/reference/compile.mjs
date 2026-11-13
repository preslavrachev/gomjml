// Compiles MJML in-process the way the pinned CLI does for
//
//   mjml <file> -o <out> --config.beautify false --config.minify true
//
// mirroring mjml-cli/lib/client.js: mjml-core with the preset components, the
// CLI's default "soft" validation, then html-minifier with the CLI's
// minifyConfig. Each document restarts the seeded Math.random, as a fresh CLI
// process would.
import { readFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const require = createRequire(import.meta.url)
const here = path.dirname(fileURLToPath(import.meta.url))
const { reset } = require('./seed-random.cjs')

// Copied from mjml-cli/lib/client.js.
const minifyConfig = {
  collapseWhitespace: true,
  minifyCSS: false,
  caseSensitive: true,
  removeEmptyAttributes: true,
}

// The CLI loads brace-expansion (through glob), which draws from Math.random
// five times at load, before the document compiles; skipping as many keeps the
// generated ids the CLI writes.
const cliStartupDraws = 5

const dependencies = JSON.parse(await readFile(path.join(here, 'package.json'), 'utf8')).dependencies

const compilers = new Map()

// compiler loads a package.json alias of mjml, such as "mjml" or "mjml-4.17",
// after checking that the installed version is the pinned one.
function compiler(pkg) {
  if (compilers.has(pkg)) return compilers.get(pkg)
  const pinned = dependencies[pkg]?.replace(/^npm:mjml@/, '')
  const installed = require(`${pkg}/package.json`).version
  if (installed !== pinned) {
    throw new Error(`${pkg} ${installed} is installed but package.json pins ${pinned}; run npm ci`)
  }
  const pkgRequire = createRequire(require.resolve(`${pkg}/package.json`))
  const mjml2html = pkgRequire(`${pkg}/lib/index.js`)
  const cliRequire = createRequire(pkgRequire.resolve('mjml-cli/package.json'))
  const { minify } = cliRequire('html-minifier')
  const entry = { version: installed, mjml2html, minify }
  compilers.set(pkg, entry)
  return entry
}

export function version(pkg = 'mjml') {
  return compiler(pkg).version
}

// compile returns { html, warnings } or { error, warnings } when MJML rejects
// the input, error being the message the CLI prints after "Error: ".
export function compile(mjml, { pkg, filePath } = {}) {
  const { mjml2html, minify } = compiler(pkg || 'mjml')
  reset()
  for (let i = 0; i < cliStartupDraws; i++) Math.random()
  try {
    const result = mjml2html(mjml, { filePath: filePath || undefined, actualPath: filePath || undefined })
    const warnings = (result.errors ?? []).map(({ line, tagName, message }) => ({ line, tagName, message }))
    return { html: minify(result.html, minifyConfig), warnings }
  } catch (err) {
    return { error: err.message, warnings: [] }
  }
}
