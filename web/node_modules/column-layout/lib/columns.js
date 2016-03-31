'use strict'
var t = require('typical')
var Padding = require('./padding')
var arrayify = require('array-back')

var _viewWidth = new WeakMap()

class Columns extends Array {
  constructor (columns) {
    super()
    this.load(columns)
  }

  /**
   * sum of all generatedWidth fields
   * @return {number}
   */
  totalWidth () {
    return this.length
      ? this.map(col => col.generatedWidth).reduce((a, b) => a + b)
      : 0
  }

  totalFixedWidth () {
    return this.getFixed()
      .map(col => col.generatedWidth)
      .reduce((a, b) => a + b, 0)
  }

  get (columnName) {
    return this.find(column => column.name === columnName)
  }

  getResizable () {
    return this.filter(column => column.isResizable())
  }

  getFixed () {
    return this.filter(column => column.isFixed())
  }

  load (columns) {
    arrayify(columns).forEach(this.add.bind(this))
  }

  add (column) {
    var col = column instanceof Column ? column : new Column(column)
    this.push(col)
    return col
  }

  set viewWidth (val) {
    _viewWidth.set(this, val)
  }

  /**
   * sets `generatedWidth` for each column
   * @chainable
   */
  autoSize () {
    var viewWidth = _viewWidth.get(this)

    /* size */
    this.forEach(column => {
      column.generateWidth()
      column.generateMinWidth()
    })

    /* adjust if user set a min or maxWidth */
    this.forEach(column => {
      if (t.isDefined(column.maxWidth) && column.generatedWidth > column.maxWidth) {
        column.generatedWidth = column.maxWidth
      }

      if (t.isDefined(column.minWidth) && column.generatedWidth < column.minWidth) {
        column.generatedWidth = column.minWidth
      }
    })

    var width = {
      total: this.totalWidth(),
      view: viewWidth,
      diff: this.totalWidth() - viewWidth,
      totalFixed: this.totalFixedWidth(),
      totalResizable: Math.max(viewWidth - this.totalFixedWidth(), 0)
    }

    /* adjust if short of space */
    if (width.diff > 0) {
      /* share the available space between resizeable columns */
      let resizableColumns = this.getResizable()
      resizableColumns.forEach(column => {
        column.generatedWidth = Math.floor(width.totalResizable / resizableColumns.length)
      })

      /* at this point, the generatedWidth should never end up bigger than the contentWidth */
      var grownColumns = this.filter(column => column.generatedWidth > column.contentWidth)
      var shrunkenColumns = this.filter(column => column.generatedWidth < column.contentWidth)
      var salvagedSpace = 0
      grownColumns.forEach(column => {
        var currentGeneratedWidth = column.generatedWidth
        column.generateWidth()
        salvagedSpace += currentGeneratedWidth - column.generatedWidth
      })
      shrunkenColumns.forEach(column => {
        column.generatedWidth += Math.floor(salvagedSpace / shrunkenColumns.length)
      })

      /* if, after autosizing, we still don't fit within viewWidth then give up */
    }

    return this
  }
}

var _padding = new WeakMap()

// setting any column property which is a factor of the width should trigger autoSize()
/**
 * @class
 * @classdesc Represents a table column
 */
class Column {
  constructor (column) {

    /**
     * @type {string}
     */
    if (t.isDefined(column.name)) this.name = column.name
    /**
     * @type {number}
     */
    if (t.isDefined(column.width)) this.width = column.width
    if (t.isDefined(column.maxWidth)) this.maxWidth = column.maxWidth
    if (t.isDefined(column.minWidth)) this.minWidth = column.minWidth
    if (t.isDefined(column.nowrap)) this.nowrap = column.nowrap
    if (t.isDefined(column.break)) this.break = column.break
    if (t.isDefined(column.contentWrappable)) this.contentWrappable = column.contentWrappable
    if (t.isDefined(column.contentWidth)) this.contentWidth = column.contentWidth
    if (t.isDefined(column.minContentWidth)) this.minContentWidth = column.minContentWidth
    this.padding = column.padding || { left: ' ', right: ' ' }
    this.generatedWidth = null
  }

  set padding (padding) {
    _padding.set(this, new Padding(padding))

  }
  get padding () {
    return _padding.get(this)
  }

  get wrappedContentWidth () {
    return Math.max(this.generatedWidth - this.padding.length(), 0)
  }

  isResizable () {
    return !this.isFixed()
  }

  isFixed () {
    return t.isDefined(this.width) || this.nowrap || !this.contentWrappable
  }

  generateWidth () {
    this.generatedWidth = this.width || (this.contentWidth + this.padding.length())
  }

  generateMinWidth () {
    this.minWidth = this.minContentWidth + this.padding.length()
  }
}

/**
 * @module columns
 */
module.exports = Columns
