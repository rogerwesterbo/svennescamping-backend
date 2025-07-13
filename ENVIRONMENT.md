# Environment Configuration

This project uses a layered environment configuration system that supports both development and production environments.

## Environment Files

### `.env` (Production/Default)

- **Committed to version control**
- Contains production/default configuration values
- Shared across all environments
- Used as the base configuration

### `.env.local` (Development)

- **NOT committed to version control** (listed in `.gitignore`)
- Contains local development overrides
- Overrides values from `.env`
- Each developer can have their own local configuration

## Loading Order (Priority)

1. **Defaults** - Hardcoded defaults in `internal/settings/settings.go`
2. **`.env`** - Production/base configuration
3. **`.env.local`** - Local development overrides
4. **Process Environment Variables** - Highest priority (runtime environment)

## Example Usage

### Development Setup

1. Copy `.env` to `.env.local`
2. Modify `.env.local` with your local development settings
3. The API will automatically use your local settings

### Production Deployment

- Set environment variables directly in your deployment environment
- Or modify `.env` for production defaults
- `.env.local` is ignored in production (not deployed)

## Available Configuration

| Variable            | Description                                | Example                                        |
| ------------------- | ------------------------------------------ | ---------------------------------------------- |
| `DEVELOPMENT`       | Enable development mode                    | `true` or `false`                              |
| `CORS_ORIGINS`      | Allowed CORS origins (semicolon-separated) | `http://localhost:5173;https://yourdomain.com` |
| `STRIPE_APIKEY`     | Stripe API key                             | `sk_test_...` or `sk_live_...`                 |
| `STRIPE_WEBHOOKKEY` | Stripe webhook secret                      | `whsec_...`                                    |
| `STRIPE_WEBHOOKURL` | Stripe webhook URL                         | `https://yourdomain.com/webhook`               |
| `STRIPE_APIURL`     | Stripe API URL                             | `https://api.stripe.com`                       |
| `STRIPE_APIVERSION` | Stripe API version                         | `2020-08-27`                                   |

## CORS Configuration

The `CORS_ORIGINS` variable accepts multiple origins separated by semicolons:

```bash
# Single origin
CORS_ORIGINS=http://localhost:5173

# Multiple origins
CORS_ORIGINS=http://localhost:5173;http://localhost:3000;https://yourdomain.com
```

## User Role Configuration

### Admin Emails (`ADMIN_EMAILS`)

Comma-separated list of email addresses that should have admin access.

**Example:**

```bash
ADMIN_EMAILS=admin@svennescamping.no,manager@svennescamping.no,owner@svennescamping.no
```

### User Emails (`USER_EMAILS`)

Comma-separated list of email addresses that should have user-level access.

**Example:**

```bash
USER_EMAILS=employee1@svennescamping.no,employee2@svennescamping.no,contractor@example.com
```

### Role Assignment Priority

The system assigns roles in the following order of priority:

1. **Admin List** - If email is in `ADMIN_EMAILS` → `admin` role
2. **User List** - If email is in `USER_EMAILS` → `user` role
3. **Manual Assignment** - If role manually set via API → assigned role
4. **OAuth Groups** - If user has groups in token:
   - `admin`/`administrators` → `admin` role
   - `user`/`users` → `user` role
   - `no_access`/`noaccess` → `no_access` role
5. **Domain Rules** - If email ends with `@svennescamping.no` → `user` role
6. **Verified Email** - If email is verified → `user` role
7. **Default** - Unverified or unknown → `no_access` role

## Security Notes

- Never commit sensitive keys to `.env`
- Use test keys in `.env` and override with production keys via environment variables
- `.env.local` is automatically ignored by Git for security
- Never commit `.env.local` to version control
- Use strong, unique email addresses for admin access
- Regularly review and audit user role assignments
