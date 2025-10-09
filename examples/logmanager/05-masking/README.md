# Data Masking Examples

Comprehensive examples demonstrating sensitive data masking capabilities for protecting privacy in logs and monitoring.

## Features Demonstrated

- **Custom Masking Configuration**: Field-specific masking rules
- **Convenience Functions**: Pre-built masking for common scenarios
- **Multiple Masking Types**: Full mask, partial mask, pattern-based
- **Struct Tag Support**: Go struct tag integration
- **JSON Path Support**: Complex nested field masking

## Masking Types

| Type | Description | Example |
|------|-------------|---------|
| **FullMask** | Complete field replacement | `"password": "********"` |
| **PartialMask** | Show first/last characters | `"email": "joh***@example.com"` |
| **PatternMask** | Custom pattern replacement | `"card": "****-****-****-1234"` |

## Examples Overview

### 1. Custom Masking Configuration

```go
maskingConfigs := []logmanager.MaskingConfig{
    {
        FieldPattern: "email",
        Type:         logmanager.PartialMask,
        ShowFirst:    3,
        ShowLast:     10, // Show @domain.com
    },
    {
        FieldPattern: "password",
        Type:         logmanager.FullMask,
    },
    {
        JSONPath:  "$.phone",
        Type:      logmanager.PartialMask,
        ShowFirst: 3,
        ShowLast:  4,
    },
}
```

### 2. Convenience Functions

Pre-built masking for common use cases:

```go
// Password masking
txn := lmresty.NewTxnWithPasswordMasking(resp)

// Email masking
txn := lmresty.NewTxnWithEmailMasking(resp)

// Credit card masking
txn := lmresty.NewTxnWithCreditCardMasking(resp)
```

### 3. Struct Tag Integration

```go
type User struct {
    Email    string `json:"email" mask:"email"`
    Password string `json:"password" mask:"password"`
}
```

## Configuration Options

### Field Pattern Matching
```go
{
    FieldPattern: "password",     // Matches any field named "password"
    Type:         logmanager.FullMask,
}
```

### JSON Path Targeting
```go
{
    JSONPath: "$..token",         // Recursive wildcard - any "token" field
    Type:     logmanager.FullMask,
}
```

### Partial Masking
```go
{
    FieldPattern: "apiKey",
    Type:         logmanager.PartialMask,
    ShowFirst:    4,              // Show first 4 characters
    ShowLast:     4,              // Show last 4 characters
}
```

## Running the Example

```bash
go run main.go
```

## Expected Output

The example demonstrates masking in action:

1. **Custom Configuration**: User data with email/password/phone masking
2. **Password Masking**: Authentication data with secrets masked
3. **Email Masking**: User data with email addresses partially masked
4. **Credit Card Masking**: Payment data with card numbers masked

## Common Masking Patterns

### Authentication Data
```go
// Masks: password, token, secret, client_secret
lmresty.NewTxnWithPasswordMasking(resp)
```

### Personal Information
```go
// Masks: email, backup_email
lmresty.NewTxnWithEmailMasking(resp)
```

### Payment Data
```go
// Masks: card_number, cvv, security_code
lmresty.NewTxnWithCreditCardMasking(resp)
```

## Best Practices

1. **Configure Early**: Set up masking in application initialization
2. **Use Conventions**: Apply consistent field naming for automatic masking
3. **Layer Security**: Combine multiple masking strategies
4. **Test Thoroughly**: Verify masking works as expected
5. **Document Rules**: Maintain clear masking rule documentation

## Custom Masking Rules

```go
// Application-level masking
app := logmanager.NewApplication(
    logmanager.WithMaskingConfig([]logmanager.MaskingConfig{
        {
            Type:     logmanager.FullMask,
            JSONPath: "$..ssn",          // Social Security Numbers
        },
        {
            Type:     logmanager.PartialMask,
            JSONPath: "$..account_number",
            ShowFirst: 2,
            ShowLast:  4,
        },
    }),
)
```

## Security Considerations

1. **Comprehensive Coverage**: Ensure all sensitive fields are masked
2. **Nested Data**: Use JSON paths for complex nested structures
3. **Dynamic Fields**: Consider masking dynamically named fields
4. **Logging Levels**: Apply masking consistently across log levels
5. **Compliance**: Meet regulatory requirements (GDPR, PCI-DSS)

## Common Sensitive Fields

- Authentication: `password`, `token`, `secret`, `api_key`
- Personal: `email`, `phone`, `ssn`, `tax_id`
- Financial: `card_number`, `cvv`, `account_number`, `routing_number`
- Medical: `patient_id`, `medical_record`, `diagnosis`

## Next Steps

- [HTTP Servers](../02-http-servers/) - Apply masking in web applications
- [Messaging](../04-messaging/) - Mask sensitive message data
- [Basic Usage](../01-basic/) - Understand core logging concepts