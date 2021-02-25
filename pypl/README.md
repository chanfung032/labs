pypl
====

A compiler written for PL/0 lanuage with python, the source will be compiled 
into python bytecode so it can run on a python vm directly.

```pascal
program main;
begin
    write('hello, world!')
end.
```

The PL/0 source can be compiled and run using the following code:

```python
import pypl
exec(pypl.compile(source, '<none>'))
```

I extended the lanuage a little to support string and function call so as to make 
it more interesting to play with.

Actually, this is still python (a subset of python) with a different syntax. And 
because we run on a real python vm, we can write some freak programs like this:

```pascal
program main;
begin
    write(range(1000))
end.
```

PL/0 does not support literal list, but we can call python builtins to return one.

Totally a useless project, just writed for fun, it is the by-product during the 
study of Python compilation process :)

refs:

+ [PL/0](http://en.wikipedia.org/wiki/PL/0)
+ [Green Tree Snakes - the missing Python AST docs](http://greentreesnakes.readthedocs.org/)
+ [Pyparsing: A libraray for building recursive descent parser](http://pyparsing.wikispaces.com/)
+ [Howto gen python bytecode from scratch](http://aisk.sinaapp.com/?p=164)
+ [The offical docs for ast module](http://docs.python.org/2/library/ast.html)
+ [Some articles on the compliation process of Python](http://blog.csdn.net/atfield/article/category/256448)
+ [Compiling Little Languages in Python](http://www.python.org/workshops/1998-11/proceedings/papers/aycock-little/aycock-little.html)
+ [Language Implementation Patterns](http://book.douban.com/subject/4030327/)
+ http://onlamp.com/pub/a/python/2006/01/26/pyparsing.html
+ http://effbot.org/zone/simple-top-down-parsing.htm
  django used this algorithm in the parsing process of `if` block, see: django/template/smartif.py
