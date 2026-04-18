package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/tra-jkv/terraform-provider-bicc/bicc"
)

func main() {
	err := providerserver.Serve(context.Background(), bicc.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/tra-jkv/bicc",
	})
	if err != nil {
		log.Fatal(err)
	}
}
