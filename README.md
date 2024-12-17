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

#### Configuration Options

The configuration file is located at `/.ssh_config` and supports the following options:

- `SSH_PORT`: The port on which the SSH server will listen. Default is `2222`.
- `SSH_USER`: The username for SSH authentication.
- `SSH_PASSWORD`: The password for SSH authentication.
- `SSH_TIMEOUT`: The timeout duration in seconds for SSH connections (leave it empty or set it to `0` to disable it).
- `SFTP_ENABLE`: Enable or disable SFTP. Set to `true` to enable.

#### Example Configuration File

```ini
SSH_PORT=2222
SSH_USER=admin
SSH_PASSWORD=securepassword
SSH_TIMEOUT=300
SFTP_ENABLE=true
```

This project is open-source and available under the MIT License. See the [LICENSE](https://github.com/ysdragon/ssh/blob/master/LICENSE) file for more details.