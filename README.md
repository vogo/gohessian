# gohessian

**This is a feature-complete golang hessian serializer.**  

It's cloned from [viant/gohessian](README_old.md) , and do the following works:
- fix lots of bugs
- large scope refactoring to let code structure simple and human-friendly
- add [ref](http://hessian.caucho.com/doc/hessian-serialization.html##ref) and [date](http://hessian.caucho.com/doc/hessian-serialization.html##date) features implement
- more unit tests

## How to Use

[test examples](tests) are great start guide to use this library. 

[This example](tests/javamessage_test.go) shows how to integration with java.

## Reference
- [Hessian 2.0 Serialization Protocol](http://hessian.caucho.com/doc/hessian-serialization.html)
