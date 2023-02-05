package main

import (
	"mini-docker/src"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	if os.Geteuid() != 0 {
		log.Fatal("You need root privileges to run this program.")
	}
	
	log.Printf("Cmd args: %v\n", os.Args)

	rand.Seed(time.Now().Unix())
	switch os.Args[1] {
	case "run":
		src.InitDockerDirs()
		src.Run(os.Args[2:]...)
	case "child":
		src.Child(os.Args[2:]...)
	default:
		fmt.Println("unknwon command...")
	}
}
