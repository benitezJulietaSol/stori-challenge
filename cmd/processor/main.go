package main

import (
	"awesomeProject2/cmd/processor/handler"
	"fmt"
)

func main() {

	handler.LambdaEvent()
	fmt.Println("LISTO")
	//lambda.Start(handler.LambdaEvent)
}
