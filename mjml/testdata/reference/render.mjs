// Renders every mjml/testdata/*.mjml fixture with the pinned MJML CLI.
//
// Each fixture gets its own CLI process, run as
//
//   mjml <name>.mjml -o <name>.html --config.beautify false --config.minify true
//
// which is the invocation that reproduces the goldens committed before this
// generator existed. Validation stays at the CLI default ("soft"). When MJML
// rejects an input, its error message is written to <name>.error instead.
import { execFile } from 'node:child_process'
import { readdir, readFile, rm, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { availableParallelism } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { promisify } from 'node:util'

const run = promisify(execFile)
const require = createRequire(import.meta.url)
const here = path.dirname(fileURLToPath(import.meta.url))
const testdata = path.resolve(here, '..')
const seed = path.join(here, 'seed-random.cjs')
const flags = ['--config.beautify', 'false', '--config.minify', 'true']

// Fixtures for features newer than the main pin, rendered by the package.json
// alias of the release that introduced them.
const overrides = {
  'mj-wrapper-gap': 'mjml-4.17', // mj-wrapper gap arrived in MJML 4.17
}

const dependencies = JSON.parse(await readFile(path.join(here, 'package.json'), 'utf8')).dependencies
for (const pkg of new Set(['mjml', ...Object.values(overrides)])) {
  const pinned = dependencies[pkg]?.replace(/^npm:mjml@/, '')
  const installed = require(`${pkg}/package.json`).version
  if (installed !== pinned) {
    console.error(`${pkg} ${installed} is installed but package.json pins ${pinned}; run npm ci`)
    process.exit(1)
  }
}

async function render(name) {
  const pkg = overrides[name] ?? 'mjml'
  const html = `${name}.html`
  const error = path.join(testdata, `${name}.error`)
  await rm(path.join(testdata, html), { force: true })
  await rm(error, { force: true })
  try {
    const { stderr } = await run(
      process.execPath,
      ['--require', seed, require.resolve(`${pkg}/bin/mjml`), `${name}.mjml`, '-o', html, ...flags],
      { cwd: testdata },
    )
    return { name, pkg, warnings: stderr.trim() }
  } catch (err) {
    const message = /^Error: (.+)$/m.exec(err.stderr ?? '')?.[1]
    if (!message) throw err
    await writeFile(error, `${message}\n`)
    return { name, pkg, rejected: message }
  }
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
let next = 0
await Promise.all(
  Array.from({ length: availableParallelism() }, async () => {
    while (next < names.length) results.push(await render(names[next++]))
  }),
)
results.sort((a, b) => a.name.localeCompare(b.name))

const rejected = results.filter((r) => r.rejected)
console.log(`mjml ${require('mjml/package.json').version}: ${results.length - rejected.length} rendered, ${rejected.length} rejected`)
for (const r of results.filter((r) => r.pkg !== 'mjml')) {
  console.log(`  ${r.name} rendered with mjml ${require(`${r.pkg}/package.json`).version}`)
}
for (const r of rejected) console.log(`  rejected ${r.name}: ${r.rejected}`)
for (const r of results.filter((r) => r.warnings)) {
  console.log(`  warnings ${r.name}:\n    ${r.warnings.split('\n').join('\n    ')}`)
}
