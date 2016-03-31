'use strict'
var columnLayout = require('column-layout')
var o = require('object-tools')
var ansi = require('ansi-escape-sequences')
var os = require('os')
var t = require('typical')
var UsageOptions = require('./usage-options')
var arrayify = require('array-back')

/**
@module command-line-usage
*/
module.exports = getUsage

class Lines extends Array {
  add (content) {
    arrayify(content).forEach(line => this.push(ansi.format(line)))
  }

  emptyLine () {
    this.push('')
  }
}

/**
@param {optionDefinition[]} - an array of [option definition](https://github.com/75lb/command-line-args#exp_module_definition--OptionDefinition) objects. In addition to the regular definition properties, command-line-usage will look for:

- `description` - a string describing the option.
- `typeLabel` - a string to replace the default type string (e.g. `<string>`). It's often more useful to set a more descriptive type label, like `<ms>`, `<files>`, `<command>` etc.

@param options {module:usage-options} - see [UsageOptions](#exp_module_usage-options--UsageOptions).
@returns {string}
@alias module:command-line-usage
*/
function getUsage (definitions, options) {
  options = new UsageOptions(options)
  definitions = definitions || []

  var output = new Lines()
  output.emptyLine()

  /* filter out hidden definitions */
  if (options.hide && options.hide.length) {
    definitions = definitions.filter(definition => options.hide.indexOf(definition.name) === -1)
  }

  if (options.header) {
    output.add(renderSection('', options.header))
  }

  if (options.title || options.description) {
    output.add(renderSection(options.title, options.description))
  }

  if (options.synopsis) {
    output.add(renderSection('Synopsis', options.synopsis))
  }

  if (definitions.length) {
    if (options.groups) {
      o.each(options.groups, (val, group) => {
        var title
        var description
        if (t.isObject(val)) {
          title = val.title
          description = val.description
        } else if (t.isString(val)) {
          title = val
        } else {
          throw new Error('Unexpected group config structure')
        }

        output.add(renderSection(title, description))

        let optionList = getUsage.optionList(definitions, group)
        output.add(renderSection(null, optionList, true))
      })
    } else {
      output.add(renderSection('Options', getUsage.optionList(definitions), true))
    }
  }

  if (options.examples) {
    output.add(renderSection('Examples', options.examples))
  }

  if (options.footer) {
    output.add(renderSection('', options.footer))
  }

  return output.join(os.EOL)
}

function getOptionNames (definition, optionNameStyles) {
  var names = []
  var type = definition.type ? definition.type.name.toLowerCase() : ''
  var multiple = definition.multiple ? '[]' : ''
  if (type) type = type === 'boolean' ? '' : `[underline]{${type}${multiple}}`
  type = ansi.format(definition.typeLabel || type)

  if (definition.alias) names.push(ansi.format('-' + definition.alias, optionNameStyles))
  names.push(ansi.format(`--${definition.name}`, optionNameStyles) + ' ' + type)
  return names.join(', ')
}

function renderSection (title, content, skipIndent) {
  var lines = new Lines()

  if (title) {
    lines.add(ansi.format(title, [ 'underline', 'bold' ]))
    lines.emptyLine()
  }

  if (!content) {
    return lines
  } else {
    if (t.isString(content)) {
      lines.add(indentString(content))
    } else if (Array.isArray(content) && content.every(t.isString)) {
      lines.add(skipIndent ? content : indentArray(content))
    } else if (Array.isArray(content) && content.every(t.isPlainObject)) {
      lines.add(columnLayout.lines(content, {
        padding: { left: '  ', right: ' ' }
      }))
    } else if (t.isPlainObject(content)) {
      if (!content.options || !content.data) {
        throw new Error('must have an "options" or "data" property\n' + JSON.stringify(content))
      }
      content.options = o.extend({
        padding: { left: '  ', right: ' ' }
      }, content.options)
      lines.add(columnLayout.lines(
        content.data.map(row => formatRow(row)),
        content.options
      ))
    } else {
      var message = `invalid input - 'content' must be a string, array of strings, or array of plain objects:\n\n${JSON.stringify(content)}`
      throw new Error(message)
    }

    lines.emptyLine()
    return lines
  }
}

function indentString (string) {
  return '  ' + string
}
function indentArray (array) {
  return array.map(indentString)
}
function formatRow (row) {
  o.each(row, (val, key) => {
    row[key] = ansi.format(val)
  })
  return row
}

/**
 * A helper for getting a column-format list of options and descriptions. Useful for inserting into a custom usage template.
 *
 * @param {optionDefinition[]} - the definitions to Display
 * @param [group] {string} - if specified, will output the options in this group. The special group `'_none'` will return options without a group specified.
 * @returns {string[]}
 */
getUsage.optionList = function (definitions, group) {
  if (!definitions || (definitions && !definitions.length)) {
    throw new Error('you must pass option definitions to getUsage.optionList()')
  }
  var columns = []

  if (group === '_none') {
    definitions = definitions.filter(def => !t.isDefined(def.group))
  } else if (group) {
    definitions = definitions.filter(def => arrayify(def.group).indexOf(group) > -1)
  }

  definitions
    .forEach(def => {
      columns.push({
        option: getOptionNames(def, 'bold'),
        description: ansi.format(def.description)
      })
    })

  return columnLayout.lines(columns, {
    padding: { left: '  ', right: ' ' },
    columns: [
      { name: 'option', nowrap: true },
      { name: 'description', maxWidth: 80 }
    ]
  })
}
