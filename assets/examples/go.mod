// This file converts assets/examples/ into its own go module.
//
// This allows us to run commands like go get ./... in the project root
// without getting errors about missing imports used by examples that
// reference generated code.
