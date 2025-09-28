# Adverax DI

[![Go Reference](https://pkg.go.dev/badge/github.com/adverax/di.svg)](https://pkg.go.dev/github.com/adverax/di)  
[![Go Report Card](https://goreportcard.com/badge/github.com/adverax/di)](https://goreportcard.com/report/github.com/adverax/di)  
[![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)

Adverax DI is a lightweight and idiomatic dependency injection (DI) framework for Go, designed to be simple, efficient, and type-safe without relying on reflection or code generation.

## Key Features

- **Lightweight**: Minimalistic design, easy to integrate.
- **No Reflection**: Relies solely on Go's static typing, ensuring better performance and reliability.
- **No Code Generation**: Simplifies integration without requiring additional tools.
- **Type-Safe**: Eliminates the need for type casting, ensuring type safety at compile time.
- **Facilitates alive code**: Facilitates eliminating dead code in the IDE. Dead code is code that is never called. This is a common problem in large projects, where it is difficult to determine which code is used and which is not.
- **Helpful for graceful shutdown**: Facilitates graceful shutdown of the application. This is important for applications that need to release resources when they are no longer needed.
- **Support collections** of components for constructing multiple instances of the same type.

## Installation

Install the package using `go get`:

```bash
go get github.com/adverax/di
```

## Usage
See file example_test.go for more information.

## Documentation
Detailed documentation is available on pkg.go.dev.

## Contributing
Contributions are welcome! Please open issues for bug reports or feature requests. Pull requests are encouraged.

## License
This project is licensed under the Apache 2.0 License. See the LICENSE file for details.
