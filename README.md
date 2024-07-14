# woltsapp

WhatsApp Wake-on-LAN (WoL) bot

## Building

```sh
git submodule update --init --recursive
cp qrcodejs/qrcode.min.js static
go build .
```

## Configuration

configuration is done via environment variables

- `WOLTSAPP_LANG`: frontend and bot language
- `WOLTSAPP_HTTP_ADDR`: qr login server bind address
