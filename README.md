# woltsapp

WhatsApp Wake-on-LAN (WoL) bot

## Building

```sh
git submodule update --init --recursive
cp qrcodejs/qrcode.min.js static
go build .
```

## Configuration

configuration is done via environment variables and command line arguments,
with the exception of LOG_LEVEL:

`LOG_LEVEL`: minimum log level, one of `DEBUG`, `INFO`, `WARN`, `ERROR`
(default: `INFO`)

see `woltsapp -help` for more information
