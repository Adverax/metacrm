package mixedConfig

import (
	"fmt"
	envFetcher "github.com/adverax/metacrm/pkg/access/fetchers/maps/env"
	yamlConfig "github.com/adverax/metacrm/pkg/configs/formats/yaml"
)

type MyConfigAddress struct {
	Host string `config:"host"`
	Port int    `config:"port"`
}

type MyConfig struct {
	Address MyConfigAddress `config:"address"`
	Name    string          `config:"name"`
}

func DefaultConfig() *MyConfig {
	return &MyConfig{
		Address: MyConfigAddress{
			Host: "unknown",
			Port: 80,
		},
		Name: "unknown",
	}
}

func Example() {
	// This example demonstrates how to use Mixed loader.
	//
	// First, create loader:
	loader, err := yamlConfig.NewFileLoaderBuilder().
		WithFile("config.global.json", false).
		WithFile("config.local.json", false).
		WithSource(
			envFetcher.New(
				envFetcher.NewPrefixGuard("MYAPP_"),
				envFetcher.NewKeyPathAccumulator("_"),
			),
		).
		Build()
	if err != nil {
		panic(err)
	}

	// Then load configuration:
	config := DefaultConfig()
	err = loader.Load(config)
	if err != nil {
		panic(err)
	}

	// Now you can use config.
	// For example, print it:
	fmt.Println(*config)

	// Output:
	// {{google.com 91} My App}
}
