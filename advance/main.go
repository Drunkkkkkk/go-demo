package main

import "fmt"

func main() {
	ch := make(chan string)
	//ch <- 1
	ch = nil
	//close(ch)
	res := <-ch
	fmt.Println(res)
}
