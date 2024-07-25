# woltsapp

WhatsApp Wake-on-LAN (WoL) bot

this is my first ever project in golang so don't expect
code quality or correctness

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

## TODO

- [ ] update command
- [ ] better documentation
