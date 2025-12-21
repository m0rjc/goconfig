# Security Guidelines for Maintainers

This document outlines security considerations and design decisions for goconfig maintainers.

## Error Messages Must Not Expose Values

**Critical Security Rule:** Error messages must NEVER include the actual configuration values.

### Rationale

Configuration values often contain sensitive data:
- API keys and tokens (e.g., `API_KEY=sk-secret-token-12345`)
- Database credentials and connection strings
- OAuth secrets
- Private keys
- Passwords

When validation or parsing fails, these values could be exposed through:
- Application logs
- Centralized logging systems (CloudWatch, Datadog, Splunk, etc.)
- Error tracking services (Sentry, Bugsnag, Rollbar, etc.)
- Console output during development
- Monitoring dashboards
- Stack traces in production

### Implementation Requirements

1. **Parsing Errors** (config.go:242)
   - ❌ Bad: `"error parsing value %s: %w", configuredValue, err`
   - ✅ Good: `"error parsing value: %w", err`

2. **Validation Errors** (validation.go)
   - ❌ Bad: `"value %s does not match pattern %s", actualValue, pattern`
   - ✅ Good: `"does not match pattern %s", pattern`
   - ❌ Bad: `"value %d is below minimum %d", actualValue, min`
   - ✅ Good: `"below minimum %d", min`

3. **Custom Validators**
   - Validators should avoid including actual values in error messages
   - Example: Return `"must start with 'sk-'"` instead of `"'invalid-key' must start with 'sk-'"`

### Code Review Checklist

When reviewing changes that modify error messages:

- [ ] Error message does not include `%s`, `%v`, or `%d` for the actual value
- [ ] Only constraint information (min, max, pattern) is included, not user data
- [ ] Tests verify the error message format
- [ ] Documentation examples reflect the correct error format

### Related Issues

- Issue #15: Original security concern about secrets being logged

### Testing

When adding new validation:
1. Write tests that check the exact error message format
2. Ensure test expectations do NOT include actual values
3. Verify error messages only describe what's wrong, not what was provided

### Exception: Debugging Information

Even in debug mode or verbose logging, actual configuration values should not be included in error messages. If debugging requires seeing values:
- Users should check their environment variables directly
- Application logs should be reviewed at the configuration loading stage (before validation)
- Never compromise security for convenience
