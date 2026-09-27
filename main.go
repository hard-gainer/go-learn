package main

import (
	"flag"
	"fmt"
	"runtime"
)

func main() {
// выводит имя пользователя из переменной окружения;
// считывает аргументы CLI и выводит их;
// показывает текущую версию Go через runtime.Version().
	userName := flag.String("name", "alex", "user name")
	flag.Parse()

	fmt.Println(*userName)
	fmt.Println(runtime.Version())
}



