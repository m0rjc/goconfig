# Structured Logging

This demonstration shows how I am currently handling logging in my wide-game-bot project.
There's a little Claude in there, and I'm tidying as go. I need to work on how I'll 
log with request ID. (I need to look into how slog's LogContext works. It should be a case
of picking up the request ID from the current GO context.Context and including it in the
log messages)

My log component here reads its own initialization rather than relying on the central
configuration loader. This allows me to initialize logging quickly. I set up some sensible
fallback defaults. There is an issue in GoConfig [#29](https://github.com/m0rjc/goconfig/issues/29) to better support this use case by having
GoConfig use the declared defaults if a user supplied value is invalid. For now, the defaulting
is provided in code. See `logger/config.go`, copied below:

```go
func loadConfig() (*LogConfig, error) {
	// Defaulting is provided here. If the user input is invalid, then GoConfig will not
	// overwrite these values. Issue 29 would have GoConfig use the declared defaults instead,
	// so avoiding the need for this defaulting code.
	config := LogConfig{
		Level:  slog.LevelInfo,
		Format: LogFormatText,
	}

	// The two custom types are registered here, so only applying them for this request.
	// This prevents the logging package affecting the wider application.
	configError := goconfig.Load(context.Background(), &config,
		goconfig.WithCustomType(typeLogFormat),
		goconfig.WithCustomType(typeLogLevel))

	return &config, configError
}
```

The code was written to support logging to a file using Luberjack. I'm running in Kubernetes
at the moment, so my dev system logs to stdout to be picked up by the Kubernetes logging
stack. The Lumberjack code is commented out here to allow this example to compile without
the Luberjack dependency.

This code is expected to error, so for that reason will exit 0. This will allow the
GitHub pull request action to pass.

# Types for Log Level and Log Format

This example uses GoConfig custom types for Log Level and Log Format. These are defined in
the logger package. To see this fail, try

```bash
export LOG_LEVEL=INCORRECT
go run .
```

I've not raised an issue to add Log Level to GoConfig yet. Is it better that GoConfig supports
so many types? Or should I leave it to end users to provide their own as needed? The cost of
providing this in GoConfig is small, but it adds to the amount of hidden knowledge in the library.
This can be easily added if there is demand.