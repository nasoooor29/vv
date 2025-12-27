# Authentication

Visory supports multiple authentication methods including username/password and OAuth providers (Google and GitHub).

## Authentication Methods

### Username/Password

Users can register and login using traditional credentials.

#### Registration

Navigate to `/auth/register` and provide:
- Username (unique)
- Email (unique)
- Password

New users are assigned a default role with limited permissions.

#### Login

Navigate to `/auth/login` and enter your credentials. On successful login, a session cookie is set.

### OAuth Authentication

Visory supports OAuth login with Google and GitHub. OAuth users are automatically registered on first login.

## OAuth Setup

### Environment Variables

Add these to your `.env` file:

```bash
# Google OAuth
GOOGLE_OAUTH_KEY="your-google-client-id"
GOOGLE_OAUTH_SECRET="your-google-client-secret"

# GitHub OAuth (optional)
GITHUB_OAUTH_KEY="your-github-client-id"
GITHUB_OAUTH_SECRET="your-github-client-secret"

# Session Secret (required for cookie encryption)
SESSION_SECRET="your-random-secret-key"
```

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable the **Google+ API**
4. Navigate to **Credentials** > **Create OAuth 2.0 credentials**
5. Choose **Web application**
6. Add authorized redirect URIs:
   - Development: `http://localhost:9999/api/auth/oauth/callback/google`
   - Production: `https://yourdomain.com/api/auth/oauth/callback/google`
7. Copy the Client ID and Client Secret to your `.env`

### GitHub OAuth Setup

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click **New OAuth App**
3. Fill in:
   - Application name
   - Homepage URL
   - Authorization callback URL:
     - Development: `http://localhost:9999/api/auth/oauth/callback/github`
     - Production: `https://yourdomain.com/api/auth/oauth/callback/github`
4. Copy the Client ID and Client Secret to your `.env`

## Session Management

Sessions are stored in the database with the following properties:

- Session tokens are unique per user session
- Cookies are HTTP-only and secure in production
- Sessions can be invalidated via logout

### Session Cookie

| Property | Value |
|----------|-------|
| Name | `token` |
| HTTP Only | Yes |
| Secure | Yes (production) |
| SameSite | Lax |

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/auth/register` | POST | Register new user |
| `/api/auth/login` | POST | Login with credentials |
| `/api/auth/logout` | POST | Logout and invalidate session |
| `/api/auth/me` | GET | Get current user info |
| `/api/auth/oauth/google` | GET | Initiate Google OAuth |
| `/api/auth/oauth/github` | GET | Initiate GitHub OAuth |
| `/api/auth/oauth/callback/:provider` | GET | OAuth callback handler |

## Frontend Integration

Both login and register pages include OAuth buttons that work automatically when configured:

- "Sign in with Google" / "Sign up with Google"
- "Sign in with GitHub" / "Sign up with GitHub"

The OAuth buttons redirect to the respective provider and handle the callback automatically.

## Troubleshooting

### "no SESSION_SECRET environment variable is set"

Generate and set a session secret:

```bash
# Generate secret
openssl rand -base64 32

# Add to .env
SESSION_SECRET=your-generated-secret
```

### "OAuth provider not configured"

- Verify `.env` file exists with correct variable names
- Ensure `GOOGLE_OAUTH_KEY` and `GOOGLE_OAUTH_SECRET` are set
- Restart the application after updating `.env`

### OAuth callback errors

- Check that callback URLs match exactly in provider settings
- Ensure the provider API is enabled (Google+ API for Google)
- Verify redirect URI includes the correct port and path
