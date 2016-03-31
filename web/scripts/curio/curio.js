import Parser from './parser';
import Node from './node';
import Scope from './scope';
import InfixShift from './infix-shift';
import ArgumentBinding from './argument-binding';
import Evaluator from './evaluator';
import {Lexer} from './lexer';

export default class Curio {
  constructor() {
    this._scope = new Scope(Curio._scope);
  }

  static function(id, func, attrs, precedence) {
    Curio._scope.put(id, Node.function(func, attrs), precedence);
  }

  static operator(id, func, precedence) {
    Curio._scope.put(id, Node.operator(func), precedence);
  }

  compile(input) {
    var lexer = new Lexer(input);
    var tokens = lexer.lex();

    var parser = new Parser();
    this._node = parser.parse(tokens);

    var infs = new InfixShift(this._scope);
    var argb = new ArgumentBinding(this._scope);

    this._scope.precedences().forEach(function(precedence) {
      infs.transform(this._node, precedence);
      argb.transform(this._node, precedence);
    }.bind(this));
  }

  run(message) {
    var evaluator = new Evaluator(message, this._scope);

    return evaluator.evaluate(this._node).value;
  }
};

Curio._scope = new Scope();

Curio.operator('is', function(a, b) { return a == b; }, Scope.precedence.COMPARISON);
Curio.operator('and', function(a, b) { return a && b; }, Scope.precedence.LOGICAL);
