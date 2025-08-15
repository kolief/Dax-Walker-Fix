# Dax Walker Fix
Routes walker.dax.cloud through your proxies.

## Download
[Download Latest Release](https://github.com/kolief/Dax-Walker-Fix/releases/latest) - Pre-built executable
Or build from source: `build.bat`

## How to use
1. Run as administrator: `daxwalkerfix.exe`
2. Select your proxy file when prompted (file dialog opens)
3. Choose SOCKS5 or HTTPS if asked

The app remembers your file location and proxy type for next time.

## Proxy file format
```
127.0.0.1:1080
proxy.example.com:3128
user:pass@10.0.0.5:1080
```

Or specify type:
```
socks5:127.0.0.1:1080
https:proxy.example.com:3128
```

## Options
`-timeout 30` - Auto-exit after 30 minutes of inactivity (default)
Press Ctrl+C to stop.

## What it does
- Tests proxies every 5 minutes
- Removes failed proxies from your file automatically
- Saves failed proxies to Desktop\DaxWalkerFix\failed_proxies.txt
- Auto-updates when new version available
- Remembers your settings in Desktop\DaxWalkerFix\remember.dat

## Notes
- Modifies Windows hosts file temporarily
- Runs local server on port 443
- Logs activity to daxwalkerfix.log
- Requires admin privileges

## Antivirus False Positives
Some antivirus software may flag this as malicious due to hosts file modification and network interception. This is a false positive - the tool only redirects walker.dax.cloud traffic.
