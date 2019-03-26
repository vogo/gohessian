# **This is a feature-complete golang hessian serializer.**  

[![Build Status](https://travis-ci.org/vogo/gohessian.png?branch=master)](https://travis-ci.org/vogo/gohessian)
[![GoCover](http://gocover.io/_badge/github.com/vogo/gohessian)](http://gocover.io/github.com/vogo/gohessian)
[![GoDoc](https://godoc.org/github.com/vogo/gohessian?status.svg)](https://godoc.org/github.com/vogo/gohessian)


It's cloned from [viant/gohessian](README_old.md) , and do the following works:
- fix lots of bugs
- large scope refactoring to let code structure simple and human-friendly
- add [ref](http://hessian.caucho.com/doc/hessian-serialization.html##ref) and [date](http://hessian.caucho.com/doc/hessian-serialization.html##date) features implement
- more unit tests
- api refactoring

# Usage

## name map and type map

- `name map`: a `map[string]string` used by `encoder` to determine the class name of a object
- `type map`: a `map[string]reflect.Type` used by `decoder` to determine the type of instance to initialize

You can use function `hessian.ExtractTypeNameMap(interface{})` to generate both the type map and name map. 
It's the recommendation way. 
Of course, you can create by yourself, but make sure them contain all names and types which encoder and decoder needed.

## simple example

```golang
type circular struct {
	Num      int
	Previous *circular
	Next     *circular
}

func main() {
	c := &circular{}
	c.Num = 12345
	c.Previous = c
	c.Next = c

	// create hessian serializer
	serializer := hessian.NewSerializer(hessian.ExtractTypeNameMap(c))

	fmt.Println("source object: ", c)

	// encode to bytes
	bytes, err := serializer.Encode(c)
	if err != nil {
		panic(err)
	}

	// decode from bytes
	decoded, err := serializer.Decode(bytes)
	if err != nil {
		panic(err)
	}
	fmt.Println("decode object: ", decoded)
}
```

## using java class name

You can define a function `HessianCodecName() string` for your struct if using `hessian.ExtractTypeNameMap(interface{})` to generate type map and name map.

```golang
type TraceVo struct {
	Key   string
	Value string
}

func (TraceVo) HessianCodecName() string {
	return "hessian.TraceVo"
}

typeMap,nameMap := hessian.ExtractTypeNameMap(&TraceVo{})
```

If you create type map and name map manually, you should also add the java class name mapping.

## concurrently

`hessian.NewSerializer` contains serialization processing data, so a serializer can't be used concurrently, you should create a new one when needed.

If there is only one type of data to serialize , a goroutine can continue use the same serializer to `Encode()` or `Decode()`.

## streaming transport

The following is a client-server streaming transport example:

server side:
```golang
_,nameMap := hessian.ExtractTypeNameMap(object)
encoder := hessian.NewEncoder(outputStreamWriter, nameMap) // write stream to outputStreamWriter
for {
    data := getNewData()
    err = encoder.Write(data) // write new data
    if err != nil {
        panic(err)
    }
}
```

client side:
```golang
typeMap,_ := hessian.ExtractTypeNameMap(object)
decoder := hessian.NewDecoder(inputStreamReader, typeMap) // read stream from inputStreamReader
for {
    obj,err := decoder.Read() // read new object
    if err != nil {
        panic(err)
    }
    
    // process obj
}
```

# Reference
- [Hessian 2.0 Serialization Protocol](http://hessian.caucho.com/doc/hessian-serialization.html)
