# LinkedIn Profile API

A Go API that reverse engineers LinkedIn's internal Voyager API to return structured profile JSON.

## 🚀 Live API

**Base URL:** `https://linkedin-profile-api-production-c6ab.up.railway.app`

**Try it:**
```bash
curl "https://linkedin-profile-api-production-c6ab.up.railway.app/profile?url=https://www.linkedin.com/in/williamhgates"
```

## ⚡ Pre-warmed Profiles (Instant Response)

These profiles are already cached and will respond instantly:

| Profile | URL |
|---------|-----|
| Bill Gates | `https://linkedin-profile-api-production-c6ab.up.railway.app/profile?url=https://www.linkedin.com/in/williamhgates` |
| Sundar Pichai | `https://linkedin-profile-api-production-c6ab.up.railway.app/profile?url=https://www.linkedin.com/in/sundarpichai` |
| Satya Nadella | `https://linkedin-profile-api-production-c6ab.up.railway.app/profile?url=https://www.linkedin.com/in/satyanadella` |

> **Note:** LinkedIn aggressively rate-limits requests from datacenter IPs (like Railway). Pre-cached profiles always return instantly. New profiles may occasionally return a 401 if LinkedIn blocks the session — this is a known LinkedIn anti-scraping measure, not a bug in the API.

## How It Works

LinkedIn's web app communicates with an undocumented internal API at `/voyager/api`. This service makes authenticated requests to that API using session cookies and returns a clean structured JSON response.

The API uses LinkedIn's `dash/profiles` endpoint with the `FullProfileWithEntities` decoration to fetch all profile data (name, headline, experience, education, skills, certifications, languages, profile image) in a single request.

**Features:**
- ✅ In-memory caching (5 min TTL) to avoid redundant LinkedIn requests
- ✅ Rate limiting (5 req/s per IP, burst of 10)
- ✅ Deployed on Railway with HTTPS

## API Documentation

### `GET /profile?url=<linkedin_url>`

Fetches a LinkedIn profile by URL.

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | ✅ | Full LinkedIn profile URL (e.g. `https://www.linkedin.com/in/username`) |

**Example:**
```bash
curl "https://linkedin-profile-api-production-c6ab.up.railway.app/profile?url=https://www.linkedin.com/in/williamhgates"
```

**Response schema:**
```json
{
  "username": "williamhgates",
  "first_name": "Bill",
  "last_name": "Gates",
  "headline": "Chair, Gates Foundation and Founder, Breakthrough Energy",
  "location": "Seattle, Washington",
  "about": "Co-founder of Microsoft...",
  "profile_image_url": "https://media.licdn.com/...",
  "experience": [
    {
      "title": "Co-chair",
      "company_name": "Gates Foundation",
      "location": "",
      "start_date": "2000",
      "end_date": "",
      "description": ""
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
  "skills": [],
  "certifications": [],
  "languages": []
}
```

**Error responses:**
| Status | Meaning |
|--------|---------|
| `400` | Missing or invalid URL |
| `401` | LinkedIn session expired |
| `404` | Profile not found |
| `429` | Rate limit exceeded |
| `500` | Unexpected error |

### `GET /health`

Returns `{"status":"ok"}` — used for deployment health checks.

## Local Setup

### Prerequisites
- Go 1.21+
- A LinkedIn account (for cookies)

### 1. Clone the repo
```bash
git clone https://github.com/Chirag-Main1/linkedin-profile-api
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

## Deploy to Railway

1. Push this repo to GitHub
2. Go to [railway.app](https://railway.app) → New Project → Deploy from GitHub
3. Add environment variables in Railway dashboard:
   - `LI_AT` = your cookie value
   - `JSESSIONID` = your cookie value
4. Generate a domain under Settings → Networking → port `8080`

## Known Limitations

- **Cookie expiry**: `li_at` cookies expire periodically. You'll need to refresh them when you get 401/403 errors.
- **Rate limiting**: LinkedIn rate-limits aggressively. Avoid rapid repeated requests to the same session.
- **Private profiles**: Returns limited data for profiles you're not connected to.
- **LinkedIn anti-scraping**: LinkedIn may block requests; the `User-Agent` and headers mimic a real browser but are not guaranteed.
- **Schema changes**: LinkedIn's internal API is undocumented and can change without notice.
- **Skills**: LinkedIn's Voyager API sometimes returns skills in a separate paginated endpoint; some profiles may show empty skills.
