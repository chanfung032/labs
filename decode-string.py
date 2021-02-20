#-*-coding: utf-8

"""
394. 字符串解码

给定一个经过编码的字符串，返回它解码后的字符串。

编码规则为: k[encoded_string]，表示其中方括号内部的 encoded_string 正好重复 k 次。注意 k 保证为正整数。

你可以认为输入字符串总是有效的；输入字符串中没有额外的空格，且输入的方括号总是符合格式要求的。

此外，你可以认为原始数据不包含数字，所有的数字只表示重复的次数 k ，例如不会出现像 3a 或 2[4] 的输入。

示例:

s = "3[a]2[bc]", 返回 "aaabcbc".
s = "3[a2[c]]", 返回 "accaccacc".
s = "2[abc]3[cd]ef", 返回 "abcabccdcdcdef".

https://leetcode-cn.com/problems/decode-string/

---

解法：

参见 http://effbot.org/zone/simple-iterator-parser.htm 手写一个递归下降 Parser
"""
import re

def atom(next, token):
    if token.isdigit():
        n = int(token)
        token = next()
        assert token == '['
        out = []
        token = next()
        while token != ']':
            out.append(atom(next, token))
            token = next()
        return (''.join(out)) * n
    else:
        return token

def simple_eval(expr):
    tokens = iter(re.findall(r'\d+|[^0-9]', expr))
    out = []
    while True:
        try:
            out.append(atom(tokens.next, tokens.next()))
        except StopIteration:
            break
    return ''.join(out)

if __name__ == '__main__':
    assert simple_eval('3[a]2[bc]') == 'aaabcbc'
    assert simple_eval('3[a2[c]]') == 'accaccacc'
    assert simple_eval('2[abc]3[cd]ef') == 'abcabccdcdcdef'
