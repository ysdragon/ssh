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
    go build -o ssh ./cmd/ssh-server
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
| `password` | Password for SSH authentication (supports plain text, bcrypt hash, or argon2 hash) | `password` |

### Password Hashing

The SSH server supports multiple password hashing algorithms:

- **Plain Text**: Store passwords in plain text (not recommended for production)
- **Bcrypt**: Store passwords as bcrypt hashes
- **Argon2**: Store passwords as argon2id hashes (recommended for security)

To generate a bcrypt hash, you can use tools like `htpasswd -B -n username` or online bcrypt generators.
To generate an argon2 hash, you can use tools like `argon2` command-line tool or online generators.
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
  password: "password"  # Can be plain text, bcrypt hash ($2a$...), or argon2 hash ($argon2id$...)
  # timeout: 30

sftp:
  enable: true
```

#### Example with Hashed Password

```yml
ssh:
  port: "2222"
  user: "root"
  password: "$argon2id$v=19$m=65536,t=3,p=2$..."  # Argon2 hash example
  timeout: 300

sftp:
  enable: true
```

This project is open-source and available under the MIT License. See the [LICENSE](LICENSE) file for more details.