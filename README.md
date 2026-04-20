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

- Go 1.22+
- Docker + Docker Compose
- A Google Cloud OAuth 2.0 credential

### 1. Clone and configure

```bash
git clone https://github.com/yourorg/reploy.git
cd reploy
cp .env.example .env
# Edit .env with your Google OAuth credentials and DB password
```

### 2. Start Supabase (self-hosted)

```bash
git clone --depth 1 https://github.com/supabase/supabase
cd supabase/docker
cp .env.example .env
# Run the key generator and fill in .env
bash utils/generate-keys.sh
# Expose the DB port by adding to the db service in docker-compose.yml:
#   ports:
#     - "5433:5432"
docker compose up -d
cd -
```

### 3. Run the database migration

```bash
docker exec -i supabase-db psql -U postgres -d postgres < migrations/001_init.sql
```

### 4. Promote yourself to teacher (after first login)

```bash
docker exec -i supabase-db psql -U postgres -d postgres \
  -c "UPDATE users SET role='teacher' WHERE email='yourname@school.pyc.edu.hk';"
```

### 5. Start the server

```bash
go run ./cmd/server
# or build first:
go build -o reploy ./cmd/server && ./reploy
```

Open http://localhost:3000

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
