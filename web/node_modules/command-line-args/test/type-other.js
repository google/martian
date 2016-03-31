var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'file', type: function (file) {
    return file
  }}
]

test('type-other: different values', function (t) {
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--file', 'one.js' ]),
    { file: 'one.js' }
  )
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--file' ]),
    { file: null }
  )

  t.end()
})
