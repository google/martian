const getUsage = require('../')

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
    name: 'timeout', description: 'Timeout value in ms',
    alias: 't', type: Number, typeLabel: '[underline]{ms}'
  }
]

const options = {
  title: 'a typical app',
  description: 'Generates something very important.',
  synopsis: [
    '$ example [[bold]{--timeout} [underline]{ms}] [bold]{--src} [underline]{file} ...',
    '$ example [bold]{--help}'
  ],
  examples: [
    {
      desc: '1. A concise example. ',
      example: '$ example -t 100 lib/*.js'
    },
    {
      desc: '2. A long example. ',
      example: '$ example --timeout 100 --src lib/*.js'
    },
    {
      desc: '3. This example will scan space for unknown things. Take cure when scanning space, it could take some time. ',
      example: '$ example --src galaxy1.facts galaxy1.facts galaxy2.facts galaxy3.facts galaxy4.facts galaxy5.facts'
    }
  ],
  footer: 'Project home: [underline]{https://github.com/me/example}'
}

console.log(getUsage(optionDefinitions, options))
