package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/yourusername/linkedin-profile-api/models"
)

// FetchProfile fetches profile information using LinkedIn's Dash profile endpoint.
func (c *Client) FetchProfile(username string) (*models.Profile, error) {
	dashURL := fmt.Sprintf("%s/identity/dash/profiles?q=memberIdentity&memberIdentity=%s&decorationId=com.linkedin.voyager.dash.deco.identity.profile.FullProfileWithEntities-87", baseURL, username)

	data, err := c.fetchJSON(dashURL)
	if err != nil {
		return nil, err
	}

	profile := &models.Profile{
		Username:       username,
		Experience:     make([]models.Experience, 0),
		Education:      make([]models.Education, 0),
		Skills:         make([]models.Skill, 0),
		Certifications: make([]models.Certification, 0),
		Languages:      make([]models.Language, 0),
	}

	parseDashProfile(profile, data)

	if profile.FirstName == "" && profile.LastName == "" && len(profile.Experience) == 0 {
		return nil, fmt.Errorf("profile not found")
	}

	return profile, nil
}

func (c *Client) fetchJSON(url string) (map[string]interface{}, error) {
	req, err := c.newRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusSeeOther {
		return nil, fmt.Errorf("authentication failed (status %d): check LI_AT and JSESSIONID cookies", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("profile not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response from %s: %w", url, err)
	}

	return result, nil
}

// --- Parsers ---

func parseDashProfile(p *models.Profile, data map[string]interface{}) {
	if data == nil {
		return
	}

	included, ok := data["included"].([]interface{})
	if !ok {
		return
	}

	for _, item := range included {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		typeStr, _ := obj["$type"].(string)
		entityUrn, _ := obj["entityUrn"].(string)

		switch {
		case strings.Contains(typeStr, "Profile") || strings.Contains(entityUrn, "fsd_profile:"):
			if fn, ok := obj["firstName"].(string); ok && fn != "" {
				p.FirstName = fn
			}
			if ln, ok := obj["lastName"].(string); ok && ln != "" {
				p.LastName = ln
			}
			if hl, ok := obj["headline"].(string); ok && hl != "" {
				p.Headline = hl
			}
			if summary, ok := obj["summary"].(string); ok && summary != "" {
				p.About = summary
			}
			if loc, ok := obj["locationName"].(string); ok && loc != "" {
				p.Location = loc
			} else if loc, ok := obj["geoLocationName"].(string); ok && loc != "" {
				p.Location = loc
			}
			if photos, ok := obj["profilePicture"].(map[string]interface{}); ok {
				p.ProfileImageURL = extractPhoto(photos)
			}

		case strings.Contains(typeStr, "Position"):
			exp := models.Experience{}
			exp.Title, _ = obj["title"].(string)
			if company, ok := obj["companyName"].(string); ok {
				exp.CompanyName = company
			}
			if loc, ok := obj["locationName"].(string); ok {
				exp.Location = loc
			} else if loc, ok := obj["geoLocationName"].(string); ok {
				exp.Location = loc
			}
			if desc, ok := obj["description"].(string); ok {
				exp.Description = desc
			}
			exp.StartDate = extractDate(obj, "dateRange", "start")
			if exp.StartDate == "" {
				exp.StartDate = extractDate(obj, "timePeriod", "startDate")
			}
			exp.EndDate = extractDate(obj, "dateRange", "end")
			if exp.EndDate == "" {
				exp.EndDate = extractDate(obj, "timePeriod", "endDate")
			}
			if exp.Title != "" || exp.CompanyName != "" {
				p.Experience = append(p.Experience, exp)
			}

		case strings.Contains(typeStr, "Education"):
			edu := models.Education{}
			edu.SchoolName, _ = obj["schoolName"].(string)
			edu.Degree, _ = obj["degreeName"].(string)
			edu.FieldOfStudy, _ = obj["fieldOfStudy"].(string)
			edu.StartDate = extractDate(obj, "dateRange", "start")
			if edu.StartDate == "" {
				edu.StartDate = extractDate(obj, "timePeriod", "startDate")
			}
			edu.EndDate = extractDate(obj, "dateRange", "end")
			if edu.EndDate == "" {
				edu.EndDate = extractDate(obj, "timePeriod", "endDate")
			}
			if edu.SchoolName != "" {
				p.Education = append(p.Education, edu)
			}

		case strings.Contains(typeStr, "Skill"):
			if name, ok := obj["name"].(string); ok && name != "" {
				p.Skills = append(p.Skills, models.Skill{Name: name})
			}

		case strings.Contains(typeStr, "Certification"):
			cert := models.Certification{}
			cert.Name, _ = obj["name"].(string)
			if auth, ok := obj["authority"].(string); ok {
				cert.Authority = auth
			} else if auth, ok := obj["issuingAuthorityName"].(string); ok {
				cert.Authority = auth
			}
			cert.LicenseNumber, _ = obj["licenseNumber"].(string)
			cert.StartDate = extractDate(obj, "dateRange", "start")
			if cert.StartDate == "" {
				cert.StartDate = extractDate(obj, "timePeriod", "startDate")
			}
			cert.EndDate = extractDate(obj, "dateRange", "end")
			if cert.EndDate == "" {
				cert.EndDate = extractDate(obj, "timePeriod", "endDate")
			}
			if cert.Name != "" {
				p.Certifications = append(p.Certifications, cert)
			}

		case strings.Contains(typeStr, "Language"):
			lang := models.Language{}
			lang.Name, _ = obj["name"].(string)
			lang.Proficiency, _ = obj["proficiency"].(string)
			if lang.Name != "" {
				p.Languages = append(p.Languages, lang)
			}
		}
	}
}

// --- Helpers ---

func extractDate(obj map[string]interface{}, keys ...string) string {
	current := obj
	for i, key := range keys {
		val, ok := current[key]
		if !ok {
			return ""
		}
		if i == len(keys)-1 {
			dateObj, ok := val.(map[string]interface{})
			if !ok {
				return ""
			}
			month, _ := dateObj["month"].(float64)
			year, _ := dateObj["year"].(float64)
			if year == 0 {
				return ""
			}
			if month > 0 {
				return fmt.Sprintf("%02d/%d", int(month), int(year))
			}
			return fmt.Sprintf("%d", int(year))
		}
		nested, ok := val.(map[string]interface{})
		if !ok {
			return ""
		}
		current = nested
	}
	return ""
}

func extractPhoto(photos map[string]interface{}) string {
	displayImage, ok := photos["displayImageReference"].(map[string]interface{})
	if !ok {
		return ""
	}
	vectorImage, ok := displayImage["vectorImage"].(map[string]interface{})
	if !ok {
		return ""
	}
	artifacts, ok := vectorImage["artifacts"].([]interface{})
	if !ok || len(artifacts) == 0 {
		return ""
	}
	rootURL, _ := vectorImage["rootUrl"].(string)
	last, ok := artifacts[len(artifacts)-1].(map[string]interface{})
	if !ok {
		return ""
	}
	fileID, _ := last["fileIdentifyingUrlPathSegment"].(string)
	return rootURL + fileID
}
