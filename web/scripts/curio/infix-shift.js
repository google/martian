'use strict';

export default class InfixShift {
  constructor(scope) {
    this._scope = scope;
  }

  transform(node, precedence) {
    if (!node) {
      return;
    }

    switch (node.type) {
      case 'LIST':
        this.transform(node.down, precedence);
        break;
      case 'IDENTIFIER':
        var scoped = this._scope.get(node.value, precedence);
        if (!scoped || !scoped.infix) {
          break;
        }

        var prev = node.prev;
        var next = node.next;

        prev.prev.next = node;
        node.prev = prev.prev;

        node.next = prev;

        prev.prev = node;
        prev.next = next;

        next.prev = prev;

        node = prev;
    }

    this.transform(node.next, precedence);
  }
}
