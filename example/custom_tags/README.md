# Custom Tags

This takes the [URL example](../custom_types/README.md) and reimplements the `SecureURL` type using a custom tag. This change
allows the `net.URL` type to be used throughout the config so avoiding the need to cast. This is a great improvement on
the URL sample.

By the time this is released I will have added native support for `net.URL` into `goconfig` with the ability to use struct
tags to validate the scheme for `https`. This sample remains as a demonstration of how to use custom tags.

This sample will fall back to `env.example` in the current working directory.

## Testing

If you run the sample from this working directory, it will output the values from `env.example`.

You can demonstrate the custom Secure URL validation using

```bash
export WHATSAPP_SERVER_URL=http://localhost:3000/
go run .
```

You can demonstrate that explicitly blank values can override both defaults and later keystore (the `env.example` file)

```bash
export WHATSAPP_SERVER_URL=
go run .
```

You can demonstrate that the `SecureURL` type is based on, but does not change, the `net.URL` type by providing a non-secure
URL for `MY_BASE_URL`

```bash
export MY_BASE_URL=http://localhost:3000
go run .
```

