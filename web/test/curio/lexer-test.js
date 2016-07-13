'use strict'; 

var chai = require('chai'),
    assert = chai.assert;

var lex = require('../../scripts/curio/lexer'),
    Lexer = lex.Lexer,
    Token = lex.Token;

describe('Lexer', function() {
  it('lexes an empty string', function() {
    var lexer = new Lexer('');
    var ts = lexer.lex();

    assert.deepEqual(ts, []);
  });

  it('lexes an empty list', function() {
    var lexer = new Lexer('()');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('LEFT_PAREN', 0),
      new Token('RIGHT_PAREN', 1),
    ]);
  });

  it('lexes an integer', function() {
    var lexer = new Lexer('100');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('INTEGER', 0, '100'),
    ]);
  });

  it('lexes a float', function() {
    var lexer = new Lexer('10.5');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('FLOAT', 0, '10.5'),
    ]);
  });

  it('lexes a float with a leading decimal', function() {
    var lexer = new Lexer('.1');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('FLOAT', 0, '.1'),
    ]);
  });

  it('lexes an identifier that looks a bit like a float', function() {
    var lexer = new Lexer('.1.1');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('IDENTIFIER', 0, '.1.1'),
    ]);
  });

  it('lexes an identifier starting with a number', function() {
    var lexer = new Lexer('2015-11-30');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('IDENTIFIER', 0, '2015-11-30'),
    ]);
  });

  it('lexes an empty string', function() {
    var lexer = new Lexer('""');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('STRING', 0, ''),
    ]);
  });

  it('lexes a string', function() {
    var lexer = new Lexer('"string"');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('STRING', 0, 'string'),
    ]);
  });

  it('lexes a string with escaped quotes', function() {
    var lexer = new Lexer('"string with \\"quotes\\""');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('STRING', 0, 'string with "quotes"'),
    ]);
  });

  it('lexes a bare identifier', function() {
    var lexer = new Lexer('identifier');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('IDENTIFIER', 0, 'identifier'),
    ]);
  });

  it('lexes a bare identifier that looks like a number at first', function() {
    var lexer = new Lexer('0.2_OH_WAIT_SURPRISE_IM_AN_IDENTIFIER');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('IDENTIFIER', 0, '0.2_OH_WAIT_SURPRISE_IM_AN_IDENTIFIER'),
    ]);
  });

  it('lexes a boolean', function() {
    var lexer = new Lexer('true false');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
      new Token('BOOLEAN', 0, 'true'),
      new Token('WHITESPACE', 4),
      new Token('BOOLEAN', 5, 'false'),
    ]);
  });

  it('lexes a series of tokens', function() {
    var lexer = new Lexer('(status >= 400) and method is "GET"');
    var ts = lexer.lex();

    assert.deepEqual(ts, [
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
    ]);
  });
});
