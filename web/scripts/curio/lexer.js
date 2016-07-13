'use strict';

export class Token {
  constructor(type, pos, value) {
    this.type = type;
    this.pos = pos;
    this.value = value;
  }
}

export class Lexer {
  constructor(input) {
    this.input = input;
    this.pos = 0;
  }

  lex() {
    var tokens = [];

    while (this.pos < this.input.length) {
      var ch = this.input[this.pos];
      var start = this.pos;

      switch (ch) {
        case '(':
          tokens.push(new Token('LEFT_PAREN', this.pos++));
          break;
        case  ')':
          tokens.push(new Token('RIGHT_PAREN', this.pos++));
          break;
        case ' ':
        case '\t':
          tokens.push(new Token('WHITESPACE', this.pos));
          this._skipWhitespace();
          break;
        case '"':
          var escaped = false;

          var value = this._consume(function(ch) {
            switch (ch) {
              case '\\':
                escaped = true;
                return false;
              case '"':
                if (escaped) {
                  escaped = false;
                  return true;
                }

                return false;
            }

            return true;
          });

          tokens.push(new Token('STRING', start, value));
          break;
        default:
          if ( ch == '.' || (ch >= '0' && ch <= '9')) {
            var type = 'INTEGER';

            var value = this._consume(function(ch) {
              if (ch == '.') {
                if (type == 'INTEGER') {
                  type = 'FLOAT';
                } else if (type == 'FLOAT') {
                  type = 'IDENTIFIER';
                }
              } else if (ch < '0' || ch > '9') {
                type = 'IDENTIFIER';
              }
            });

            tokens.push(new Token(type, start, value));
            continue;
          }

          var type = 'IDENTIFIER';
          var value = this._consume();

          switch (value.toLowerCase()) {
            case 'true':
            case 'false':
              type = 'BOOLEAN';
          }

          tokens.push(new Token(type, start, value));
      }
    }

    return tokens;
  }

  _skipWhitespace() {
    while (this.pos < this.input.length) {
      switch (this.input[this.pos++]) {
        case ' ':
        case '\t':
        default:
          return;
      }
    }
  }

  _consume(func) {
    var value = '';

    while (this.pos < this.input.length) {
      var ch = this.input[this.pos++];

      var result;
      if (func) {
        result = func(ch);
      }

      // Possible return values for callback:
      //
      // undefined: continue as normal stopping at boundary tokens (whitespace,
      //            parens, etc.)
      // false:     skip the current character and do not append it to the
      //            value.
      // true:      force the current character to be appended to the value.
      switch (result) {
        case true:
          value += ch; 
          continue;
        case false:
          continue;
        case undefined:
          value += ch; 
      }

      switch (this.input[this.pos]) {
        case ')':
        case ' ':
        case '\t':
          return value;
      }
    }

    return value;
  }
}
