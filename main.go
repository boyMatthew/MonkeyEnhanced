package main

import (
	"fmt"
	repl "myMonkey/monkey_repl"
	"os"
	"os/user"
	"time"
)

func main() {
	curUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is Monkey Programming Language Enhanced Version! \nFeel free to type in commands!\n", curUser.Username)
	time.Sleep(1 * time.Second)
	if err = repl.Start(os.Stdin, os.Stderr); err != nil {
		panic(err)
	}
}
