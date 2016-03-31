'use strict';

export default class Node {
  constructor(type, value) {
    this._type = 'curio.Node'; // Type checking.

    this.type = type;
    this.value = value;
  }

  static fromValue(value, attrs) {
    switch (typeof(value)) {
      case 'boolean':
        return Node.boolean(value);
      case 'string':
        return Node.string(value);
      case 'number':
        if (n % 1 === 0) {
          return Node.integer(value);
        }
        return Node.float(value);
      case 'function':
        attrs.argc = attrs.argc;

        return Node.function(function() {
          var values = [];

          for (var i = 0; i < arguments.length; i++) {
            values.push(Node.fromValue(arguments[i]));
          }

          return Node.fromValue(func.apply(this, values));
        }, attrs);
      case 'object':
        if (value._type == 'curio.Node') {
          return value;
        }

        if (Array.isArray(value)) {
          var list = Node.list();
          var last = list;

          for (var i = 0; i < value.length; i++) {
            last = last.insert(Node.fromValue(value[i]));
          }

          return list.end();
        }
      default:
        throw `cannot convert Javascript value to Node: ${value}`;
    }
  }

  static list() {
    var node = new Node('LIST');

    node.down = Node.sentinel();
    node.down.up = node;

    return node.down;
  }

  static sentinel() {
    return new Node('SENTINEL');
  }

  static string(value) {
    return new Node('STRING', value);
  }

  static integer(value) {
    return new Node('INTEGER', value);
  }

  static float(value) {
    return new Node('FLOAT', value);
  }

  static boolean(value) {
    return new Node('BOOLEAN', value);
  }

  static identifier(id) {
    return new Node('IDENTIFIER', id);
  }

  static function(value, attrs) {
    var node = new Node('FUNCTION', value);

    node.argc = attrs && attrs.argc || value && value.length || 0;
    node.infix = attrs && attrs.infix || false;

    return node;
  }

  static operator(value) {
    var node = new Node('FUNCTION', value);

    node.argc = 2;
    node.infix = true;

    return node;
  }

  toValue() {
    switch (node.type) {
      case 'LIST':
        var values = [];

        for (node = node.down.next; node; node = node.next) {
          values.push(node.toValue());
        }

        return values;
      default:
        return node.value;
    }
  }

  print(includeSentinel) {
    var string = '';

    switch (this.type) {
      case 'LIST':
        string = this.down.print();

        if (this.up) {
          string = '(' + string + ')';
        }

        break;
      case 'STRING':
        string = '"' + this.value + '"';
        break;
      case 'SENTINEL':
        if (includeSentinel) {
          string = '*';
        }
        break;
      case 'FUNCTION':
        var call = ['function', `argc:${this.argc}`];

        if (this.infix) {
          call.push('infix');
        }

        string = '[' + call.join(' ') + ']';
        break;
      case 'IDENTIFIER':
      case 'INTEGER':
      case 'FLOAT':
      case 'BOOLEAN':
        string = this.value.toString();
    }

    if (this.next) {
      if (!string) {
        return this.next.print();
      }

      return string + ' ' + this.next.print();
    }

    return string;
  }

  list() {
    var node = this._next('LIST');

    node.down = new Node('SENTINEL');
    node.down.up = node;

    return node.down;
  }

  end() {
    return this.up;
  }

  insert(node) {
    this.next = node;
    node.prev = this;
    node.up = this.up;

    return node;
  }

  _next(type, value) {
    return this.insert(new Node(type, value));
  }

  identifier(id) {
    return this._next('IDENTIFIER', id);
  }

  string(value) {
    return this._next('STRING', value);
  }

  integer(value) {
    return this._next('INTEGER', value);
  }

  float(value) {
    return this._next('FLOAT', value);
  }

  boolean(value) {
    return this._next('BOOLEAN', value);
  }

  sentinel() {
    return this._next('SENTINEL');
  }

  function(value, attrs) {
    var node = this._next('FUNCTION', value);

    node.argc = attrs && attrs.argc || value && value.length || 0;
    node.infix = attrs && attrs.infix || false;

    return node;
  }

  operator(value) {
    var node = this._next('FUNCTION', value);

    node.argc = 2;
    node.infix = true;

    return node;
  }
}
