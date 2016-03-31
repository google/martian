module.exports = [
  { name: 'help', alias: 'h', type: Boolean, description: 'Display this usage guide.' },
  { name: 'src', type: String, multiple: true, defaultOption: true, description: 'The input files to process', typeLabel: '<files>' },
  { name: 'timeout', alias: 't', type: Number, description: 'Timeout value in ms', typeLabel: '<ms>' },
  { name: 'log', alias: 'l', type: Boolean, description: 'info, warn or error' }
]
