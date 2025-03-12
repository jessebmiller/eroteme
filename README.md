# Eroteme

Adding a Rusty `?` to Go.

Eroteme is a code transformation tool that adds Rust-like error handling to Go. It simplifies Go's error handling by transforming specially commented code into proper error-checking patterns.

## Installation

```bash
go install github.com/yourusername/eroteme/cmd/eroteme@latest
```

## Usage

### Basic Usage

Write Go code with `//?` comments:

```go
data, _ := readFile("example.txt") //? 
```

Run the `eroteme` tool:

```bash
eroteme yourfile.go
```

Get transformed Go code:

```go
data, err := readFile("example.txt")
if err != nil {
    return err
}
```

### Custom Return Values

Specify custom return values:

```go
value, _ := processValue(42) //? value, err
```

After transformation:

```go
value, err := processValue(42)
if err != nil {
    return value, err
}
```

### Multiple Return Values

```go
result, count, _ := complexOperation() //? result, count, err
```

After transformation:

```go
result, count, err := complexOperation()
if err != nil {
    return result, count, err
}
```

## Directory Processing

Process all Go files in a directory:

```bash
eroteme ./path/to/directory
```

## Limitations

- The eroteme tool modifies your code files in place.
- Currently only supports assignments with blank identifiers for error values.
- The function containing the eroteme statement must have a compatible return type.
