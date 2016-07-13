'use strict';

var chai = require('chai'),
    assert = chai.assert;

var parse = require('../../scripts/curio/parser'),
    lex = require('../../scripts/curio/lexer'),
    Parser = parse.Parser,
    Node = parse.Node,
    Token = lex.Token;

describe('Parser', function() {
  it('parses an empty list', function() {
    var parser = new Parser();
    var tokens = [];

    var root = parser.parse(tokens);

    assert.equal(root.print(), '');

    assert.deepEqual(root, Node.list().end());
  });

  it('parses an identifier', function() {
    var parser = new Parser();
    var tokens = [
      new Token('IDENTIFIER', 0, 'name'),
    ];

    var root = parser.parse(tokens);

    assert.equal(root.print(), 'name');

    assert.deepEqual(root, Node.list().identifier('name').end());
  });

  it('parses a string', function() {
    var parser = new Parser();
    var tokens = [
      new Token('STRING', 0, 'Curio'),
    ];

    var root = parser.parse(tokens);

    assert.equal(root.print(), '"Curio"');

    assert.deepEqual(root, Node.list().string('Curio').end());
  });

  it('parses an integer', function() {
    var parser = new Parser();
    var tokens = [
      new Token('INTEGER', 0, 4),
    ];

    var root = parser.parse(tokens);

    assert.equal(root.print(), '4');

    assert.deepEqual(root, Node.list().integer(4).end());
  });

  it('parses a float', function() {
    var parser = new Parser();
    var tokens = [
      new Token('FLOAT', 0, 10.5),
    ];

    var root = parser.parse(tokens);

    assert.equal(root.print(), '10.5');

    assert.deepEqual(root, Node.list().float(10.5).end());
  });

  it('parses a boolean', function() {
    var parser = new Parser();
    var tokens = [
      new Token('BOOLEAN', 0, 'false'),
    ];

    var root = parser.parse(tokens);

    assert.equal(root.print(), 'false');

    assert.deepEqual(root, Node.list().boolean(false).end());
  });

  it('parses a list of tokens', function() {
    var parser = new Parser();
    var tokens = [
      new Token('LEFT_PAREN', 0),
      new Token('IDENTIFIER', 1, 'status'),
      new Token('WHITESPACE', 7),
      new Token('IDENTIFIER', 8, '>='),
      new Token('WHITESPACE', 10),
      new Token('INTEGER', 11, '400'),
      new Token('RIGHT_PAREN', 14),
      new Token('WHITESPACE', 15),
      new Token('IDENTIFIER', 16, 'and'),
      new Token('WHITESPACE', 19),
      new Token('IDENTIFIER', 20, 'method'),
      new Token('WHITESPACE', 26),
      new Token('IDENTIFIER', 27, 'is'),
      new Token('WHITESPACE', 29),
      new Token('STRING', 30, 'GET'),
    ];

    var root = parser.parse(tokens);

    assert.equal(root.print(), '(status >= 400) and method is "GET"');

    assert.deepEqual(root, Node.list()
      .list()
        .identifier('status')
        .identifier('>=')
        .integer(400)
      .end()
      .identifier('and')
      .identifier('method')
      .identifier('is')
      .string('GET')
    .end());
  });

  it('prints a function node', function() {
    var node = Node.function(2, true);
    assert.equal(node.print(), '[func 2 infix]');

    node = Node.function();
    assert.equal(node.print(), '[func ?]');
  });
});
