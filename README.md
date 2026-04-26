# Reploy

A platform for high school ICT students to rapidly deploy LLM-generated HTML apps. Students sign in with their school Google account, paste or write a single HTML file, and get a shareable public link instantly.

## Features

- Google OAuth login restricted to `@school.pyc.edu.hk`
- In-browser HTML editor (CodeMirror) with auto-save
- Public preview links: `https://reploy.spyc.hk/<pyccode>/<app-slug>`
- Teacher admin panel: ban/unban students, hide/show apps, edit student code
- Self-hosted Supabase (PostgreSQL) as backend

## Tech Stack

- **Backend**: Go + Gin
- **Frontend**: Vanilla JS + Tailwind CSS (CDN) + CodeMirror 5 (CDN)
- **Database**: Supabase self-hosted (PostgreSQL)
- **Reverse proxy**: Caddy (auto HTTPS)

---

## Local Development

### Prerequisites

- Go 1.25+
- Docker + Docker Compose
- [air](https://github.com/air-verse/air) for hot reload: `go install github.com/air-verse/air@latest`
- A Google Cloud OAuth 2.0 credential

### 1. Clone and configure

```bash
git clone https://github.com/kingkenleung/reploy.git
cd reploy
cp .env.example .env
# Edit .env — set DATABASE_URL to point at your local supabase-db (see step 2)
```

### 2. Start Supabase locally

```bash
# Clone supabase into a sibling directory (one-time setup)
git clone --depth 1 https://github.com/supabase/supabase ~/supabase-local
cd ~/supabase-local/docker
cp .env.example .env
# Expose the DB on port 5433 — add this under the `db` service in docker-compose.yml:
#   ports:
#     - "5433:5432"
docker compose up -d
```

Set `DATABASE_URL` in `.env`:
```
DATABASE_URL=postgresql://postgres:<your-supabase-db-password>@localhost:5433/postgres?sslmode=disable
```

### 3. Run migrations

```bash
docker exec -i supabase-db psql -U postgres < migrations/001_init.sql
docker exec -i supabase-db psql -U postgres -c "ALTER TABLE apps ADD COLUMN IF NOT EXISTS category JSONB NOT NULL DEFAULT '[]';"
docker exec -i supabase-db psql -U postgres -c "ALTER TABLE apps ADD COLUMN IF NOT EXISTS approved BOOLEAN NOT NULL DEFAULT false;"
```

### 4. Start the server with hot reload

```bash
air
```

`air` watches for `.go` file changes and automatically rebuilds and restarts the server. Open http://localhost:3000.

If you don't want hot reload:
```bash
go run ./cmd/server
```

### 5. Promote yourself to teacher (after first login)

```bash
docker exec -i supabase-db psql -U postgres \
  -c "UPDATE users SET role='teacher' WHERE email='yourname@school.pyc.edu.hk';"
```

---

## Production Deployment (Ubuntu 24.04 LTS)

### 1. Install dependencies

```bash
sudo apt update && sudo apt install -y docker.io docker-compose-v2 golang-go
# Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install caddy
```

### 2. Build the binary

```bash
go build -o reploy ./cmd/server
```

### 3. Configure environment

```bash
cp .env.example .env
# Edit .env — set production GOOGLE_REDIRECT_URL to https://reploy.spyc.hk/auth/google/callback
```

### 4. Set up Caddy

```
# /etc/caddy/Caddyfile
reploy.spyc.hk {
    reverse_proxy localhost:3000
}
```

```bash
sudo systemctl reload caddy
```

### 5. Run as a systemd service

```ini
# /etc/systemd/system/reploy.service
[Unit]
Description=Reploy
After=network.target

[Service]
WorkingDirectory=/opt/reploy
ExecStart=/opt/reploy/reploy
EnvironmentFile=/opt/reploy/.env
Restart=always
User=www-data

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now reploy
```

---

## Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/) → APIs & Services → Credentials
2. Create an OAuth 2.0 Client ID (Web application)
3. Add Authorised redirect URIs:
   - `http://localhost:3000/auth/google/callback` (dev)
   - `https://reploy.spyc.hk/auth/google/callback` (prod)
4. Copy Client ID and Secret into `.env`

---

## Environment Variables

| Variable | Description |
|---|---|
| `PORT` | Server port (default: 3000) |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL |
| `ALLOWED_EMAIL_DOMAIN` | Restrict login to this domain (e.g. `school.pyc.edu.hk`) |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret for signing JWT tokens (min 32 chars) |
