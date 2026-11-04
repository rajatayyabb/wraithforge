package main

import (
	"fmt"
	"os"

	"wraithforge/backend/pkg/taxonomy"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: taxoncheck <taxonomy.yaml>")
		os.Exit(1)
	}
	t, err := taxonomy.Load(os.Args[1])
	if err != nil {
		fmt.Println("parse error:", err)
		os.Exit(1)
	}
	for _, a := range t.Assets {
		fmt.Printf("%s\n", a.Kind)
		for _, f := range a.Fields {
			fmt.Printf("  %s %s\n", f.Name, f.Type)
		}
	}
}
