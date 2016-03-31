const getUsage = require('../')

const optionDefinitions = [
  {
    name: 'help', description: 'Display this usage guide.',
    alias: 'h', type: Boolean,
    group: 'main'
  },
  {
    name: 'src', description: 'The input files to process',
    type: String, multiple: true, defaultOption: true, typeLabel: '[underline]{file} ...',
    group: 'main'
  },
  {
    name: 'timeout', description: 'Timeout value in ms',
    alias: 't', type: Number, typeLabel: '[underline]{ms}',
    group: 'main'
  },
  {
    name: 'plugin', description: 'A plugin path',
    type: String
  }
]

const options = {
  title: 'a typical app',
  description: 'Generates something [italic]{very} important.',
  groups: {
    main: 'Main options',
    _none: {
      title: 'Misc',
      description: 'Miscelaneous ungrouped options.'
    }
  }
}

console.log(getUsage(optionDefinitions, options))
