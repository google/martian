const getUsage = require('../')
const header = require('./assets/ansi-header')
const ansi = require('ansi-escape-sequences')

const optionDefinitions = [
  {
    name: 'help', description: 'Display this usage guide.',
    alias: 'h', type: Boolean
  },
  {
    name: 'src', description: 'The input files to process',
    type: String, multiple: true, defaultOption: true, typeLabel: '[underline]{file} ...'
  },
  {
    name: 'timeout', description: 'Timeout value in ms. This description is needlessly long unless you count testing of the description column maxWidth useful.',
    alias: 't', type: Number, typeLabel: '[underline]{ms}'
  }
]

const options = {
  header: ansi.format(header, 'red'),
  synopsis: [
    '$ example [[bold]{--timeout} [underline]{ms}] [bold]{--src} [underline]{file} ...',
    '$ example [bold]{--help}'
  ]
}

console.log(getUsage(optionDefinitions, options))
