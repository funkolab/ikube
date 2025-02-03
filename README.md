# ikube

`ikube` is a command-line tool for managing Kubernetes configurations stored in Infisical. It allows you to authenticate, store, list, and delete kubeconfigs securely.

## Features

- **Authenticate**: Authenticate with Infisical using environment variables, keyring, or manual input.
- **Store Kubeconfig**: Store a new kubeconfig securely in Infisical.
- **List Kubeconfigs**: List and select kubeconfigs stored in Infisical.
- **Delete Kubeconfigs**: Delete kubeconfigs stored in Infisical.
- **Temporary Shell**: Load kubeconfig in a temporary shell session.

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/funkolab/ikube.git
    cd ikube
    ```

2. Build the binary:
    ```sh
    task build
    ```

3. Install the binary to `$GOPATH/bin`:
    ```sh
    task install
    ```

## Usage

### Command Line Flags

- `-v`: Enable verbose mode.
- `-l`: Load kubeconfig in a temporary shell.
- `-d`: Delete kubeconfig(s).

### Environment Variables

- `INFISICAL_SERVER`: The server URL for Infisical (default `app.infisical.com`).
- `INFISICAL_PROJECT_ID`: The project ID for Infisical.
- `INFISICAL_CLIENT_ID`: The client ID for Infisical (optional).
- `INFISICAL_CLIENT_SECRET`: The client secret for Infisical (optional).

### Examples

#### Authenticate and List Kubeconfigs

```sh
ikube
```

#### Store a New Kubeconfig

```sh
cat /path/to/kubeconfig | ikube
```

#### Delete Kubeconfigs

```sh
ikube -d
```

#### Load Kubeconfig in Temporary Shell

```sh
ikube -l
```

## Development

### Taskfile

This project uses [Task](https://taskfile.dev) for task management. The available tasks are defined in the `Taskfile.yml`.

- `task build`: Build the binary.
- `task install`: Install the binary to `$GOPATH/bin`.
- `task test`: Run tests.
- `task lint`: Run linter.
- `task all`: Run all tasks (lint, test, build).

### Running Tests

To run the tests, use:
```sh
task test
```

### Linting

To run the linter, use:
```sh
task lint
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Acknowledgements

- [Infisical](https://github.com/infisical/go-sdk)
- [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder)
- [go-keyring](https://github.com/zalando/go-keyring)
- [client-go](https://github.com/kubernetes/client-go)

---

For more information, please refer to the source code and documentation.
