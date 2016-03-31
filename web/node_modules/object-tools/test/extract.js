var test = require('tape')
var o = require('../')

test('.extract(object, function)', function (t) {
  var object = { a: 1, b: 0, c: 2 }
  var result = o.extract(object, function (value, key) {
    return value > 0
  })
  t.deepEqual(result, { a: 1, c: 2 })
  t.deepEqual(object, { b: 0 })
  t.end()
})

test('.extract(object, propertyArray)', function (t) {
  var object = { a: 1, b: 0, c: 2 }
  var result = o.extract(object, [ 'b' ])
  t.deepEqual(result, { b: 0 })
  t.deepEqual(object, { a: 1, c: 2 })
  t.end()
})
