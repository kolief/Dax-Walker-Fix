# Dax Walker Fix

Redirects walker.dax.cloud traffic through SOCKS5 or HTTPS proxies to bypass rate limiting.

## Download

**[Download Latest Release](https://github.com/kolief/Dax-Walker-Fix/releases/latest)** - Pre-built executable

Or build from source: `build.bat`

## Setup

Create `proxy.txt` on your Desktop with your SOCKS5 or HTTPS proxies:
The app looks for `proxy.txt` on your Desktop. If it doesn't find one, it will create a starter file there for you.

You can also paste proxies without a type. On startup the app will ask if these are SOCKS5 or HTTPS and use that for all untyped lines.

```
socks5:127.0.0.1:9050
https:proxy.example.com:3128
127.0.0.1:9050
proxy.example.com:3128
10.0.0.5:1080:username:password
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