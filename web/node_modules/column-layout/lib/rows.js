'use strict'
const Columns = require('./columns')
const ansi = require('./ansi')
const arrayify = require('array-back')
const wrap = require('wordwrapjs')
const Cell = require('./cell')

/**
 * @class Rows
 * @extends Array
 */
class Rows extends Array {
  constructor (rows, columns) {
    super()
    this.load(rows, columns)
  }

  load (rows, columns) {
    arrayify(rows).forEach(row => this.push(new Map(objectToIterable(row, columns))))
  }

  /**
   * returns all distinct columns from input
   * @param  {object[]}
   * @return {module:columns}
   */
  static getColumns (rows) {
    var columns = new Columns()
    arrayify(rows).forEach(row => {
      for (let columnName in row) {
        let column = columns.get(columnName)
        if (!column) {
          column = columns.add({ name: columnName, contentWidth: 0, minContentWidth: 0 })
        }
        let cell = new Cell(row[columnName], column)
        let cellValue = cell.value
        if (ansi.has(cellValue)) {
          cellValue = ansi.remove(cellValue)
        }

        if (cellValue.length > column.contentWidth) column.contentWidth = cellValue.length

        let longestWord = getLongestWord(cellValue)
        if (longestWord > column.minContentWidth) {
          column.minContentWidth = longestWord
        }
        if (!column.contentWrappable) column.contentWrappable = wrap.isWrappable(cellValue)
      }
    })
    return columns
  }
}

function getLongestWord (line) {
  const words = wrap.getWords(line)
  return words.reduce((max, word) => {
    return Math.max(word.length, max)
  }, 0)
}

function objectToIterable (row, columns) {
  return columns.map(column => {
    return [ column, new Cell(row[column.name], column) ]
  })
}

/**
 * @module rows
 */
module.exports = Rows
