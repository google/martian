var test = require('tape')
var o = require('../')

test('.select(object, fields)', function (t) {
  var object = {
    one: 1,
    two: 2,
    three: 3,
    four: 4
  }
  t.deepEqual(o.select(object, [ 'two', 'three' ]), {
    two: 2,
    three: 3
  })
  t.end()
})
