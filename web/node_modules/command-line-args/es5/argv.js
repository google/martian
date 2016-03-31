'use strict';

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

var _get = function get(_x, _x2, _x3) { var _again = true; _function: while (_again) { var object = _x, property = _x2, receiver = _x3; _again = false; if (object === null) object = Function.prototype; var desc = Object.getOwnPropertyDescriptor(object, property); if (desc === undefined) { var parent = Object.getPrototypeOf(object); if (parent === null) { return undefined; } else { _x = parent; _x2 = property; _x3 = receiver; _again = true; desc = parent = undefined; continue _function; } } else if ('value' in desc) { return desc.value; } else { var getter = desc.get; if (getter === undefined) { return undefined; } return getter.call(receiver); } } };

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

function _inherits(subClass, superClass) { if (typeof superClass !== 'function' && superClass !== null) { throw new TypeError('Super expression must either be null or a function, not ' + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var arrayify = require('array-back');
var option = require('./option');
var findReplace = require('find-replace');

var Argv = (function (_Array) {
  _inherits(Argv, _Array);

  function Argv(argv) {
    _classCallCheck(this, Argv);

    _get(Object.getPrototypeOf(Argv.prototype), 'constructor', this).call(this);

    if (argv) {
      argv = arrayify(argv);
    } else {
      argv = process.argv.slice(0);
      argv.splice(0, 2);
    }

    this.load(argv);
  }

  _createClass(Argv, [{
    key: 'load',
    value: function load(array) {
      var _this = this;

      arrayify(array).forEach(function (item) {
        return _this.push(item);
      });
    }
  }, {
    key: 'clear',
    value: function clear() {
      this.length = 0;
    }
  }, {
    key: 'expandOptionEqualsNotation',
    value: function expandOptionEqualsNotation() {
      var optEquals = option.optEquals;
      if (this.some(optEquals.test.bind(optEquals))) {
        var expandedArgs = [];
        this.forEach(function (arg) {
          var matches = arg.match(optEquals.re);
          if (matches) {
            expandedArgs.push(matches[1], matches[2]);
          } else {
            expandedArgs.push(arg);
          }
        });
        this.clear();
        this.load(expandedArgs);
      }
    }
  }, {
    key: 'expandGetoptNotation',
    value: function expandGetoptNotation() {
      var combinedArg = option.combined;
      var hasGetopt = this.some(combinedArg.test.bind(combinedArg));
      if (hasGetopt) {
        findReplace(this, combinedArg.re, function (arg) {
          arg = arg.slice(1);
          return arg.split('').map(function (letter) {
            return '-' + letter;
          });
        });
      }
    }
  }, {
    key: 'validate',
    value: function validate(definitions) {
      var invalidOption;

      var optionWithoutDefinition = this.filter(function (arg) {
        return option.isOption(arg);
      }).some(function (arg) {
        if (definitions.get(arg) === undefined) {
          invalidOption = arg;
          return true;
        }
      });
      if (optionWithoutDefinition) {
        halt('UNKNOWN_OPTION', 'Unknown option: ' + invalidOption);
      }
    }
  }]);

  return Argv;
})(Array);

function halt(name, message) {
  var err = new Error(message);
  err.name = name;
  throw err;
}

module.exports = Argv;