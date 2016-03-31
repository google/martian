var test = require('tape')
var cliArgs = require('../')

var optionDefinitions = [
  { name: 'array', type: Number, multiple: true }
]

test('number multiple: 1', function (t) {
  var argv = [ '--array', '1', '2', '3' ]
  var result = cliArgs(optionDefinitions).parse(argv)
  t.deepEqual(result, {
    array: [ 1, 2, 3 ]
  })
  t.notDeepEqual(result, {
    array: [ '1', '2', '3' ]
  })
  t.end()
})

test('number multiple: 2', function (t) {
  var argv = [ '--array', '1', '--array', '2', '--array', '3' ]
  var result = cliArgs(optionDefinitions).parse(argv)
  t.deepEqual(result, {
    array: [ 1, 2, 3 ]
  })
  t.notDeepEqual(result, {
    array: [ '1', '2', '3' ]
  })
  t.end()
})
