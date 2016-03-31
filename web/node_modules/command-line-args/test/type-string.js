var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'one', type: String }
]

test('type-string: different values', function (t) {
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--one', 'yeah' ]),
    { one: 'yeah' }
  )
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--one' ]),
    { one: null }
  )
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--one', '3' ]),
    { one: '3' }
  )

  t.end()
})
