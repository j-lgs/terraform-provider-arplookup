package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"terraform-provider-arplookup/internal/arplookup"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

/*go:generate go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...*/
//go:generate go run honnef.co/go/tools/cmd/staticcheck ./...

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/jlgs/arplookup",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), arplookup.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
