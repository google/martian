'use strict';

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

var _get = function get(_x, _x2, _x3) { var _again = true; _function: while (_again) { var object = _x, property = _x2, receiver = _x3; _again = false; if (object === null) object = Function.prototype; var desc = Object.getOwnPropertyDescriptor(object, property); if (desc === undefined) { var parent = Object.getPrototypeOf(object); if (parent === null) { return undefined; } else { _x = parent; _x2 = property; _x3 = receiver; _again = true; desc = parent = undefined; continue _function; } } else if ('value' in desc) { return desc.value; } else { var getter = desc.get; if (getter === undefined) { return undefined; } return getter.call(receiver); } } };

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

function _inherits(subClass, superClass) { if (typeof superClass !== 'function' && superClass !== null) { throw new TypeError('Super expression must either be null or a function, not ' + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var Columns = require('./columns');
var ansi = require('./ansi');
var arrayify = require('array-back');
var wrap = require('wordwrapjs');
var Cell = require('./cell');

var Rows = (function (_Array) {
  _inherits(Rows, _Array);

  function Rows(rows, columns) {
    _classCallCheck(this, Rows);

    _get(Object.getPrototypeOf(Rows.prototype), 'constructor', this).call(this);
    this.load(rows, columns);
  }

  _createClass(Rows, [{
    key: 'load',
    value: function load(rows, columns) {
      var _this = this;

      arrayify(rows).forEach(function (row) {
        return _this.push(new Map(objectToIterable(row, columns)));
      });
    }
  }], [{
    key: 'getColumns',
    value: function getColumns(rows) {
      var columns = new Columns();
      arrayify(rows).forEach(function (row) {
        for (var columnName in row) {
          var column = columns.get(columnName);
          if (!column) {
            column = columns.add({ name: columnName, contentWidth: 0, minContentWidth: 0 });
          }
          var cell = new Cell(row[columnName], column);
          var cellValue = cell.value;
          if (ansi.has(cellValue)) {
            cellValue = ansi.remove(cellValue);
          }

          if (cellValue.length > column.contentWidth) column.contentWidth = cellValue.length;

          var longestWord = getLongestWord(cellValue);
          if (longestWord > column.minContentWidth) {
            column.minContentWidth = longestWord;
          }
          if (!column.contentWrappable) column.contentWrappable = wrap.isWrappable(cellValue);
        }
      });
      return columns;
    }
  }]);

  return Rows;
})(Array);

function getLongestWord(line) {
  var words = wrap.getWords(line);
  return words.reduce(function (max, word) {
    return Math.max(word.length, max);
  }, 0);
}

function objectToIterable(row, columns) {
  return columns.map(function (column) {
    return [column, new Cell(row[column.name], column)];
  });
}

module.exports = Rows;