var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'array', type: Boolean, multiple: true }
]

test('type-boolean-multiple: 1', function (t) {
  var argv = [ '--array', '--array', '--array' ]
  var result = cliArgs(optionDefinitions).parse(argv)
  t.deepEqual(result, {
    array: [ true, true, true ]
  })
  t.end()
})
