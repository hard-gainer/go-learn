package main

import "fmt"

// Декоратор(обертка функций) - функция, принимающая другую функцию и
// возвращающая функцию того же типа.

type Handler func(func()) func()

func printHello() {
	fmt.Println("Hello world!")
}

func main() {
	var h Handler = func(f func()) func() {
		return func() {
			fmt.Println("before")
			f()
			fmt.Println("after")
		}
	}

	f := h(printHello)
	f()
}
