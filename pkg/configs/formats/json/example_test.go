package jsonConfig

import "fmt"

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
	// This example demonstrates how to use JSON loader.
	//
	// First, create loader:
	loader, err := NewFileLoaderBuilder().
		WithFile("config.global.json", false).
		WithFile("config.local.json", false).
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
