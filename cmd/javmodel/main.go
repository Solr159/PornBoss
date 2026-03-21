package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"pornboss/internal/jav"
)

func main() {
	var name string
	flag.StringVar(&name, "name", "", "actress japanese name (e.g. 桃乃木かな)")
	flag.Parse()

	if name == "" && flag.NArg() > 0 {
		name = flag.Arg(0)
	}
	if name == "" {
		log.Fatal("usage: go run ./cmd/javmodel -name 桃乃木かな")
	}

	info, err := jav.JavModelProvider.LookupActressByJapaneseName(name)
	if err != nil {
		if errors.Is(err, jav.ResourceNotFonud) {
			fmt.Println("actress not found")
			return
		}
		log.Fatalf("lookup error: %v", err)
	}
	if info == nil {
		fmt.Println("actress not found")
		return
	}

	fmt.Printf("RomanName: %s\n", info.RomanName)
	fmt.Printf("JapaneseName: %s\n", info.JapaneseName)
	fmt.Printf("ChineseName: %s\n", info.ChineseName)
	fmt.Printf("HeightCM: %d\n", info.HeightCM)
	fmt.Printf("BirthDate: %d\n", info.BirthDate)
	fmt.Printf("Bust: %d\n", info.Bust)
	fmt.Printf("Waist: %d\n", info.Waist)
	fmt.Printf("Hips: %d\n", info.Hips)
	fmt.Printf("Cup: %d\n", info.Cup)
	fmt.Printf("ProfileURL: %s\n", info.ProfileURL)
}
