# Dax Walker Fix

Redirects walker.dax.cloud traffic through SOCKS5 proxies to bypass rate limiting.

## Download

**[Download Latest Release](https://github.com/kolief/Dax-Walker-Fix/releases/latest)** - Pre-built executable

Or build from source: `build.bat`

## Setup

Edit `proxy.txt` with your SOCKS5 proxies:

```
192.168.1.100:1080:username:password
127.0.0.1:9050
```

## Usage

Run as administrator:

```
daxwalkerfix.exe
```

Options: `-timeout 10` (minutes), `-debug`

Press Ctrl+C to stop.

## Notes

- Modifies Windows hosts file temporarily
- Runs local server on port 443
- Logs activity to daxwalkerfix.log
- Requires admin privileges

## Antivirus False Positives

Some antivirus software may flag this as malicious due to:
- Hosts file modification (common in malware)
- Network interception capabilities
- Admin privilege requirements

This is a false positive. The tool only redirects walker.dax.cloud traffic and does not collect personal data or perform malicious activities. You can verify this by reviewing the open source code.