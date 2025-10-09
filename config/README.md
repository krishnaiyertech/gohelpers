# Config Package

A Go package for automatic configuration management with YAML files and CLI flags using reflection.

## Features

- **Auto-generate CLI flags** from struct tags
- **YAML config file support** with flag override
- **Nested struct support** with dot notation
- **Type-safe** reflection-based flag generation
- **Precedence order**: config file < CLI flags

## Quick Start

### 1. Define your config struct

```go
type Config struct {
    Name    string        `name:"name" short:"n" description:"Application name"`
    Port    int           `name:"port" short:"p" description:"Server port"`
    Debug   bool          `name:"debug" description:"Enable debug mode"`
    Timeout time.Duration `name:"timeout" description:"Request timeout"`
}
```

### 2. Create manager and integrate with Cobra

```go
func main() {
    config := &Config{}
    manager, err := config.New(config, "")
    if err != nil {
        log.Fatal(err)
    }

    cmd := &cobra.Command{
        Use: "myapp",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Parse config file and apply flag overrides
            if err := manager.ParseConfiguration(cmd); err != nil {
                return err
            }

            // Use your config
            fmt.Printf("Running %s on port %d\n", config.Name, config.Port)
            return nil
        },
    }

    cmd.Flags().AddFlagSet(manager.FlagSet())
    cmd.Execute()
}
```

### 3. Create config.yml

```yaml
name: "myapp"
port: 8080
debug: true
timeout: "30s"
```

### 4. Run with overrides

```bash
# Use config file
./myapp

# Override with flags
./myapp --name "override" --port 9090

# Use custom config file
./myapp --config ./custom.yml --debug=false
```

## Struct Tags

| Tag           | Description           | Example                     |
| ------------- | --------------------- | --------------------------- |
| `name`        | Flag name (required)  | `name:"port"`               |
| `short`       | Short flag (optional) | `short:"p"`                 |
| `description` | Help text             | `description:"Server port"` |

## Supported Types

- Basic types: `string`, `int`, `bool`, `float32/64`, `time.Duration`
- Integer types: `int8/16/32/64`, `uint8/16/32/64`
- Collections: `[]string`, `map[string]string`
- Nested structs (with dot notation: `server.port`)

## Nested Configuration

```go
type ServerConfig struct {
    Host string `name:"host" description:"Server host"`
    Port int    `name:"port" description:"Server port"`
}

type Config struct {
    Server ServerConfig `name:"server"`
}
```

Generates flags: `--server.host`, `--server.port`
