package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"pornboss/internal/jav"
)

func main() {
	var code string
	flag.StringVar(&code, "code", "", "solo movie code (e.g. IPX-633)")
	flag.Parse()

	if code == "" && flag.NArg() > 0 {
		code = flag.Arg(0)
	}
	if code == "" {
		log.Fatal("usage: go run ./cmd/javdatabaseactress -code IPX-633")
	}

	info, err := jav.JavDatabaseProvider.LookupActressByCode(code)
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
	fmt.Printf("HeightCM: %d\n", info.HeightCM)
	fmt.Printf("BirthDate: %d\n", info.BirthDate)
	fmt.Printf("Bust: %d\n", info.Bust)
	fmt.Printf("Waist: %d\n", info.Waist)
	fmt.Printf("Hips: %d\n", info.Hips)
	fmt.Printf("Cup: %d\n", info.Cup)
	fmt.Printf("ProfileURL: %s\n", info.ProfileURL)
}
