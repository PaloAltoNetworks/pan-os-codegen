package main

import (
    "context"
    "fmt"
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
    }

    if err != nil {
        fmt.Fprintf(os.Stderr, err.Error() + "\n")
        os.Exit(1)
    }
}
