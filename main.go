package main

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googlecalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// ---- Timing and config constants (overridden in testmode builds via testmode.go) ----

var errInvalidCredentials = errors.New("INVALID_LOGIN: invalid username, password, or security token")

var (
	networkWaitDuration   = 5 * time.Minute
	sfRetryShort          = 5 * time.Minute
	sfRetryLong           = 1 * time.Hour
	syncRetryDuration     = 5 * time.Minute
	schedulerTickDuration = 5 * time.Minute
	syncInterval          = 24 * time.Hour
	listenAddr            = ":5001"
	appBaseURL            = "http://localhost:5001"
)

// ---- Cross-platform notifications ----

func notify(title, message string) {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, message, title)
		exec.Command("osascript", "-e", script).Run()
	case "windows":
		// PowerShell toast notification
		ps := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
$template = [Windows.UI.Notifications.ToastTemplateType]::ToastText02
$xml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent($template)
$xml.GetElementsByTagName('text')[0].AppendChild($xml.CreateTextNode(%q)) | Out-Null
$xml.GetElementsByTagName('text')[1].AppendChild($xml.CreateTextNode(%q)) | Out-Null
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('CalSync').Show($toast)
`, title, message)
		exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps).Run()
	}
}

// ---- App directory (absolute, so app works from any working dir) ----

func appDir() string {
	// os.Executable can return a symlink when run via launchd — resolve it
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return filepath.Dir(exe)
	}
	return filepath.Dir(resolved)
}

func appPath(rel string) string {
	return filepath.Join(appDir(), rel)
}

// ---- Session store + mutex ----

type Session struct {
	SFInstanceURL string
	SFSessionID   string
	SFUserID      string
	GoogleToken   *oauth2.Token
	OAuthState    string
	CodeVerifier  string
	EventColor    string
}

var (
	session   Session
	sessionMu sync.RWMutex
)

// ---- Credential persistence ----

const (
	googleTokenFile = "google_token.json"
	sfCredsFile     = "sf_credentials.json"
	colorFile       = "color.json"
)

func saveEventColor(color string) {
	data, _ := json.Marshal(map[string]string{"color": color})
	os.WriteFile(appPath(colorFile), data, 0644)
}

func loadEventColor() string {
	data, err := os.ReadFile(appPath(colorFile))
	if err != nil {
		return ""
	}
	var v map[string]string
	if err := json.Unmarshal(data, &v); err != nil {
		return ""
	}
	return v["color"]
}

type sfCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
	Domain   string `json:"domain"`
}

func saveGoogleToken(t *oauth2.Token) {
	data, err := json.Marshal(t)
	if err != nil {
		log.Printf("[creds] failed to marshal Google token: %v", err)
		return
	}
	if err := os.WriteFile(appPath(googleTokenFile), data, 0600); err != nil {
		log.Printf("[creds] failed to save Google token: %v", err)
	}
}

func loadGoogleToken() *oauth2.Token {
	data, err := os.ReadFile(appPath(googleTokenFile))
	if err != nil {
		return nil
	}
	var t oauth2.Token
	if err := json.Unmarshal(data, &t); err != nil {
		return nil
	}
	return &t
}

func deleteGoogleToken() {
	os.Remove(appPath(googleTokenFile))
}

func saveSFCredentials(creds sfCredentials) {
	data, err := json.Marshal(creds)
	if err != nil {
		log.Printf("[creds] failed to marshal SF credentials: %v", err)
		return
	}
	if err := os.WriteFile(appPath(sfCredsFile), data, 0600); err != nil {
		log.Printf("[creds] failed to save SF credentials: %v", err)
	}
}

func loadSFCredentials() *sfCredentials {
	data, err := os.ReadFile(appPath(sfCredsFile))
	if err != nil {
		return nil
	}
	var c sfCredentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil
	}
	return &c
}

func deleteSFCredentials() {
	os.Remove(appPath(sfCredsFile))
}

// loadSavedCredentials restores persisted credentials on startup so the
// scheduler can run without requiring the user to reconnect after a reboot.
func loadSavedCredentials() {
	if t := loadGoogleToken(); t != nil {
		sessionMu.Lock()
		session.GoogleToken = t
		sessionMu.Unlock()
		log.Println("[startup] Google token restored from disk")
		go fetchGoogleColors()
	}

	if creds := loadSFCredentials(); creds != nil {
		if err := sfLogin(creds.Username, creds.Password, creds.Token, creds.Domain); err != nil {
			log.Printf("[startup] SF re-login failed: %v — will retry on next sync", err)
		} else {
			log.Println("[startup] Salesforce session restored")
		}
	}

	if color := loadEventColor(); color != "" {
		sessionMu.Lock()
		session.EventColor = color
		sessionMu.Unlock()
		log.Printf("[startup] event color restored: %s", color)
	}
}

// ---- Google Calendar colors ----

type gcalColor struct {
	ID   string
	Name string
	Hex  string
}

var defaultGcalColors = []gcalColor{
	{"1", "Blueberry", "#4285F4"},
	{"2", "Sage", "#33B679"},
	{"3", "Grape", "#8E24AA"},
	{"4", "Flamingo", "#E67C73"},
	{"5", "Banana", "#F6BF26"},
	{"6", "Tangerine", "#F4511E"},
	{"7", "Peacock", "#039BE5"},
	{"8", "Graphite", "#616161"},
	{"9", "Lavender", "#7986CB"},
	{"10", "Basil", "#0B8043"},
	{"11", "Tomato", "#D50000"},
}

var (
	gcalColors   = defaultGcalColors
	gcalColorsMu sync.RWMutex
)

// knownColorNames maps Google Calendar event color IDs to human-readable names.
var knownColorNames = map[string]string{
	"1": "Blueberry", "2": "Sage", "3": "Grape", "4": "Flamingo",
	"5": "Banana", "6": "Tangerine", "7": "Peacock", "8": "Graphite",
	"9": "Lavender", "10": "Basil", "11": "Tomato",
}

func fetchGoogleColors() {
	sessionMu.RLock()
	token := session.GoogleToken
	sessionMu.RUnlock()
	if token == nil {
		return
	}
	oauthConfig, err := getOAuthConfig()
	if err != nil {
		return
	}
	ctx := context.Background()
	tokenSource := oauthConfig.TokenSource(ctx, token)
	svc, err := googlecalendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return
	}
	resp, err := svc.Colors.Get().Do()
	if err != nil {
		log.Printf("[colors] failed to fetch from Google: %v — using defaults", err)
		return
	}
	var fetched []gcalColor
	for id, c := range resp.Event {
		name, ok := knownColorNames[id]
		if !ok {
			name = c.Background // fall back to hex for unknown future colors
		}
		fetched = append(fetched, gcalColor{ID: id, Name: name, Hex: c.Background})
	}
	if len(fetched) == 0 {
		return
	}
	sort.Slice(fetched, func(i, j int) bool {
		ni, ei := strconv.Atoi(fetched[i].ID)
		nj, ej := strconv.Atoi(fetched[j].ID)
		if ei == nil && ej == nil {
			return ni < nj
		}
		return fetched[i].ID < fetched[j].ID
	})
	gcalColorsMu.Lock()
	gcalColors = fetched
	gcalColorsMu.Unlock()
	log.Printf("[colors] loaded %d event colors from Google", len(fetched))
}

// ---- Sync map ----

const syncMapFile = "sync_map.json"

func loadSyncMap() map[string]string {
	data, err := os.ReadFile(appPath(syncMapFile))
	if err != nil {
		return map[string]string{}
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		log.Printf("[syncmap] failed to parse: %v", err)
		return map[string]string{}
	}
	return m
}

func saveSyncMap(m map[string]string) {
	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("[syncmap] failed to marshal: %v", err)
		return
	}
	if err := os.WriteFile(appPath(syncMapFile), data, 0644); err != nil {
		log.Printf("[syncmap] failed to write: %v", err)
	}
}

// ---- Google OAuth config ----

const googleCredsFile = "google_credentials.json"
const redirectURL = "http://localhost:5001/connect/google/callback"

func getOAuthConfig() (*oauth2.Config, error) {
	data, err := os.ReadFile(appPath(googleCredsFile))
	if err != nil {
		return nil, fmt.Errorf("google_credentials.json not found")
	}
	config, err := google.ConfigFromJSON(data, googlecalendar.CalendarScope)
	if err != nil {
		return nil, err
	}
	config.RedirectURL = redirectURL
	return config, nil
}

// ---- Salesforce ----

type SFRecordType struct {
	Name string `json:"Name"`
}

type SFEventRecord struct {
	ID            string        `json:"Id"`
	Subject       string        `json:"Subject"`
	StartDateTime string        `json:"StartDateTime"`
	EndDateTime   string        `json:"EndDateTime"`
	Description   string        `json:"Description"`
	Location      string        `json:"Location"`
	RecordType    SFRecordType  `json:"RecordType"`
}

type SFQueryResult struct {
	Records []SFEventRecord `json:"records"`
}

func sfQuery(instanceURL, sessionID, query string) (*SFQueryResult, error) {
	endpoint := instanceURL + "query/?q=" + url.QueryEscape(query)
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+sessionID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Session expired — try to re-login automatically using saved credentials
	if resp.StatusCode == 401 {
		log.Println("[sf] session expired, attempting automatic re-login")
		creds := loadSFCredentials()
		if creds != nil {
			if err := sfLogin(creds.Username, creds.Password, creds.Token, creds.Domain); err != nil {
				log.Printf("[sf] auto re-login failed: %v", err)
				sessionMu.Lock()
				session.SFInstanceURL = ""
				session.SFSessionID = ""
				session.SFUserID = ""
				sessionMu.Unlock()
				if errors.Is(err, errInvalidCredentials) {
					notify("CalSync — Action Required", fmt.Sprintf("Salesforce password or security token is incorrect. Open %s to reconnect.", appBaseURL))
					return nil, fmt.Errorf("Salesforce credentials are invalid — will retry at next sync interval")
				}
				return nil, fmt.Errorf("Salesforce session expired and re-login failed — will retry next sync")
			}
			// Re-login succeeded — retry the query with the new session
			log.Println("[sf] auto re-login successful, retrying query")
			sessionMu.RLock()
			instanceURL = session.SFInstanceURL
			sessionID = session.SFSessionID
			sessionMu.RUnlock()
			return sfQuery(instanceURL, sessionID, query)
		}
		sessionMu.Lock()
		session.SFInstanceURL = ""
		session.SFSessionID = ""
		session.SFUserID = ""
		sessionMu.Unlock()
		return nil, fmt.Errorf("Salesforce session expired — please reconnect in the app")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Salesforce query failed (%d): %s", resp.StatusCode, string(body))
	}

	var result SFQueryResult
	err = json.Unmarshal(body, &result)
	return &result, err
}

func sfLogin(username, password, token, domain string) error {
	loginURL := fmt.Sprintf("https://%s.salesforce.com/services/Soap/u/59.0", domain)
	soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:urn="urn:partner.soap.sforce.com">
  <soapenv:Body>
    <urn:login>
      <urn:username>%s</urn:username>
      <urn:password>%s</urn:password>
    </urn:login>
  </soapenv:Body>
</soapenv:Envelope>`, xmlEscape(username), xmlEscape(password+token))

	req, _ := http.NewRequest("POST", loginURL, strings.NewReader(soapBody))
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("SOAPAction", "login")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if strings.Contains(bodyStr, "<faultstring>") {
		start := strings.Index(bodyStr, "<faultstring>") + len("<faultstring>")
		end := strings.Index(bodyStr, "</faultstring>")
		msg := "Salesforce login failed"
		if start > 0 && end > start {
			msg = bodyStr[start:end]
		}
		if strings.Contains(msg, "INVALID_LOGIN") || strings.Contains(msg, "LOGIN_MUST_USE_SECURITY_TOKEN") {
			return errInvalidCredentials
		}
		return fmt.Errorf("%s", msg)
	}

	sessionID := extractXMLTag(bodyStr, "sessionId")
	if sessionID == "" {
		return fmt.Errorf("Could not get session ID from Salesforce")
	}

	serverURL := extractXMLTag(bodyStr, "serverUrl")
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("Could not parse Salesforce server URL")
	}
	instanceURL := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	userID := extractXMLTag(bodyStr, "userId")

	sessionMu.Lock()
	session.SFInstanceURL = instanceURL + "/services/data/v59.0/"
	session.SFSessionID = sessionID
	session.SFUserID = userID
	sessionMu.Unlock()
	return nil
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func extractXMLTag(body, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	start := strings.Index(body, open)
	if start == -1 {
		return ""
	}
	start += len(open)
	end := strings.Index(body[start:], close)
	if end == -1 {
		return ""
	}
	return body[start : start+end]
}

// ---- Rate-limit aware retry ----

const (
	rateLimitDelay = 250 * time.Millisecond
	maxRetries     = 5
)

// isRateLimit returns true only for retriable quota errors, not permission errors.
func isRateLimit(err error) bool {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.Code == 429 {
		return true
	}
	if apiErr.Code == 403 {
		for _, e := range apiErr.Errors {
			if e.Reason == "rateLimitExceeded" || e.Reason == "userRateLimitExceeded" {
				return true
			}
		}
	}
	return false
}

func withBackoff(fn func() error) error {
	delay := time.Second
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		if !isRateLimit(err) || attempt == maxRetries-1 {
			return err
		}
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))
		time.Sleep(delay + jitter)
		delay *= 2
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
	}
	return fmt.Errorf("withBackoff: exhausted %d retries", maxRetries)
}

// ---- Last-sync persistence ----

const lastSyncFile = "last_sync.json"

func loadLastSync() time.Time {
	data, err := os.ReadFile(appPath(lastSyncFile))
	if err != nil {
		return time.Time{}
	}
	var v struct{ LastSync time.Time }
	if err := json.Unmarshal(data, &v); err != nil {
		return time.Time{}
	}
	return v.LastSync
}

func saveLastSync(t time.Time) {
	data, err := json.Marshal(struct{ LastSync time.Time }{t})
	if err != nil {
		log.Printf("[scheduler] failed to marshal last sync time: %v", err)
		return
	}
	if err := os.WriteFile(appPath(lastSyncFile), data, 0644); err != nil {
		log.Printf("[scheduler] failed to save last sync time: %v", err)
	}
}

// todayAt9 returns 9:00am today in local time.
func todayAt9() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
}

// nextDailyAt9 returns the next 9:00am — today if it hasn't passed, otherwise tomorrow.
func nextDailyAt9() time.Time {
	t := todayAt9()
	if time.Now().After(t) {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

// needsCatchUp returns true if the last sync was before today's 9am.
func needsCatchUp() bool {
	last := loadLastSync()
	if last.IsZero() {
		return false
	}
	return last.Before(todayAt9())
}

// startScheduler runs a catch-up sync on startup if needed, then fires every Monday at 9am.
func runScheduledSync(label string) {
	log.Printf("[scheduler] %s — waiting %v for network", label, networkWaitDuration)
	time.Sleep(networkWaitDuration)
	// Re-check after the wait — a manual sync may have run in the meantime
	if last := loadLastSync(); !last.IsZero() {
		alreadyDone := syncInterval < 24*time.Hour && time.Since(last) < syncInterval
		if !alreadyDone && syncInterval >= 24*time.Hour {
			alreadyDone = last.After(todayAt9())
		}
		if alreadyDone {
			log.Println("[scheduler] sync already completed during network wait, skipping")
			return
		}
	}

	sfFailures := 0
	for {
		sessionMu.RLock()
		connected := session.SFInstanceURL != "" && session.GoogleToken != nil
		sessionMu.RUnlock()
		if !connected {
			// SF session may be missing because startup login failed (e.g. no network at boot).
			// Retry sfLogin with saved credentials before waiting again.
			if creds := loadSFCredentials(); creds != nil && session.SFInstanceURL == "" {
				if err := sfLogin(creds.Username, creds.Password, creds.Token, creds.Domain); err != nil {
					if errors.Is(err, errInvalidCredentials) {
						log.Printf("[scheduler] SF re-login failed: invalid credentials — will retry at next sync interval")
						notify("CalSync — Action Required", fmt.Sprintf("Salesforce password or security token is incorrect. Open %s to reconnect.", appBaseURL))
						return
					}
					sfFailures++
					retryIn := sfRetryShort
					if sfFailures > 3 {
						retryIn = sfRetryLong
					}
					log.Printf("[scheduler] SF re-login failed (%d): %v — retrying in %v", sfFailures, err, retryIn)
					switch {
					case sfFailures <= 3:
						notify("CalSync — Action Required", fmt.Sprintf("Salesforce login failed. Open %s to reconnect.", appBaseURL))
					case sfFailures == 4:
						notify("CalSync — Action Required", fmt.Sprintf("Salesforce login is still failing. Retrying hourly. Open %s to reconnect.", appBaseURL))
					case sfFailures == 7:
						notify("CalSync — Action Required", fmt.Sprintf("CalSync has been unable to log in to Salesforce after several attempts. Open %s to reconnect.", appBaseURL))
					}
					time.Sleep(retryIn)
				} else {
					log.Println("[scheduler] SF re-login succeeded")
					sfFailures = 0
					continue
				}
			} else {
				log.Println("[scheduler] accounts not connected, retrying in 5 minutes")
				notify("CalSync — Action Required", fmt.Sprintf("Accounts not connected. Open %s to connect.", appBaseURL))
				time.Sleep(sfRetryShort)
			}
			continue
		}
		if result, err := runSync(); err != nil {
			log.Printf("[scheduler] sync error: %v — retrying in %v", err, syncRetryDuration)
			notify("CalSync — Sync Failed", fmt.Sprintf("Sync error: %v", err))
			time.Sleep(syncRetryDuration)
		} else {
			saveLastSync(time.Now())
			logSyncDetails(result)
			log.Println("[scheduler] sync complete")
			notify("CalSync — Sync Complete", fmt.Sprintf("%d events synced, %d deleted, %d errors", result.Synced, result.Deleted, result.Errors))
			return
		}
	}
}

func startScheduler() {
	go func() {
		notifiedDisconnected := false
		for {
			now := time.Now()
			last := loadLastSync()
			due := !last.IsZero() && now.Sub(last) >= syncInterval

			syncDue := false
			if syncInterval < 24*time.Hour {
				syncDue = due
			} else {
				todayNine := todayAt9()
				syncedToday := !last.IsZero() && last.After(todayNine)
				if !last.IsZero() && now.After(todayNine) && !syncedToday {
					syncDue = true
				} else if !now.After(todayNine) && now.Minute() < 5 {
					log.Printf("[scheduler] next scheduled sync at %s", todayNine.Format("Mon 2 Jan 2006 15:04"))
				}
			}

			if syncDue {
				// Check credentials only when a sync is actually due
				if loadSFCredentials() == nil {
					sessionMu.RLock()
					connected := session.SFInstanceURL != ""
					sessionMu.RUnlock()
					if !connected {
						if !notifiedDisconnected {
							log.Println("[scheduler] Salesforce not connected — waiting for manual reconnect")
							notify("CalSync — Action Required", fmt.Sprintf("Salesforce is not connected. Open %s to reconnect.", appBaseURL))
							notifiedDisconnected = true
						}
						time.Sleep(schedulerTickDuration)
						continue
					}
				}
				notifiedDisconnected = false
				runScheduledSync("sync due")
			}

			time.Sleep(schedulerTickDuration)
		}
	}()
}

func normalizeDateTime(s string) string {
	t := parseDateTime(s)
	if t.IsZero() {
		return s
	}
	return t.UTC().Format(time.RFC3339)
}

func parseDateTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, _ = time.Parse("2006-01-02T15:04:05.000+0000", s)
	}
	return t.UTC()
}

func eventsMatch(existing *googlecalendar.Event, desired *googlecalendar.Event) bool {
	if existing.Summary != desired.Summary {
		return false
	}
	if existing.Description != desired.Description {
		return false
	}
	if existing.Location != desired.Location {
		return false
	}
	if existing.ColorId != desired.ColorId {
		return false
	}
	if !parseDateTime(existing.Start.DateTime).Equal(parseDateTime(desired.Start.DateTime)) {
		return false
	}
	if !parseDateTime(existing.End.DateTime).Equal(parseDateTime(desired.End.DateTime)) {
		return false
	}
	return true
}

func logSyncDetails(result *SyncResult) {
	for _, d := range result.Details {
		log.Printf("[sync] %-10s %s", d.Status, d.Subject)
	}
	log.Printf("[sync] summary: %d synced, %d deleted, %d errors", result.Synced, result.Deleted, result.Errors)
}

// ---- Sync logic ----

type SyncResult struct {
	Synced  int          `json:"synced"`
	Skipped int          `json:"skipped"`
	Deleted int          `json:"deleted"`
	Errors  int          `json:"errors"`
	Details []SyncDetail `json:"details"`
}

type SyncDetail struct {
	Subject string `json:"subject"`
	Status  string `json:"status"`
}

func runSync() (*SyncResult, error) {
	result := &SyncResult{}

	// Snapshot session state under lock — avoids races with HTTP handlers
	sessionMu.RLock()
	instanceURL := session.SFInstanceURL
	sessionID := session.SFSessionID
	userID := session.SFUserID
	googleToken := session.GoogleToken
	cid := session.EventColor
	sessionMu.RUnlock()

	if cid == "" {
		cid = "1"
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UTC().Format("2006-01-02T15:04:05Z")
	query := fmt.Sprintf(
		"SELECT Id, Subject, StartDateTime, EndDateTime, Description, Location, RecordType.Name "+
			"FROM Event WHERE StartDateTime != NULL "+
			"AND StartDateTime >= %s "+
			"AND OwnerId = '%s' "+
			"AND RecordType.Name = 'Billable Utilization' "+
			"ORDER BY StartDateTime ASC",
		today, userID,
	)
	sfResult, err := sfQuery(instanceURL, sessionID, query)
	if err != nil {
		return nil, fmt.Errorf("Salesforce error: %v", err)
	}
	log.Printf("[sync] query returned %d events (OwnerId=%s, RecordType=Billable Utilization)", len(sfResult.Records), userID)
	for i, e := range sfResult.Records {
		log.Printf("[sync]   [%d] Id=%s Subject=%q RecordType=%q Start=%s", i+1, e.ID, e.Subject, e.RecordType.Name, e.StartDateTime)
	}

	ctx := context.Background()
	oauthConfig, err := getOAuthConfig()
	if err != nil {
		return nil, err
	}
	tokenSource := oauthConfig.TokenSource(ctx, googleToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("Google token error: %v", err)
	}

	// Persist refreshed token
	sessionMu.Lock()
	session.GoogleToken = newToken
	sessionMu.Unlock()
	saveGoogleToken(newToken)

	svc, err := googlecalendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}

	syncMap := loadSyncMap()
	currentSFIDs := map[string]bool{}

	// Build a lookup of existing Google Calendar events to avoid duplicates on first sync
	existingGCalEvents := map[string]string{} // "subject|startDateTime" -> gID
	pageToken := ""
	for {
		call := svc.Events.List("primary").
			TimeMin(today).
			SingleEvents(true).
			MaxResults(250)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		gcalList, gcalErr := call.Do()
		if gcalErr != nil {
			log.Printf("[sync] could not fetch existing Google Calendar events for dedup: %v", gcalErr)
			break
		}
		for _, e := range gcalList.Items {
			if e.Status == "cancelled" {
				continue
			}
			start := ""
			if e.Start != nil {
				start = e.Start.DateTime
				if start == "" {
					start = e.Start.Date
				}
			}
			key := e.Summary + "|" + normalizeDateTime(start)
			existingGCalEvents[key] = e.Id
		}
		if gcalList.NextPageToken == "" {
			break
		}
		pageToken = gcalList.NextPageToken
	}

	for _, event := range sfResult.Records {
		if event.StartDateTime == "" || event.EndDateTime == "" {
			result.Skipped++
			continue
		}
		currentSFIDs[event.ID] = true

		subject := event.Subject
		if subject == "" {
			subject = "Salesforce Event"
		}

		gEvent := &googlecalendar.Event{
			Summary:     subject,
			Description: event.Description,
			Location:    event.Location,
			ColorId:     cid,
			Start:       &googlecalendar.EventDateTime{DateTime: event.StartDateTime, TimeZone: "UTC"},
			End:         &googlecalendar.EventDateTime{DateTime: event.EndDateTime, TimeZone: "UTC"},
		}

		time.Sleep(rateLimitDelay)
		if gID, exists := syncMap[event.ID]; exists {
			existing, fetchErr := svc.Events.Get("primary", gID).Do()
			if fetchErr != nil {
				var apiErr *googleapi.Error
				if errors.As(fetchErr, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 410) {
					delete(syncMap, event.ID)
					var created *googlecalendar.Event
					err2 := withBackoff(func() error {
						var e error
						created, e = svc.Events.Insert("primary", gEvent).Do()
						return e
					})
					if err2 != nil {
						result.Errors++
						result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "error: " + err2.Error()})
					} else {
						syncMap[event.ID] = created.Id
						result.Synced++
						result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "recreated"})
					}
				} else {
					result.Errors++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "error: " + fetchErr.Error()})
				}
			} else if existing.Status == "cancelled" {
				delete(syncMap, event.ID)
				var created *googlecalendar.Event
				err2 := withBackoff(func() error {
					var e error
					created, e = svc.Events.Insert("primary", gEvent).Do()
					return e
				})
				if err2 != nil {
					result.Errors++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "error: " + err2.Error()})
				} else {
					syncMap[event.ID] = created.Id
					result.Synced++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "recreated"})
				}
			} else if !eventsMatch(existing, gEvent) {
				err := withBackoff(func() error {
					_, e := svc.Events.Update("primary", gID, gEvent).Do()
					return e
				})
				if err != nil {
					result.Errors++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "error: " + err.Error()})
				} else {
					result.Synced++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "updated"})
				}
			} else {
				result.Skipped++
				result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "skipped"})
			}
		} else {
			// Check if an identical event already exists in Google Calendar (e.g. from a previous install)
			dedupKey := subject + "|" + normalizeDateTime(event.StartDateTime)
			if existingGID, found := existingGCalEvents[dedupKey]; found {
				syncMap[event.ID] = existingGID
				// Fetch the existing event to check if it needs updating (e.g. color change)
				existing, fetchErr := svc.Events.Get("primary", existingGID).Do()
				if fetchErr == nil && !eventsMatch(existing, gEvent) {
					err := withBackoff(func() error {
						_, e := svc.Events.Update("primary", existingGID, gEvent).Do()
						return e
					})
					if err != nil {
						result.Errors++
						result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "error: " + err.Error()})
					} else {
						result.Synced++
						result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "updated"})
					}
				} else {
					result.Skipped++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "skipped"})
				}
			} else {
				var created *googlecalendar.Event
				err := withBackoff(func() error {
					var e error
					created, e = svc.Events.Insert("primary", gEvent).Do()
					return e
				})
				if err != nil {
					result.Errors++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "error: " + err.Error()})
				} else {
					syncMap[event.ID] = created.Id
					result.Synced++
					result.Details = append(result.Details, SyncDetail{Subject: subject, Status: "created"})
				}
			}
		}
	}

	// Delete events removed from Salesforce (only future events — never delete past events)
	for sfID, gID := range syncMap {
		if !currentSFIDs[sfID] {
			existing, fetchErr := svc.Events.Get("primary", gID).Do()
			deleteSubject := sfID
			if fetchErr == nil && existing != nil {
				if existing.Summary != "" {
					deleteSubject = existing.Summary
				}
				if existing.Start != nil {
					start := parseDateTime(existing.Start.DateTime)
					if start.IsZero() {
						start = parseDateTime(existing.Start.Date)
					}
					if !start.IsZero() && start.Before(now) {
						continue
					}
				}
			}
			time.Sleep(rateLimitDelay)
			err := withBackoff(func() error {
				return svc.Events.Delete("primary", gID).Do()
			})
			if err != nil {
				var apiErr *googleapi.Error
				if errors.As(err, &apiErr) && (apiErr.Code == 404 || apiErr.Code == 410) {
					delete(syncMap, sfID)
					result.Deleted++
					result.Details = append(result.Details, SyncDetail{Subject: deleteSubject, Status: "deleted"})
				} else {
					result.Errors++
					result.Details = append(result.Details, SyncDetail{Subject: deleteSubject, Status: "delete error: " + err.Error()})
				}
			} else {
				delete(syncMap, sfID)
				result.Deleted++
				result.Details = append(result.Details, SyncDetail{Subject: deleteSubject, Status: "deleted"})
			}
		}
	}

	saveSyncMap(syncMap)
	return result, nil
}

// ---- HTML template ----

var indexTmpl *template.Template

func loadTemplate() {
	tmplPath := appPath("templates/index.html")
	log.Printf("[startup] loading template from: %s", tmplPath)
	var err error
	indexTmpl, err = template.ParseFiles(tmplPath)
	if err != nil {
		log.Fatalf("[startup] failed to load template: %v\nMake sure the 'templates' folder is in the same directory as SyncApp", err)
	}
}

// ---- Page data ----

type ColorOption struct {
	ID       string
	Name     string
	Hex      string
	Selected bool
}

type PageData struct {
	SFConnected     bool
	GoogleConnected bool
	SFError         string
	GoogleError     string
	Colors          []ColorOption
	SelectedColor   string
	SelectedHex     string
	LastSync        string
	NextSync        string
}

func pageData(sfErr, googleErr string) PageData {
	sessionMu.RLock()
	sfConnected := session.SFInstanceURL != ""
	googleConnected := session.GoogleToken != nil
	selectedColor := session.EventColor
	sessionMu.RUnlock()

	if selectedColor == "" {
		selectedColor = "1"
	}
	gcalColorsMu.RLock()
	colorList := gcalColors
	gcalColorsMu.RUnlock()

	selectedHex := "#4285F4"
	colors := make([]ColorOption, len(colorList))
	for i, c := range colorList {
		sel := c.ID == selectedColor
		colors[i] = ColorOption{ID: c.ID, Name: c.Name, Hex: c.Hex, Selected: sel}
		if sel {
			selectedHex = c.Hex
		}
	}
	lastSync := loadLastSync()
	lastSyncStr := "Never"
	if !lastSync.IsZero() {
		lastSyncStr = lastSync.Format("Mon 2 Jan 2006 at 3:04 PM")
	}
	nextSync := nextDailyAt9()
	nextSyncStr := nextSync.Format("Mon 2 Jan 2006 at 9:00 AM")

	return PageData{
		SFConnected:     sfConnected,
		GoogleConnected: googleConnected,
		SFError:         sfErr,
		GoogleError:     googleErr,
		Colors:          colors,
		SelectedColor:   selectedColor,
		SelectedHex:     selectedHex,
		LastSync:        lastSyncStr,
		NextSync:        nextSyncStr,
	}
}

// ---- Handlers ----

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if err := indexTmpl.ExecuteTemplate(w, "index.html", pageData("", "")); err != nil {
		log.Printf("[handler] template error: %v", err)
	}
}

func handleConnectSalesforce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	r.ParseForm()
	username := strings.TrimSpace(r.FormValue("sf_username"))
	password := strings.TrimSpace(r.FormValue("sf_password"))
	token := strings.TrimSpace(r.FormValue("sf_token"))
	domain := strings.TrimSpace(r.FormValue("sf_domain"))
	if domain == "" {
		domain = "login"
	}
	if err := sfLogin(username, password, token, domain); err != nil {
		if err2 := indexTmpl.ExecuteTemplate(w, "index.html", pageData("Login failed: "+err.Error(), "")); err2 != nil {
			log.Printf("[handler] template error: %v", err2)
		}
		return
	}
	saveSFCredentials(sfCredentials{Username: username, Password: password, Token: token, Domain: domain})
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleConnectGoogle(w http.ResponseWriter, r *http.Request) {
	config, err := getOAuthConfig()
	if err != nil {
		if err2 := indexTmpl.ExecuteTemplate(w, "index.html", pageData("", err.Error())); err2 != nil {
			log.Printf("[handler] template error: %v", err2)
		}
		return
	}
	b := make([]byte, 32)
	crand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	verifierBytes := make([]byte, 32)
	crand.Read(verifierBytes)
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	sessionMu.Lock()
	session.OAuthState = state
	session.CodeVerifier = verifier
	sessionMu.Unlock()

	authURL := config.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.S256ChallengeOption(verifier),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	sessionMu.RLock()
	oauthState := session.OAuthState
	codeVerifier := session.CodeVerifier
	sessionMu.RUnlock()

	if r.FormValue("state") != oauthState {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	config, err := getOAuthConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := config.Exchange(context.Background(), r.FormValue("code"),
		oauth2.VerifierOption(codeVerifier),
	)
	if err != nil {
		if err2 := indexTmpl.ExecuteTemplate(w, "index.html", pageData("", "Google login failed: "+err.Error())); err2 != nil {
			log.Printf("[handler] template error: %v", err2)
		}
		return
	}
	sessionMu.Lock()
	session.GoogleToken = token
	sessionMu.Unlock()
	saveGoogleToken(token)
	go fetchGoogleColors()
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleSetColor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	id := strings.TrimSpace(r.FormValue("color_id"))
	gcalColorsMu.RLock()
	colorList := gcalColors
	gcalColorsMu.RUnlock()
	for _, c := range colorList {
		if c.ID == id {
			sessionMu.Lock()
			session.EventColor = id
			sessionMu.Unlock()
			saveEventColor(id)
			break
		}
	}
	sessionMu.RLock()
	current := session.EventColor
	sessionMu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"color": current})
}

func handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	sessionMu.RLock()
	connected := session.SFInstanceURL != "" && session.GoogleToken != nil
	sessionMu.RUnlock()

	if !connected {
		json.NewEncoder(w).Encode(map[string]string{"error": "Not connected to both accounts"})
		return
	}
	result, err := runSync()
	if err != nil {
		notify("CalSync — Sync Failed", err.Error())
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	saveLastSync(time.Now())
	logSyncDetails(result)
	notify("CalSync — Sync Complete", fmt.Sprintf("%d events synced, %d deleted, %d errors", result.Synced, result.Deleted, result.Errors))
	json.NewEncoder(w).Encode(result)
}

func handleDisconnectSalesforce(w http.ResponseWriter, r *http.Request) {
	sessionMu.Lock()
	session.SFInstanceURL = ""
	session.SFSessionID = ""
	session.SFUserID = ""
	sessionMu.Unlock()
	deleteSFCredentials()
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleDisconnectGoogle(w http.ResponseWriter, r *http.Request) {
	sessionMu.Lock()
	session.GoogleToken = nil
	sessionMu.Unlock()
	deleteGoogleToken()
	http.Redirect(w, r, "/", http.StatusFound)
}

func handleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	http.ServeFile(w, r, appPath("templates/ace-logo.svg"))
}

// ---- Main ----

func main() {
	logFile, err := os.OpenFile(filepath.Join(appDir(), "calsync.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		// Write only to the log file — launchd already captures stdout, so
		// MultiWriter(os.Stdout, logFile) would produce duplicate entries.
		log.SetOutput(logFile)
	}

	log.Printf("[startup] app directory: %s", appDir())
	loadTemplate()
	loadSavedCredentials()
	startScheduler()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/connect/salesforce", handleConnectSalesforce)
	http.HandleFunc("/connect/google", handleConnectGoogle)
	http.HandleFunc("/connect/google/callback", handleGoogleCallback)
	http.HandleFunc("/set-color", handleSetColor)
	http.HandleFunc("/sync", handleSync)
	http.HandleFunc("/disconnect/salesforce", handleDisconnectSalesforce)
	http.HandleFunc("/disconnect/google", handleDisconnectGoogle)
	http.HandleFunc("/logo", handleLogo)

	fmt.Println("Starting CalSync v0.0.9 — Salesforce → Google Calendar Sync")
	fmt.Printf("Open your browser at: %s\n", appBaseURL)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatal(err)
	}
}
