# LinkedIn Profile API

A Go API that reverse engineers LinkedIn's internal Voyager API to return structured profile JSON.

## How It Works

LinkedIn's web app communicates with an undocumented internal API at `/voyager/api`. This service makes authenticated requests to that API using session cookies, fetches profile data from **4 endpoints concurrently** (using goroutines + `errgroup`), and returns a clean structured JSON response.

### Concurrent fetch strategy

```
GET /profileView  ─┐
GET /skills       ─┤─ all fire at the same time via goroutines
GET /certifications─┤
GET /languages    ─┘
                   └─► merge → return JSON
```

## Setup

### Prerequisites
- Go 1.21+
- A LinkedIn account (for cookies)

### 1. Clone the repo
```bash
git clone https://github.com/yourusername/linkedin-profile-api
cd linkedin-profile-api
```

### 2. Get your LinkedIn cookies

1. Log into LinkedIn in Chrome
2. Open DevTools (F12) → **Application** → **Cookies** → `https://www.linkedin.com`
3. Copy the values for:
   - `li_at`
   - `JSESSIONID`

### 3. Configure environment
```bash
cp .env.example .env
# Edit .env and paste your cookie values
```

### 4. Install dependencies and run
```bash
go mod tidy
go run main.go
```

Server starts on `http://localhost:8080`.

## API

### `GET /profile?url=<linkedin_url>`

**Example:**
```bash
curl "http://localhost:8080/profile?url=https://www.linkedin.com/in/williamhgates"
```

**Response schema:**
```json
{
  "username": "williamhgates",
  "first_name": "Bill",
  "last_name": "Gates",
  "headline": "Co-chair, Bill & Melinda Gates Foundation",
  "location": "Seattle, Washington",
  "about": "...",
  "profile_image_url": "https://...",
  "experience": [
    {
      "title": "Co-chair",
      "company_name": "Bill & Melinda Gates Foundation",
      "location": "Seattle, WA",
      "start_date": "01/2000",
      "end_date": "",
      "description": "..."
    }
  ],
  "education": [
    {
      "school_name": "Harvard University",
      "degree": "",
      "field_of_study": "",
      "start_date": "1973",
      "end_date": "1975"
    }
  ],
  "skills": [{ "name": "Strategy" }],
  "certifications": [],
  "languages": [{ "name": "English", "proficiency": "NATIVE_OR_BILINGUAL" }]
}
```

### `GET /health`
Returns `{"status":"ok"}` — use for deployment health checks.

## Deploy to Railway

1. Push this repo to GitHub
2. Go to [railway.app](https://railway.app) → New Project → Deploy from GitHub
3. Add environment variables in Railway dashboard:
   - `LI_AT` = your cookie value
   - `JSESSIONID` = your cookie value
4. Railway auto-detects Go, builds, and deploys. You get a public HTTPS URL.

## Known Limitations

- **Cookie expiry**: `li_at` cookies expire periodically. You'll need to refresh them when you get 401/403 errors.
- **Rate limiting**: LinkedIn rate-limits aggressively. Avoid rapid repeated requests.
- **Private profiles**: Returns limited data for profiles you're not connected to.
- **LinkedIn anti-scraping**: LinkedIn may block requests; the `User-Agent` and headers mimic a real browser but are not guaranteed.
- **Schema changes**: LinkedIn's internal API is undocumented and can change without notice.
