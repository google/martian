const getUsage = require('../')
const ussr = require('./assets/ascii-ussr')

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
  title: 'brezhnev',
  description: {
    options: {
      columns: [
        { name: 'one', maxWidth: 40 },
        { name: 'two', width: 40, nowrap: true }
      ]
    },
    data: [
      {
        one: 'On his 70th birthday he was awarded the rank of Marshal of the Soviet Union – the highest military honour in the Soviet Union. After being awarded the medal, he attended an 18th Army Veterans meeting, dressed in a long coat and saying; "Attention, Marshal\'s coming!" He also conferred upon himself the rare [bold]{Order of Victory} in 1978 — the only time the decoration was ever awarded outside of World War II. (This medal was posthumously revoked in 1989 for not meeting the criteria for citation.) \n\nBrezhnev\'s weakness for undeserved medals was proven by his poorly written memoirs recalling his military service during World War II, which treated the little-known and minor Battle of Novorossiysk as the decisive military theatre.',
        two: ussr
      }
    ]
  },
  synopsis: [
    '$ example [[bold]{--timeout} [underline]{ms}] [bold]{--src} [underline]{file} ...',
    '$ example [bold]{--help}'
  ]
}

console.log(getUsage(optionDefinitions, options))
