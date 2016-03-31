var test = require('tape')
var o = require('../')

test('.get(object, expression)', function (t) {
  var fixture = {
    one: 1,
    two: {
      three: 3,
      four: {
        five: 5
      }
    }
  }
  t.strictEqual(o.get(fixture, 'one'), 1)
  t.strictEqual(o.get(fixture, 'two.three'), 3)
  t.strictEqual(o.get(fixture, 'two.four.five'), 5)
  t.deepEqual(o.get(fixture, 'two'), {
    three: 3,
    four: {
      five: 5
    }
  })
  t.deepEqual(o.get(fixture, 'two.four'), {
    five: 5
  })
  t.strictEqual(o.get(fixture, 'ksfjglfshg'), undefined)
  t.end()
})
