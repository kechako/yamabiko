# yamabiko

yamabiko is a SKK server written in Go.

## Configuration

The configuration file is located at `~/.config/yamabiko/config.toml`.

```toml
host = ""
port = 1178
send_encoding = "utf-8"
recv_encoding = "utf-8"

[logging]
path = ""
level = "info"
json = false

[[dictionaries]]
path = "/path/to/dictionary"
encoding = "euc-jp"
```

## Usage

To run the server, use the following command:

```bash
yamabiko server
```
