

# This project is no longer maintained. TraeCN officially supports adding custom endpoints now. Continuing maintenance is no longer necessary.

# Trae-Proxy Go

This is the Golang implementation of Trae-Proxy, a high-performance API proxy tool specifically designed to intercept and redirect OpenAI API requests to custom backend services.

## Features

- **Smart Proxy**: Intercepts OpenAI API requests and forwards them to custom backends
- **Multi-Backend Support**: Configure multiple API backends with dynamic switching
- **Model Mapping**: Custom model ID mapping for seamless target model replacement
- **Streaming Responses**: Supports switching between streaming and non-streaming response modes
- **SSL Certificates**: Automatically generates and manages self-signed certificates
- **Auto Configuration**: Automatically installs CA certificates and configures the hosts file after certificate generation
- **TUI Interface**: A friendly terminal user interface for easy configuration management
- **High Performance**: Built with Go for excellent performance

## Usage

### System Requirements

- Go 1.21 or higher
- OpenSSL (used for certificate generation)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd trae-proxy-go

# Install dependencies
go mod download

# Build
go build -o trae-proxy ./cmd/proxy
go build -o trae-proxy-cli ./cmd/cli
```

### Configuration

Edit the `config.yaml` file:

```yaml
domain: api.openai.com

apis:
  - name: "deepseek-r1"
    endpoint: "https://api.deepseek.com"
    custom_model_id: "deepseek-reasoner"
    target_model_id: "deepseek-reasoner"
    stream_mode: null
    active: true

server:
  port: 443
  debug: true
```

### Using the TUI Interface (Recommended)

Running the CLI tool directly (without arguments) will start the TUI interface:

```bash
./trae-proxy-cli
```

TUI Interface Shortcuts:
- `a` - Add a new API configuration
- `e` - Edit the selected API configuration
- `d` - Delete the selected API configuration
- `Space` - Activate/Deactivate the selected API configuration
- `D` - Set proxy domain
- `C` - Generate SSL certificate
- `q` - Quit
- `↑↓` - Navigate up/down

### Using CLI Commands

You can also use traditional CLI commands:

#### Generate Certificate

```bash
# Generate certificate using CLI tool
./trae-proxy-cli cert

# Or specify a domain
./trae-proxy-cli cert --domain api.openai.com

# Generate certificate and auto-configure system (recommended, requires admin/root privileges)
./trae-proxy-cli cert --auto-config

# Install CA certificate to system trust store only
./trae-proxy-cli cert --install-ca

# Update hosts file only
./trae-proxy-cli cert --update-hosts
```

**Auto Configuration Feature:**

Using the `--auto-config` flag allows automatic completion of the following configurations after certificate generation:
1. Install the CA certificate to the system trust store
2. Update the hosts file to point the domain to 127.0.0.1

This allows you to use the proxy service immediately without manual configuration.

> **Note: Auto-configuration requires administrator privileges**
> - Windows: Run Command Prompt or PowerShell as Administrator
> - macOS/Linux: Run commands with `sudo`, or ensure passwordless sudo is configured
> - After generating a certificate in the TUI interface, you will also be prompted to perform auto-configuration
> - If auto-configuration fails, the tool will display detailed manual configuration instructions

### Start Server

```bash
# Use the compiled binary
./trae-proxy

# Or use go run
go run cmd/proxy/main.go

# Enable debug mode
./trae-proxy --debug
```

### CLI Tool Usage

#### List Configurations

```bash
./trae-proxy-cli list
```

#### Add API Configuration

```bash
./trae-proxy-cli add \
  --name "my-api" \
  --endpoint "https://api.example.com" \
  --custom-model "my-model" \
  --target-model "target-model" \
  --stream-mode none \
  --active
```

#### Update API Configuration

```bash
./trae-proxy-cli update \
  --index 0 \
  --name "updated-name" \
  --endpoint "https://new-api.example.com"
```

#### Activate API Configuration

```bash
./trae-proxy-cli activate --index 0
```

#### Delete API Configuration

```bash
./trae-proxy-cli remove --index 0
```

#### Update Domain

```bash
./trae-proxy-cli domain --name api.openai.com
```

### Client Configuration

#### Coexistence with Other Proxies / Conflict Troubleshooting

If you have other proxies running simultaneously on your machine (e.g., Clash / Surge / System Proxy / TUN mode, etc.), you may encounter two common types of conflicts:

1. **Port Conflict**: Trae-Proxy listens on `443` by default (see `server.port` in `config.yaml`). If another program is already using port 443, Trae-Proxy will fail to start.
   - **Solution**: Close the program occupying port 443, or change `server.port` to another port (and ensure the client can be configured with the corresponding port / base_url), using port forwarding if necessary.
2. **Routing Conflict (Most Common)**: If the client uses a "system/global proxy", requests may be intercepted by other proxies first (e.g., CONNECT `api.openai.com:443`). In this case, the local `hosts` file pointing to `127.0.0.1` may not take effect, and traffic won't reach Trae-Proxy.
   - **Solution**: Set `api.openai.com` to **Direct / Bypass** in your proxy software, and ensure `127.0.0.1` and `localhost` are not proxied.

##### One-Click Diagnostic

Use the CLI's `doctor` command to detect environment/system proxies and port usage, and generate recommended settings:

```bash
./trae-proxy-cli doctor

# Optional: Override config file, domain, and port
./trae-proxy-cli doctor --config config.yaml --domain api.openai.com --port 443
```

> **Note**: When forwarding to backends, Trae-Proxy uses Go's default proxy rules, which may be affected by the process environment variables `HTTP_PROXY/HTTPS_PROXY/ALL_PROXY/NO_PROXY`. The `doctor` command will provide recommended `NO_PROXY` values to prevent requests from being intercepted by other proxies.

#### 1. Obtain Server Self-Signed Certificate

Copy the CA certificate from the server to your local machine:

```bash
scp user@your-server-ip:/path/to/trae-proxy-go/ca/ca.crt .
```

#### 2. Install CA Certificate

##### Windows

1. Double-click the `ca.crt` file
2. Select "Install Certificate"
3. Select "Local Machine"
4. Select "Place all certificates in the following store" → "Browse" → "Trusted Root Certification Authorities"
5. Complete the installation

##### macOS

1. Double-click the `ca.crt` file to open "Keychain Access"
2. Add the certificate to the "System" keychain
3. Double-click the imported certificate and expand the "Trust" section
4. Set "When using this certificate" to "Always Trust"
5. Close the window and enter your admin password to confirm

#### 3. Modify hosts File

##### Windows

Edit `C:\Windows\System32\drivers\etc\hosts` and add:

```
your-server-ip api.openai.com
```

> If Trae-Proxy is running locally, replace `your-server-ip` with `127.0.0.1` (or `::1`).

##### macOS

Edit `/etc/hosts` and add:

```
your-server-ip api.openai.com
```

> If Trae-Proxy is running locally, replace `your-server-ip` with `127.0.0.1` (or `::1`).

#### 4. Test Connection

```bash
curl https://api.openai.com/v1/models
```

If configured correctly, you should see the list of models returned by the proxy server.

## Project Structure and Development

### Project Structure

```
trae-proxy-go/
├── cmd/
│   ├── proxy/          # Main proxy server executable
│   └── cli/            # CLI management tool
├── internal/
│   ├── config/         # Configuration management
│   ├── proxy/          # Core proxy logic
│   ├── cert/           # Certificate management
│   └── logger/         # Logging
├── pkg/
│   └── models/         # Data models
├── config.yaml         # Configuration file
├── ca/                 # Certificate directory
└── README.md
```

### Architecture

```
 +------------------+    +--------------+    +------------------+
 |                  |    |              |    |                  |
 |  DeepSeek API    +--->+              +--->+  Trae IDE        |
 |                  |    |              |    |                  |
 |  Moonshot API    +--->+              +--->+  VSCode          |
 |                  |    |  Trae-Proxy  |    |                  |
 |  Aliyun API      +--->+     Go       +--->+  JetBrains       |
 |                  |    |              |    |                  |
 |  Self-hosted LLM +--->+              +--->+  OpenAI Clients  |
 |                  |    |              |    |                  |
 +------------------+    +--------------+    +------------------+
   Backend Services       Proxy Server        Client Apps
```

### Differences from Python Version

1. **Performance**: The Go version offers better performance and lower resource usage
2. **Deployment**: Compiled into a single binary file, no Python runtime required
3. **Dependencies**: Fewer dependencies, only requires OpenSSL (for certificate generation)
4. **CLI Tool**: The CLI tool has similar functionality, but is implemented differently

### Development

```bash
# Run tests (if any)
go test ./...

# Format code
go fmt ./...

# Check code
go vet ./...
```

### GitHub Actions (CI / Release)

The repository includes a built-in GitHub Actions workflow: every push to `main` (and manual triggers via `workflow_dispatch`) will automatically run `go test ./...`, build executables for Windows/macOS/Linux, and finally upload the artifacts to the `nightly` pre-release version on GitHub Releases.

Reference for triggering workflows:
https://docs.github.com/zh/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow

## License

This project is licensed under the MIT License.

## Contributing

Issues and Pull Requests are welcome!
