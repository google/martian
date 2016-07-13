'use strict';

export class Node {
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

  static function(argc, infix) {
    var node = new Node('FUNCTION');

    node.argc = argc;
    node.infix = infix;

    return node;
  }

  constructor(type, value) {
    this.type = type;
    this.value = value;
  }

  print() {
    var string;

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
        return this.next && this.next.print() || '';
      case 'FUNCTION':
        var call = ['func'];

        call.push(this.argc != undefined ? this.argc : '?');

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

  function(id, argc, infix) {
    var node = this._next('FUNCTION', id);

    node.argc = argc;
    node.infix = infix;

    return node;
  }
}

export class Parser {
  parse(tokens) {
    var node = Node.list(); 

    for (var i = 0; i < tokens.length; i++) {
      var token = tokens[i];

      switch (token.type) {
        case 'WHITESPACE':
          continue;
        case 'LEFT_PAREN':
          node = node.list();
          break;
        case 'RIGHT_PAREN':
          node = node.end();
          break;
        case 'IDENTIFIER':
          node = node.identifier(token.value);
          break;
        case 'STRING':
          node = node.string(token.value);
          break;
        case 'INTEGER':
          node = node.integer(Number.parseInt(token.value));
          break;
        case 'FLOAT':
          node = node.float(Number.parseFloat(token.value));
          break;
        case 'BOOLEAN':
          node = node.boolean(token.value === 'true');
          break;
        default:
          throw 'Unknown token type: ' + token.type;
      }
    }

    return node.end(); 
  }
}
