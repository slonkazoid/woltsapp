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
with the exception of LOG_LEVEL, which is only available as an env. variable:

`LOG_LEVEL`: minimum log level, one of `DEBUG`, `INFO`, `WARN`, `ERROR`
(default: `INFO`)

see `woltsapp -help` for more information

## Usage

see [i18n/help.en.md](i18n/help.en.md) for a list of commands.  
the default prefix is `w,`

## TODO

- [ ] update command
- [ ] better documentation

## License

woltsapp is licensed under the [MIT License](LICENSE)

[whatsmeow](https://github.com/tulir/whatsmeow) is licensed under the
[Mozilla Public License 2.0](https://github.com/tulir/whatsmeow/blob/main/LICENSE)  
Copyright The whatsmeow Contributors
