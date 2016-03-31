var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'one', type: Boolean }
]

test('type-boolean: different values', function (t) {
  var cli = cliArgs(optionDefinitions)
  t.deepEqual(
    cli.parse([ '--one' ]),
    { one: true }
  )
  t.deepEqual(
    cli.parse([ '--one', 'true' ]),
    { one: true }
  )
  t.deepEqual(
    cli.parse([ '--one', 'false' ]),
    { one: true }
  )
  t.deepEqual(
    cli.parse([ '--one', 'sfsgf' ]),
    { one: true }
  )

  t.end()
})
