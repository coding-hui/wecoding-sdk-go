# WeCoding Go Management SDK

This repository contains the IAM management SDK for Go, providing client libraries for IAM (Identity and Access Management) and AI/LLM services.

## Installation

```bash
go get github.com/coding-hui/wecoding-sdk-go
```

## Quick Start

### Programmatic Configuration

```go
import (
    "github.com/coding-hui/wecoding-sdk-go/services/iam"
    "github.com/coding-hui/wecoding-sdk-go/rest"
)

func main() {
    // Single endpoint for all operations
    client, err := iam.NewForConfig(rest.NewConfig(
        rest.WithEndpoint("https://iam.example.com"),
        rest.WithAccessKeyId("your-access-key"),
        rest.WithSecretAccessKey("your-secret-key"),
    ))

    // Separate endpoints for API and Authz servers
    client, err := iam.NewForConfig(rest.NewConfig(
        rest.WithAPIServerEndpoint("https://api.example.com"),
        rest.WithAuthzEndpoint("https://authz.example.com"),
        rest.WithAccessKeyId("your-access-key"),
        rest.WithSecretAccessKey("your-secret-key"),
    ))

    // With custom timeout
    client, err := iam.NewForConfig(rest.NewConfig(
        rest.WithEndpoint("https://iam.example.com"),
        rest.WithAccessKeyId("your-access-key"),
        rest.WithSecretAccessKey("your-secret-key"),
        rest.WithTimeout(60 * time.Second),
    ))
}
```

### Configuration Options

| Option | Description |
|--------|-------------|
| `WithEndpoint` | Primary endpoint (used when API and Authz share same host) |
| `WithAPIServerEndpoint` | Separate API server endpoint |
| `WithAuthzEndpoint` | Separate Authz server endpoint |
| `WithAccessKeyId` | AWS-style access key ID |
| `WithSecretAccessKey` | AWS-style secret access key |
| `WithBearerToken` | Bearer token for authentication |
| `WithTimeout` | Request timeout (default: 30s) |
| `WithCACertificate` | CA certificate data for TLS |

### YAML Configuration

Create a config file at `~/.iam/config`:

```yaml
server:
  endpoint: https://iam.example.com
  # For distributed deployment:
  # apiServerEndpoint: https://api.example.com
  # authzEndpoint: https://authz.example.com
credentials:
  accessKeyId: your-access-key
  secretAccessKey: your-secret-key
```

Load config from file:

```go
import "github.com/coding-hui/wecoding-sdk-go/tools/clientcmd"

config, err := clientcmd.BuildConfigFromFlags("", "~/.iam/config")
client, err := iam.NewForConfig(config)
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `IAM_ENDPOINT` | Primary endpoint |
| `IAM_API_SERVER_ENDPOINT` | API server endpoint |
| `IAM_AUTHZ_ENDPOINT` | Authz server endpoint |
| `IAM_ACCESS_KEY_ID` | Access key ID |
| `IAM_SECRET_ACCESS_KEY` | Secret access key |

## Usage Examples

### Authentication

```go
client, _ := iam.NewForConfig(rest.NewConfig(
    rest.WithEndpoint("http://localhost:8080"),
    rest.WithAccessKeyId("test-key-id"),
    rest.WithSecretAccessKey("test-secret-key"),
))

// Login
authResp, err := client.APIV1().Authentication().Authenticate(ctx, v1.AuthenticateRequest{
    Username: "ADMIN",
    Password: "WECODING",
})

// Get user info
userInfo, err := client.APIV1().Authentication().UserInfo(ctx, authResp.AccessToken)

// Refresh token
refreshResp, err := client.APIV1().Authentication().RefreshToken(ctx, authResp.RefreshToken)
```

### User Management

```go
// Create user
user, err := client.APIV1().Users().Create(ctx, &v1.CreateUserRequest{
    Name:     "username",
    Password: "Password@123",
    Email:    "user@example.com",
}, metav1.CreateOptions{})

// Get user
gotUser, err := client.APIV1().Users().Get(ctx, user.InstanceID, metav1.GetOptions{})

// List users
users, err := client.APIV1().Users().List(ctx, metav1.ListOptions{Offset: 0, Limit: 10})
```

### Authorization

```go
// Using clientset with separate endpoints
clientset, err := services.NewForConfig(rest.NewConfig(
    rest.WithAPIServerEndpoint("http://localhost:8080"),
    rest.WithAuthzEndpoint("http://localhost:9090"),
    rest.WithAccessKeyId("test-key-id"),
    rest.WithSecretAccessKey("test-secret-key"),
))

// Authorize request
result, err := clientset.Iam().AuthzV1().Authz().Authorize(ctx, &v1.Request{
    Resource: "project:name",
    Action:   "delete",
    Subject:  "users:username",
})
```

## Architecture

```
Clientset → IamClient → APIV1Client/AuthzV1Client → RESTClient → gorequest
```

- `services/clientset.go` - Top-level clientset containing all service clients
- `services/iam/iam_client.go` - IAM client grouping APIV1 and AuthzV1
- `services/iam/apiserver/v1/` - API server operations (authentication, users)
- `services/iam/authz/v1/` - Authorization operations
- `rest/` - HTTP REST client implementation using gorequest
