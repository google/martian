var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'file', multiple: true, type: function (file) {
    return file
  }}
]

test('type-other-multiple: different values', function (t) {
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--file', 'one.js' ]),
    { file: [ 'one.js' ] }
  )
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--file', 'one.js', 'two.js' ]),
    { file: [ 'one.js', 'two.js' ] }
  )
  t.deepEqual(
    cliArgs(optionDefinitions).parse([ '--file' ]),
    { file: [] }
  )

  t.end()
})
