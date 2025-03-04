# SSH Server

This is a custom SSH server primarily designed for use with my [Pterodactyl-VPS-Egg](https://github.com/ysdragon/Pterodactyl-VPS-Egg) project. However, it can be used in any Docker container.

## Installation

### Prerequisites

- Go (for building from source)

### Building from Source

1. Clone the repository:
    ```sh
    git clone https://github.com/ysdragon/ssh.git
    cd ssh
    ```

2. Build the project:
    ```sh
    go build -o ssh main.go
    ```

3. Run the built binary:
    ```sh
    ./ssh
    ```

## Configuration

### Configuration Options

The configuration file is located at `/ssh_config.yml` and supports the following options:

### SSH Options

| Option | Description | Default |
|--------|-------------|---------|
| `port` | Port number for SSH server | `2222` |
| `user` | Username for SSH authentication | `root` |
| `password` | Password for SSH authentication (supports plain text or bcrypt hash) | `password` |
| `timeout` | Connection timeout in seconds (comment out or set to 0 to disable) | `300` |

### SFTP Options

| Option | Description | Default |
|--------|-------------|---------|
| `enable` | Enable or disable SFTP support | `true` |

> [!NOTE] 
> The `timeout` setting is optional and can be omitted from the configuration.

### Example `/ssh_config.yml` Configuration

Here is an example configuration file:

```yml
ssh:
  port: "2222"
  user: "root"
  password: "password"
  # timeout: 30

sftp:
  enable: true
```

This project is open-source and available under the MIT License. See the [LICENSE](https://github.com/ysdragon/ssh/blob/master/LICENSE) file for more details.