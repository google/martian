'use strict';

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

var _get = function get(_x, _x2, _x3) { var _again = true; _function: while (_again) { var object = _x, property = _x2, receiver = _x3; _again = false; if (object === null) object = Function.prototype; var desc = Object.getOwnPropertyDescriptor(object, property); if (desc === undefined) { var parent = Object.getPrototypeOf(object); if (parent === null) { return undefined; } else { _x = parent; _x2 = property; _x3 = receiver; _again = true; desc = parent = undefined; continue _function; } } else if ('value' in desc) { return desc.value; } else { var getter = desc.get; if (getter === undefined) { return undefined; } return getter.call(receiver); } } };

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

function _inherits(subClass, superClass) { if (typeof superClass !== 'function' && superClass !== null) { throw new TypeError('Super expression must either be null or a function, not ' + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var columnLayout = require('column-layout');
var o = require('object-tools');
var ansi = require('ansi-escape-sequences');
var os = require('os');
var t = require('typical');
var UsageOptions = require('./usage-options');
var arrayify = require('array-back');

module.exports = getUsage;

var Lines = (function (_Array) {
  _inherits(Lines, _Array);

  function Lines() {
    _classCallCheck(this, Lines);

    _get(Object.getPrototypeOf(Lines.prototype), 'constructor', this).apply(this, arguments);
  }

  _createClass(Lines, [{
    key: 'add',
    value: function add(content) {
      var _this = this;

      arrayify(content).forEach(function (line) {
        return _this.push(ansi.format(line));
      });
    }
  }, {
    key: 'emptyLine',
    value: function emptyLine() {
      this.push('');
    }
  }]);

  return Lines;
})(Array);

function getUsage(definitions, options) {
  options = new UsageOptions(options);
  definitions = definitions || [];

  var output = new Lines();
  output.emptyLine();

  if (options.hide && options.hide.length) {
    definitions = definitions.filter(function (definition) {
      return options.hide.indexOf(definition.name) === -1;
    });
  }

  if (options.header) {
    output.add(renderSection('', options.header));
  }

  if (options.title || options.description) {
    output.add(renderSection(options.title, options.description));
  }

  if (options.synopsis) {
    output.add(renderSection('Synopsis', options.synopsis));
  }

  if (definitions.length) {
    if (options.groups) {
      o.each(options.groups, function (val, group) {
        var title;
        var description;
        if (t.isObject(val)) {
          title = val.title;
          description = val.description;
        } else if (t.isString(val)) {
          title = val;
        } else {
          throw new Error('Unexpected group config structure');
        }

        output.add(renderSection(title, description));

        var optionList = getUsage.optionList(definitions, group);
        output.add(renderSection(null, optionList, true));
      });
    } else {
      output.add(renderSection('Options', getUsage.optionList(definitions), true));
    }
  }

  if (options.examples) {
    output.add(renderSection('Examples', options.examples));
  }

  if (options.footer) {
    output.add(renderSection('', options.footer));
  }

  return output.join(os.EOL);
}

function getOptionNames(definition, optionNameStyles) {
  var names = [];
  var type = definition.type ? definition.type.name.toLowerCase() : '';
  var multiple = definition.multiple ? '[]' : '';
  if (type) type = type === 'boolean' ? '' : '[underline]{' + type + multiple + '}';
  type = ansi.format(definition.typeLabel || type);

  if (definition.alias) names.push(ansi.format('-' + definition.alias, optionNameStyles));
  names.push(ansi.format('--' + definition.name, optionNameStyles) + ' ' + type);
  return names.join(', ');
}

function renderSection(title, content, skipIndent) {
  var lines = new Lines();

  if (title) {
    lines.add(ansi.format(title, ['underline', 'bold']));
    lines.emptyLine();
  }

  if (!content) {
    return lines;
  } else {
    if (t.isString(content)) {
      lines.add(indentString(content));
    } else if (Array.isArray(content) && content.every(t.isString)) {
      lines.add(skipIndent ? content : indentArray(content));
    } else if (Array.isArray(content) && content.every(t.isPlainObject)) {
      lines.add(columnLayout.lines(content, {
        padding: { left: '  ', right: ' ' }
      }));
    } else if (t.isPlainObject(content)) {
      if (!content.options || !content.data) {
        throw new Error('must have an "options" or "data" property\n' + JSON.stringify(content));
      }
      content.options = o.extend({
        padding: { left: '  ', right: ' ' }
      }, content.options);
      lines.add(columnLayout.lines(content.data.map(function (row) {
        return formatRow(row);
      }), content.options));
    } else {
      var message = 'invalid input - \'content\' must be a string, array of strings, or array of plain objects:\n\n' + JSON.stringify(content);
      throw new Error(message);
    }

    lines.emptyLine();
    return lines;
  }
}

function indentString(string) {
  return '  ' + string;
}
function indentArray(array) {
  return array.map(indentString);
}
function formatRow(row) {
  o.each(row, function (val, key) {
    row[key] = ansi.format(val);
  });
  return row;
}

getUsage.optionList = function (definitions, group) {
  if (!definitions || definitions && !definitions.length) {
    throw new Error('you must pass option definitions to getUsage.optionList()');
  }
  var columns = [];

  if (group === '_none') {
    definitions = definitions.filter(function (def) {
      return !t.isDefined(def.group);
    });
  } else if (group) {
    definitions = definitions.filter(function (def) {
      return arrayify(def.group).indexOf(group) > -1;
    });
  }

  definitions.forEach(function (def) {
    columns.push({
      option: getOptionNames(def, 'bold'),
      description: def.description
    });
  });

  return columnLayout.lines(columns, {
    padding: { left: '  ', right: ' ' },
    columns: [{ name: 'option', nowrap: true }, { name: 'description', maxWidth: 80 }]
  });
};