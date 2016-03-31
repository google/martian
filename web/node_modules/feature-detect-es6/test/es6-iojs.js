var detect = require('../')
var test = require('tape')

test('.class()', function (t) {
  t.strictEqual(detect.class(), true)
  t.end()
})

test('.arrowFunction()', function (t) {
  t.strictEqual(detect.arrowFunction(), false)
  t.end()
})

test('.let()', function (t) {
  t.strictEqual(detect.let(), true)
  t.end()
})

test('.const()', function (t) {
  t.strictEqual(detect.const(), true)
  t.end()
})

test('.newArrayFeatures()', function (t) {
  t.strictEqual(detect.newArrayFeatures(), false)
  t.end()
})

test('.collections()', function (t) {
  t.strictEqual(detect.collections(), true)
  t.end()
})

test('.generators()', function (t) {
  t.strictEqual(detect.generators(), true)
  t.end()
})

test('.promises()', function (t) {
  t.strictEqual(detect.generators(), true)
  t.end()
})

test('.templateStrings()', function (t) {
  t.strictEqual(detect.templateStrings(), true)
  t.end()
})

test('.symbols()', function (t) {
  t.strictEqual(detect.symbols(), true)
  t.end()
})

test('.destructuring', function (t) {
  t.strictEqual(detect.destructuring(), false)
  t.end()
})

test('.spread', function (t) {
  t.strictEqual(detect.spread(), false)
  t.end()
})
