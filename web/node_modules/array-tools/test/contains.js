var test = require('tape')
var a = require('../')

var fixture = [ 1, 2, 3 ]

test('a.contains(array, value)', function (t) {
  t.strictEqual(a.contains(fixture, 1), true)
  t.strictEqual(a.contains(fixture, 4), false)
  t.end()
})

test('a.contains(array, array)', function (t) {
  t.strictEqual(a.contains(fixture, [ 1, 2 ]), true)
  t.strictEqual(a.contains(fixture, [ 1, 2, 3, 4 ]), false)
  t.end()
})
