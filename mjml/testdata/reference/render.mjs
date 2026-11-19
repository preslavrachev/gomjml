// Renders every mjml/testdata/*.mjml fixture with the pinned MJML, as
//
//   mjml <name>.mjml -o <name>.html --config.beautify false --config.minify true
//
// would (see compile.mjs), which is the invocation that reproduces the goldens
// committed before this generator existed. Validation stays at the CLI default
// ("soft"). When MJML rejects an input, its error message is written to
// <name>.error instead.
import { readdir, readFile, rm, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { compile, version } from './compile.mjs'

const here = path.dirname(fileURLToPath(import.meta.url))
const testdata = path.resolve(here, '..')

// Fixtures for features newer than the main pin, rendered by the package.json
// alias of the release that introduced them.
const overrides = {
  'mj-wrapper-gap': 'mjml-4.17', // mj-wrapper gap arrived in MJML 4.17
}

const names = (await readdir(testdata))
  .filter((file) => file.endsWith('.mjml'))
  .map((file) => file.slice(0, -'.mjml'.length))
  .sort()
const unknown = Object.keys(overrides).filter((name) => !names.includes(name))
if (unknown.length) {
  console.error(`overrides name missing fixtures: ${unknown.join(', ')}`)
  process.exit(1)
}

const results = []
for (const name of names) {
  const pkg = overrides[name] ?? 'mjml'
  const html = path.join(testdata, `${name}.html`)
  const error = path.join(testdata, `${name}.error`)
  await rm(html, { force: true })
  await rm(error, { force: true })
  const result = compile(await readFile(path.join(testdata, `${name}.mjml`), 'utf8'), { pkg, filePath: path.join(testdata, `${name}.mjml`) })
  if (result.error) await writeFile(error, `${result.error}\n`)
  else await writeFile(html, result.html)
  results.push({ name, pkg, ...result })
}

const rejected = results.filter((r) => r.error)
console.log(`mjml ${version()}: ${results.length - rejected.length} rendered, ${rejected.length} rejected`)
for (const r of results.filter((r) => r.pkg !== 'mjml')) {
  console.log(`  ${r.name} rendered with mjml ${version(r.pkg)}`)
}
for (const r of rejected) console.log(`  rejected ${r.name}: ${r.error}`)
for (const r of results.filter((r) => r.warnings.length)) {
  const lines = r.warnings.map((w) => `Line ${w.line} of ${r.name}.mjml (${w.tagName}) — ${w.message}`)
  console.log(`  warnings ${r.name}:\n    ${lines.join('\n    ')}`)
}
