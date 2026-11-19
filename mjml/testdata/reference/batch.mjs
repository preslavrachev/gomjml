// Renders many MJML documents in one process, for the differential harness
// (mjml/differential_test.go).
//
//   node batch.mjs <manifest.jsonl> <results.jsonl>
//
// Each manifest line is a case, {"name", "mjml", "pkg"?, "filePath"?}, where
// pkg is a package.json alias of mjml (default "mjml") and filePath resolves
// mj-include. Each result line is {"name", "html"} or {"name", "error"}, with
// the validation "warnings" MJML reports in its default soft mode. compile.mjs
// has the options, which are the CLI's.
import { createReadStream, createWriteStream } from 'node:fs'
import { once } from 'node:events'
import readline from 'node:readline'
import { compile, version } from './compile.mjs'

const [manifest, output] = process.argv.slice(2)
if (!manifest || !output) {
  console.error('usage: node batch.mjs <manifest.jsonl> <results.jsonl>')
  process.exit(2)
}

const out = createWriteStream(output)
let rendered = 0
let rejected = 0
for await (const line of readline.createInterface({ input: createReadStream(manifest), crlfDelay: Infinity })) {
  if (!line.trim()) continue
  const { name, mjml, pkg, filePath } = JSON.parse(line)
  const result = compile(mjml, { pkg, filePath })
  if (result.error) rejected++
  else rendered++
  if (!out.write(`${JSON.stringify({ name, ...result })}\n`)) await once(out, 'drain')
}
out.end()
await once(out, 'finish')
console.log(`mjml ${version()}: ${rendered} rendered, ${rejected} rejected`)
