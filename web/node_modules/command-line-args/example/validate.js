/*
  command-line-args parses the command line but does not validate what was collected.
  This is one method of testing the values received suit your taste.
*/

'use strict'
var commandLineArgs = require('../')
var testValue = require('test-value')
var fs = require('fs')

var cli = commandLineArgs([
  { name: 'help', type: Boolean },
  { name: 'files', type: String, multiple: true, defaultOption: true },
  { name: 'log-level', type: String }
])

var options = cli.parse()

/* all supplied files should exist and --log-level should be one from the list */
var correctUsageForm1 = {
  files: function (files) {
    return files && files.length && files.every(fs.existsSync)
  },
  'log-level': [ 'info', 'warn', 'error', undefined ]
}

/* passing a single --help flag is also valid */
var correctUsageForm2 = {
  help: true
}

/* test the options for usage forms 1 or 2 */
var valid = testValue(options, [ correctUsageForm1, correctUsageForm2 ])

console.log('your options are', valid ? 'valid' : 'invalid', options)
