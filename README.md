# tickerfile

Windows service that writes UTC timestamps every 2 seconds to two log files. Each tick generates one timestamp and writes the same value to both files.

## Requirements

- Windows
- Go 1.22 or later
- Administrator privileges for `install`, `remove`, `start`, and `stop`

## Build

```powershell
go build -o tickerfile.exe .
```

## Usage

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-name` | `tickerfile` | Windows service name |
| `-logdir` | executable directory | Directory for log output |

### Commands

```powershell
# Run in foreground (no install, no admin)
.\tickerfile.exe debug

# Install and run as a Windows service (elevated PowerShell)
.\tickerfile.exe install
.\tickerfile.exe start

# Stop and remove
.\tickerfile.exe stop
.\tickerfile.exe remove
```

When the Service Control Manager starts the service, it launches `tickerfile.exe` with no arguments. The program detects service mode automatically and runs the ticker loop.

## Log files

Both files are written in the log directory (by default, next to the executable).

| File | Mechanism |
|------|-----------|
| `text.log` | Standard Go file append (`os.OpenFile` with `O_APPEND`) |
| `win32.log` | Win32 APIs: `CreateFile`, `ReadFile`, `WriteFile`, `FlushFileBuffers`, `SetEndOfFile` |

The Win32 log is opened with:

- `FILE_FLAG_OVERLAPPED`
- `FILE_FLAG_WRITE_THROUGH`

Each 2-second tick:

1. Generates one RFC3339 UTC timestamp (e.g. `2026-06-15T12:34:56Z`)
2. Writes the same line to `text.log`
3. Appends the same line to `win32.log` at EOF via overlapped I/O, then verifies with `ReadFile`

Comparing the two files line-by-line should show matching timestamps.

## Service workflow

1. `go build -o tickerfile.exe .`
2. `.\tickerfile.exe install` (as Administrator)
3. `.\tickerfile.exe start`
4. Check `text.log` and `win32.log` in the executable directory
5. `.\tickerfile.exe stop`
6. `.\tickerfile.exe remove`

## Troubleshooting

- **Access denied on install/start/stop** — run PowerShell or Command Prompt as Administrator.
- **Log files not created** — ensure the service account can write to the log directory (default: directory containing `tickerfile.exe`).
- **Debug without installing** — use `.\tickerfile.exe debug` to run in the console and write logs immediately.
