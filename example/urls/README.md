# Custom URL Types

This is how I registered types for URL and Secure URL in my wide-game-bot project. I've since added an URL type to this
library, though the code here still works.

My example registers `*net.URL` and `*SecureURL` globally. That wasn't really necessary. I only have one config load
call and could have overridden them using options in the call to Load. If I wanted all URLs to be secure then I could
have used my Secure URL setup as the override for `*net.URL` at that level.

My project has a large central config package which all the other components depend on. It's not very encapsulated
of them, but fine at the moment in a monolithic application like that. If I ever split the configuration, so that each package
exports its own configuration structure, I'll have to look at how best to set up this tool to use them. (I can imagine
a key prefix mechanism being useful and easy to implement). The monolithic configuration package does mean that all recognized
environment variables are listed in one place, and using `goconfig` has greatly reduced the noise in that file.

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

## Learning Points

When I created my custom type `SecureURL` I lost all the methods that `net.URL` had. This came as a surprise. It's not
simple inheritance as in other languages. My sample provides a `String()` method that prints the URL in a format that
`net.URL` would print. By the time this is released I will have added native support for `net.URL` into `goconfig`
with the ability to use struct tags to validate the scheme for `https`. This will remove the need for `SecureURL`.

I could have used a custom `Wrapper` to add a `secure:bool` tag to the `URL` type. I will do this in another demonstration.
The simplest solution using `goconfig` as of `v0.3.0` would have been to use the existing `pattern` tag.