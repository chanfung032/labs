<?php
/*
 * 实现一个类似于gmail filter的过滤器
 *
 * gmail filter的语法见：https://support.google.com/mail/answer/7190
 */

define('openParen',     0);
define('closeParen',    1);
define('exclude',       2);
define('or_',           3);
define('and_',          4);
define('whitespace',    5);
define('operator',      6);
define('word',          7);
define('quotedWord',    8);

function last($stack, $n=0) {
    $size = count($stack);
    return ($size - $n) ? $stack[$size - $n - 1] : null;
}

class SyntaxError extends Exception {
    public function __construct($col, $message) {
        parent::__construct("col: $col, $message");
    }
}

class Filter {
    public function __construct($str, $default_operator) {
        $this->str = $str;
        if ($str == '*') return;
        $this->default_operator = $default_operator;
        $this->rule = $this->parse($str);
    }

    public function __toString() {
        return $this->str;
    }

    public function match($dict) {
        if ($this->str == '*') return True;
        return $this->_match($dict, $this->rule);
    }

    public function _match($dict, $rule) {
        if (is_string($rule)) {
            if (is_string($dict)) {
                return strstr(@$dict, $rule) !== False;
            } else {
                $op = $this->default_operator;
                return strstr(@$dict[$op], $rule) !== False;
            }
        }

        if (is_array($rule)) {
            if (isset($rule[or_])) {
                $v = False;
                foreach ($rule[or_] as $r) {
                    $v = $this->_match($dict, $r);
                    if ($v) break;
                }
                return $v;
            } else if (isset($rule[and_])) {
                foreach ($rule[and_] as $r) {
                    if (!$this->_match($dict, $r)) {
                        return False;
                    }
                }
                return True;
            } else if (isset($rule[exclude])) {
                return !$this->_match($dict, $rule[exclude]);
            }

            assert(count($rule) == 1);

            $operator = key($rule);
            return $this->_match(@$dict[$operator], $rule[$operator]);
        }

        assert(0, "should not be here :(");
    }

    private function parse($string) {
        $tokens = $this->tokenize($string);

        $stack = Array();
        $valStack = Array();
        $valStackTopLevel = Array();
        $expectedExpr = False;

        foreach ($tokens as $token) {
            list($type, $col, $value) = $token;

            $canAnd = False;

            if ($type == openParen) {
                array_push($stack, $token);
                array_push($valStackTopLevel, $valStack);
                $valStack = Array();
            } else if ($type == closeParen) {
                while (last($stack) && current(last($stack)) !== openParen) {
                    $token = array_pop($stack);
                    array_push($valStack, $this->processOp($token, $valStack));
                }

                if (!last($stack) || current(last($stack)) !== openParen) {
                    throw new SyntaxError($col, 'no matching open paren');
                }

                // get the open paren out
                array_pop($stack);
                // the current valStack should now have only one element, otherwise
                // something went wrong
                if (count($valStack) !== 1) {
                    throw new SyntaxError($col, 'not enough operators in this sub expr');
                }

                $value = last($valStack);
                //$valStack = last($valStackTopLevel);
                if (!$valStack) {
                    throw new SyntaxError(
                        $col, 'not enought stacks for exprs?'
                    );
                }
                $valStack = array_pop($valStackTopLevel);
                array_push($valStack, $value);
                $canAnd = true;
            } else if ($type == whitespace) {
                if ($expectedExpr) {
                    throw new SyntaxError(
						$col,
                        "argument should be followed immediately for operator: $value"
                    );
                }
            } else if ($type === or_) {
                while (last($stack) && 
                        $this->priority(last($stack)) >= $this->priority($token)) {
                    $token = array_pop($stack);
                    array_push($valStack, $this->translate($token));
                }
                array_push($stack, $token);
            } else if ($type == word) {
                array_push($valStack, $value);
                $canAnd = True;
				$expectedExpr = False;
			} else if ($type == quotedWord) {
				array_push($valStack, substr($value, 1, strlen($value) - 2));
                $canAnd = True;
				$expectedExpr = False;
            } else {
                array_push($stack, $token);
                $expectedExpr = $token;
            }

            if ($canAnd) {
                // insert AND between two rules
                $token = Array(and_, $col - 1, '');
                while (last($stack) &&
                    $this->priority(last($stack)) >= $this->priority($token)) {
                        $token1 = array_pop($stack);
                        array_push($valStack, $this->processOp($token1, $valStack));
                }
                // check if two items on top of the stack are rules
                if (last($valStack) && last($valStack, 1)) {
                    array_push($stack, $token);
                }
            }
        }

        while (count($stack)) {
            $token = array_pop($stack);
            if ($token[0] === openParen) {
                throw new SyntaxError(
                    $token[1], 'no matching close paren'
                );
            }
            array_push($valStack, $this->processOp($token, $valStack));
        }

        if (count($valStack) !== 1) {
            throw new SyntaxError(0, 'unknown :(');
        }

        return $valStack[0];
    }

    private function priority($token) {
        $type = $token[0];
        if ($type === openParen)
            return -1;
        else if ($type === and_)
            return 1;
        else if ($type === or_)
            return 2;
        else
            return 3;
    }

    private function processOp($token, &$valStack) {
        list($type, $col, $value) = $token;
        if ($type == or_ || $type == and_) {
            $r = array_pop($valStack);
            $l = array_pop($valStack);
            if (!$r || !$l) {
                throw new SyntaxError(
                    $col,
                    "missing required arguments for operator $value"
                );
            }

            if ($type === or_) {
                if (isset($l[or_]) && is_array($l[or_])) {
                    $l[or_][] = $r;
                    return $l;
                }
                if (isset($r[or_]) && is_array($r[or_])) {
                    array_unshift($r[or_], $l);
                    return $r;
                }
                return Array(or_ => Array($l, $r));
            } else {
                if (isset($l[and_]) && is_array($l[and_])) {
                    $l[and_][] = $r;
                    return $l;
                }
                if (isset($r[and_]) && is_array($r[and_])) {
                    array_unshift($r[and_], $l);
                    return $r;
                }
                return Array(and_=> Array($l, $r));
            }
        } else {
            $r = array_pop($valStack);
            if (!$r) {
                throw new SyntaxError(
                    $col,
                    "missing the required argument for operator $value"
                );
            }
            if ($type != exclude) {
                $type = substr($value, 0, strlen($value)-1);
            }
            return Array($type => $r);
        }
    }

    private function tokenize($string) {
        $tokens = Array();
        $col = 0;
        while ($col < strlen($string)) {
            $token = $this->next_($string, $col);
            if ($token === False) {
                throw Exception("Unexpected token at col: $col");
            }
            $tokens[] = $token;
            $col += strlen($token[2]);
        }
        return $tokens;
    }

    private function next_($string, $i) {
        $string = substr($string, $i);
        foreach (self::$lexRegexStrings as $n => $regex) {
            if (preg_match("/^($regex)/", $string, $m)) {
                return Array($n, $i, $m[1]);
            }
        }
        return False;
    }

    private static $lexRegexStrings = Array(
        openParen => '\(',
        closeParen => '\)',
        exclude => '-',
        or_ => 'OR',
        whitespace => '\s+',
        operator => '[a-zA-Z]+:',
        quotedWord => '"[^"]*"',
        word => '[^\s()]+',
    );
}
