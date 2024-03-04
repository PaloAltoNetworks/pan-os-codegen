package main

import (
	"context"
	"log"
	"os"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/mktp"
)

func main() {
    var err error
    ctx := context.Background()

	cmd := mktp.Command(ctx, os.Args[1:]...)
	err = cmd.Setup()
	if err == nil {
		err = cmd.Execute()
	} else {
		log.Fatalf("There was an error when the execution: %s", err)
	}
}
