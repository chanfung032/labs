# Chop Chop

一个修改版过的 yacc，生成的解析器可以识别语法片段（可以构建出部分语法树）。

大致逻辑：

1. 使用 yacc 生成 action & goto 表。
2. 初始状态的确定：根据第一个 token（terminal）找出所有初始可能的状态。
3. 然后对所有可能状态，shift reduce 进行规约。
4. 如果 reduce 后栈为空，则遍历入栈的 non-terminal 所有可能的下一个状态。
5. 最后如果最终规约的 token 数 > 3 ，ast 层数 > 1，则很有可能是一个符合语法规则的语法片段。

相关：

![lrparser](lrparser.png?raw=1)

- https://blog.chaitin.cn/sqlchop-the-sqli-detection-engine/
- [Creation Of LR Parser Table](https://www.youtube.com/watch?v=wwc3pUUahJk)
- https://www.cs.umd.edu/class/spring2014/cmsc430/lectures/lec07.pdf
- http://www3.cs.stonybrook.edu/~cse304/Fall08/Lectures/lrparser-handout.pdf
