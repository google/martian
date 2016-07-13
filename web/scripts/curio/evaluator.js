'use strict';

var parse = require('./parser'),
    Node = parse.Node;

export class Evaluator {
  constructor(message, scope) {
    this.message = message;
    this.scope = scope;
  }

  evaluate(node) {
    var rnode;
    var error;

    this._evaluate(node,
        function(node) { rnode = node; },
        function(node, message) {
          error = { node: node, message: message };
        });

    if (error) {
      throw error;
    }

    return rnode;
  }

  _evaluate(node, ret, raise) {
    switch (node.type) {
      case 'LIST':
        var list = Node.list();

        this._evaluate(node.down, function(subnode) {
          list.insert(subnode);
        }, raise);

        ret(list.end());

        break;
      case 'IDENTIFIER':
        var scoped = this.scope.get(node.value);
        if (!scoped) {
          raise(node, 'unbound identifier');
          return;
        }

        if (scoped.type != 'FUNCTION') {
          this._evaluate(scoped, ret, raise);
          return;
        }

        var args = []; 
        this._evaluate(node.next, function(arg) {
          args.unshift(arg);
        }, raise);

        this._evaluate(scoped.value.apply(this.message, args), ret, raise);
        break;
      case 'SENTINEL':
        if (node.next) {
          this._evaluate(node.next, ret, raise);
        }
        break;
      default:
        if (node.next) {
          this._evaluate(node.next, ret, raise);
        }

        ret(node);
    }
  }
}
