# Implementation Guide

Step-by-step guide to build this project from scratch.

---

## Step 1 — Project Setup

```bash
mkdir linkedin-profile-api && cd linkedin-profile-api
go mod init github.com/yourusername/linkedin-profile-api
go get github.com/go-chi/chi/v5
go get github.com/joho/godotenv
go get golang.org/x/sync
go get golang.org/x/time
go get github.com/patrickmn/go-cache
```

Create folders:

```bash
mkdir models scraper handlers cache middleware
```

---

## Step 2 — Models (`models/profile.go`)

Define your structs for the JSON response:

- `Profile` — top level (name, headline, location, about, photo URL)
- `Experience` — title, company, location, start/end date, description
- `Education` — school, degree, field, start/end date
- `Skill` — just name
- `Certification` — name, authority, license number, start/end date
- `Language` — name, proficiency

All fields get `json:"..."` tags.

---

## Step 3 — Scraper Client (`scraper/client.go`)

Create a struct with:
- `*http.Client` with a 15s timeout
- `liAt` and `jsessionid` strings loaded from `os.Getenv`

Write a `NewClient()` function that errors if either env var is missing.

Write a `newRequest(url string)` method that sets these headers on every request:

```
Cookie: li_at=...; JSESSIONID="..."
Csrf-Token: <jsessionid value>
User-Agent: Mozilla/5.0 ...
Accept: application/vnd.linkedin.normalized+json+2.1
X-RestLi-Protocol-Version: 2.0.0
X-Li-Lang: en_US
```

---

## Step 4 — Explore Voyager API in DevTools (do this before coding the parser)

1. Log into LinkedIn in Chrome
2. DevTools → Network → filter by `voyager`
3. Visit a profile and look at these 4 requests:
   - `/voyager/api/identity/profiles/{username}/profileView`
   - `/voyager/api/identity/profiles/{username}/skills`
   - `/voyager/api/identity/profiles/{username}/certifications`
   - `/voyager/api/identity/profiles/{username}/languages`
4. Copy the raw JSON responses — study the shape before writing parsers

---

## Step 5 — Fetch + Parse (`scraper/linkedin.go`)

**5a — `fetchJSON(url string)`**
Makes the HTTP call, reads body, unmarshals to `map[string]interface{}`. Handle 401/403/404 explicitly.

**5b — Parsers** — one function per endpoint:
- `parseProfileView` — loop through `data["included"]`, check `$type` field, extract fields based on whether it's a `Profile`, `Position`, or `Education` type
- `parseSkills` — loop through `data["elements"]`, grab `name`
- `parseCertifications` — loop through `data["elements"]`
- `parseLanguages` — loop through `data["elements"]`

Helper functions you'll need:
- `extractDate(obj, "timePeriod", "startDate")` — digs into nested maps to get month/year
- `extractPhoto(photos map)` — finds the largest artifact URL

**5c — `FetchProfile(username string)`** — this is where goroutines go:

```go
g := new(errgroup.Group)

g.Go(func() error { profileView, err = fetchJSON(...profileView URL); return err })
g.Go(func() error { skillsData, err = fetchJSON(...skills URL); return err })
g.Go(func() error { certsData, err = fetchJSON(...certs URL); return err })
g.Go(func() error { langsData, err = fetchJSON(...languages URL); return err })

if err := g.Wait(); err != nil { return nil, err }

// then call all 4 parsers and return merged Profile
```

---

## Step 6 — Cache (`cache/cache.go`)

- Create a `ProfileCache` struct wrapping `*gocache.Cache`
- `NewProfileCache()` — init with 30min TTL, 10min cleanup interval
- `Get(username string)` — returns `(*models.Profile, bool)`
- `Set(username string, profile *models.Profile)` — stores with TTL

---

## Step 7 — Rate Limiter (`middleware/ratelimiter.go`)

- Create a `RateLimiter` struct with a `map[string]*rate.Limiter` (one per IP) and a `sync.Mutex`
- `getLimiter(ip string)` — gets or creates a limiter for that IP (5 req/s, burst 10)
- `Middleware` — standard `http.Handler` wrapper, call `limiter.Allow()`, return 429 if false

---

## Step 8 — Handler (`handlers/profile.go`)

`GetProfile(w, r)`:
1. Read `url` query param — 400 if missing
2. Call `extractUsername(rawURL)` — parse the `/in/slug` part
3. Check cache — if hit, set `X-Cache: HIT` header and return
4. Call `client.FetchProfile(username)`
5. On error — map to correct HTTP status (404/401/500)
6. On success — `cache.Set(...)`, set `X-Cache: MISS`, return JSON

`extractUsername(rawURL)`:
- Prepend `https://` if missing
- `url.Parse(...)`
- Split path, check `parts[0] == "in"` and `parts[1] != ""`

---

## Step 9 — Main (`main.go`)

1. `godotenv.Load()` — loads `.env` locally
2. `scraper.NewClient()` — errors out if cookies missing
3. `cache.NewProfileCache()`
4. `handlers.NewProfileHandler(client, cache)`
5. `middleware.NewRateLimiter()`
6. Wire chi router — attach rate limiter, logger, recoverer middleware
7. Register `GET /health` and `GET /profile`
8. `http.ListenAndServe(":"+port, r)`

---

## Step 10 — `.env.example` + `.gitignore`

`.env.example`:
```
LI_AT=your_li_at_here
JSESSIONID=your_jsessionid_here
```

`.gitignore`:
```
.env
linkedin-profile-api
```

---

## Step 11 — Test Locally

```bash
cp .env.example .env
# fill in real cookies from Chrome DevTools
go run main.go
curl "http://localhost:8080/profile?url=https://linkedin.com/in/williamhgates"
```

---

## Step 12 — Deploy to Railway

1. Push to GitHub
2. Go to railway.app → New Project → Deploy from GitHub repo
3. Add `LI_AT` and `JSESSIONID` as environment variables in the Railway dashboard
4. Get your public HTTPS URL
5. Submit at https://tally.so/r/KYK6qg

---

## Tips

- **Do Step 4 before Step 5** — spend real time in DevTools understanding the raw Voyager JSON before writing any parsers. It will save a lot of debugging.
- The trickiest part is `parseProfileView` — the `included` array mixes many different `$type` objects, you have to filter by type string.
- If you get 401/403 errors, your cookies have expired — go get fresh ones from Chrome.
- The fallback implementation is in this same folder if you run out of time.
