'use strict';

export class Scope {
  constructor(up) {
    this.symbols = {};
    this.up = up;
  }

  push() {
    return new Scope(this); 
  }

  pop() {
    return this.up;
  }

  put(symbol, node) {
    this.symbols[symbol] = node;
  }

  get(symbol) {
    var node = this.symbols[symbol];
    if (node) {
      return node;
    }

    if (this.up) {
      return this.up.get(symbol);
    }

    return null;
  }
}
