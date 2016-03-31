var test = require('tape')
var a = require('../')

test('a.remove(arr, item)', function (t) {
  var fixture = [ 1, 2, 3 ]
  var result = a.remove(fixture, 1)
  t.deepEqual(result, 1)
  t.deepEqual(fixture, [ 2, 3 ])
  t.end()
})

test('a(arr).remove(item)', function (t) {
  var fixture = [ 1, 2, 3 ]
  var result = a(fixture).remove(1)
  t.deepEqual(result, 1)
  t.deepEqual(fixture, [ 2, 3 ])
  t.end()
})
