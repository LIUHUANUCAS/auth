# Auth Server with WeChat Mini Program Support

This is an authentication server that supports both traditional username/password authentication and WeChat Mini Program authentication.

## Features

- User registration and login with username/password
- JWT-based authentication with access and refresh tokens
- WeChat Mini Program authentication
- Token refresh and revocation
- Protected API endpoints

## WeChat Mini Program Authentication Flow

1. The Mini Program client calls `wx.login()` to get a temporary code
2. The client sends this code to the server's `/wechat/login` endpoint
3. The server exchanges the code for an OpenID by calling WeChat's API
4. The server creates or retrieves a user account associated with this OpenID
5. The server generates JWT tokens and returns them to the client
6. The client can use these tokens to access protected API endpoints

## API Endpoints

### Public Endpoints

- `POST /register` - Register a new user with username/password
- `POST /login` - Login with username/password
- `POST /wechat/login` - Login with WeChat Mini Program code
- `POST /refresh` - Refresh an access token using a refresh token
- `POST /logout` - Logout (revoke a refresh token)
- `GET /health` - Health check endpoint

### Protected Endpoints

- `GET /me` - Get the current user's information
- `GET /api/protected` - Example protected endpoint

## Request/Response Examples

### WeChat Mini Program Login

Request:
```json
POST /wechat/login
{
  "code": "temporary_code_from_wx_login"
}
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900
}
```

## Configuration

Update the WeChat Mini Program configuration in `config/config.go`:

```go
WeChat: WeChatConfig{
    AppID:     "your-wechat-appid",
    AppSecret: "your-wechat-appsecret",
},
```

## Running the Server

1. Make sure Redis is running
2. Run the server: `go run main.go`
3. The server will start on port 8080 (configurable in `config/config.go`)
