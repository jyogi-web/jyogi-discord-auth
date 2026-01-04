# API Reference

## Authentication (Auth)

### Discord Login

- **URL**: `/auth/login`
- **Method**: `GET`
- **Description**: Initiates the Discord OAuth2 authentication flow.

### Discord Callback

- **URL**: `/auth/callback`
- **Method**: `GET`
- **Description**: Receives the redirect from Discord and creates a session.

### Logout

- **URL**: `/auth/logout`
- **Method**: `POST`
- **Description**: Destroys the session and logs out.

## Token

### Issue JWT

- **URL**: `/token`
- **Method**: `POST`
- **Header**: `Cookie: session=...`
- **Description**: Issues a JWT for a user with a valid session.

### Refresh Token

- **URL**: `/token/refresh`
- **Method**: `POST`
- **Description**: Retrieves a new access token using a refresh token.

## Protected Resources

The following endpoints require a valid JWT.

### Verify JWT

- **URL**: `/api/verify`
- **Method**: `GET`
- **Header**: `Authorization: Bearer <token>`
- **Description**: Verifies the validity of the token.

### Get User Info

- **URL**: `/api/user`
- **Method**: `GET`
- **Header**: `Authorization: Bearer <token>`
- **Description**: Returns information about the logged-in user.
