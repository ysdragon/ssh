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

The configuration file is located at `/ssh_config.yml` and supports the following options:

- `port` (under `ssh`): The port on which the SSH server will listen. Default is `2222`.
- `user` (under `ssh`): The username for SSH authentication.
- `password` (under `ssh`): The password for SSH authentication.
- `timeout` (under `ssh`): The timeout duration in seconds for SSH connections (leave it commented out or set it to `0` to disable it).
- `enable` (under `sftp`): Enable or disable SFTP functionality. Set to `true` to enable.


#### Example Configuration File

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