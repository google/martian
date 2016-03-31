var test = require('tape')
var wrap = require('../')

test('non-string input', function (t) {
  t.strictEqual(wrap(undefined), '')
  t.strictEqual(wrap(function () {}), 'function () {}')
  t.strictEqual(wrap({}), '[object Object]')
  t.strictEqual(wrap(null), 'null')
  t.strictEqual(wrap(true), 'true')
  t.strictEqual(wrap(0), '0')
  t.strictEqual(wrap(NaN), 'NaN')
  t.strictEqual(wrap(Infinity), 'Infinity')
  t.end()
})
