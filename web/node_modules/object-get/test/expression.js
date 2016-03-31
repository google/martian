var test = require('tape')
var objectGet = require('../')

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
  t.strictEqual(objectGet(fixture, 'one'), 1)
  t.strictEqual(objectGet(fixture, 'two.three'), 3)
  t.strictEqual(objectGet(fixture, 'two.four.five'), 5)
  t.deepEqual(objectGet(fixture, 'two'), {
    three: 3,
    four: {
      five: 5
    }
  })
  t.deepEqual(objectGet(fixture, 'two.four'), {
    five: 5
  })
  t.strictEqual(objectGet(fixture, 'ksfjglfshg'), undefined)
  t.end()
})
