package main

import "fmt"

type Service interface {
	Call(key string)
}

type ServiceFunc func(key string)

func (s ServiceFunc) Call(key string) {
	s(key)
}

type EchoService struct{}

func (echo EchoService) Call(key string) {
	fmt.Println("echo service call, key : ", key)
}

func PrintKey(key string) {
	fmt.Println("print key:", key)
}

func Foo(s Service, key string) {
	s.Call(key)
}

func main() {
	Foo(new(EchoService), "echo") //struct 作为参数

	Foo(ServiceFunc(func(key string) {
		fmt.Println("service func, call key,", key)
	}), "serviceFunc") //匿名函数就地实现 上面的接口型函数

	Foo(ServiceFunc(PrintKey), "printkey") //普通函数为参数 上面的接口型函数类型
}
