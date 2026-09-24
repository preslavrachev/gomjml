// Replaces Math.random when loaded. mj-navbar and mj-carousel draw their ids
// from Math.random, so a fixed-seed PRNG (mulberry32) makes goldens
// byte-reproducible. The Go comparison masks these ids, so their values are
// not part of the contract. reset() restarts the sequence, so that a renderer
// that compiles many documents in one process gives each the ids a fresh CLI
// process would.
const initial = 0x6d6a6d6c
let state = initial

Math.random = function mulberry32() {
  state = (state + 0x6d2b79f5) | 0
  let t = Math.imul(state ^ (state >>> 15), 1 | state)
  t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296
}

exports.reset = () => {
  state = initial
}
