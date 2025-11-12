package web

import (
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HandleGETKeys shows the user's keys and management UI
func (s *Server) HandleGETKeys(w http.ResponseWriter, r *http.Request) {
	u, ok := s.getUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	keys, err := s.store.GetAPIKeysByUserID(u.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	// Sort newest first (already ordered by created_at desc, but ensure)
	sort.SliceStable(keys, func(i, j int) bool { return keys[i].CreatedAt.After(keys[j].CreatedAt) })

	type keyView struct {
		KID       string
		CreatedAt time.Time
		ExpiresAt string
		Revoked   bool
	}
	var list []keyView
	for _, k := range keys {
		kv := keyView{
			KID:       k.KID,
			CreatedAt: k.CreatedAt,
			ExpiresAt: "",
			Revoked:   k.RevokedAt.Valid,
		}
		if k.ExpiresAt.Valid {
			kv.ExpiresAt = k.ExpiresAt.Time.UTC().Format(time.RFC3339)
		}
		list = append(list, kv)
	}

	// Count active keys
	activeCount := 0
	for _, k := range keys {
		if k.RevokedAt.Valid {
			continue
		}
		if k.ExpiresAt.Valid && time.Now().After(k.ExpiresAt.Time) {
			continue
		}
		activeCount++
	}

	data := struct {
		Email          string
		Keys           []keyView
		MaxKeysPerUser int
		ActiveCount    int
		CanCreateMore  bool
	}{
		Email:          u.Email,
		Keys:           list,
		MaxKeysPerUser: s.config.MaxKeysPerUser,
		ActiveCount:    activeCount,
		CanCreateMore:  activeCount < s.config.MaxKeysPerUser,
	}
	t := template.Must(template.New("keys").Parse(htmlHeader + `
<h2>API Keys Management</h2>
<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
  <span class="user-badge">{{.Email}}</span>
  <form method="post" action="/logout">
    <button type="submit" class="secondary">Logout</button>
  </form>
</div>

<h3>Create New Key</h3>
{{if .CanCreateMore}}
<div class="alert">
  <strong>Important:</strong> API keys are shown only once. Save them securely.
</div>
<form method="post" action="/keys">
  <label>
    Expiry Date (optional, max 1 year)
    <input type="date" name="expiry" placeholder="Leave empty for no expiration">
  </label>
  <button type="submit">Create New API Key</button>
</form>
{{else}}
<div class="alert" style="background: #f8d7da; border-left-color: #dc3545; color: #721c24;">
  <strong>Maximum keys reached.</strong><br>
  You have {{.ActiveCount}} active key(s). Maximum allowed: {{.MaxKeysPerUser}}.<br>
  Revoke an existing key to create a new one.
</div>
{{end}}

<h3>Your Keys ({{.ActiveCount}} active / {{.MaxKeysPerUser}} max)</h3>
{{if .Keys}}
<table>
  <tr>
    <th>Key ID</th>
    <th>Created</th>
    <th>Expires</th>
    <th>Status</th>
    <th>Actions</th>
  </tr>
  {{range .Keys}}
  <tr>
    <td><code>{{.KID}}</code></td>
    <td>{{.CreatedAt.Format "2006-01-02 15:04"}}</td>
    <td>{{if .ExpiresAt}}{{.ExpiresAt}}{{else}}<em>Never</em>{{end}}</td>
    <td>{{if .Revoked}}<span style="color: #e74c3c;">Revoked</span>{{else}}<span style="color: #27ae60;">Active</span>{{end}}</td>
    <td>
      {{if not .Revoked}}
      <form method="post" action="/keys/revoke" class="form-inline">
        <input type="hidden" name="kid" value="{{.KID}}">
        <button type="submit" class="danger">Revoke</button>
      </form>
      <form method="post" action="/keys/expire" class="form-inline">
        <input type="hidden" name="kid" value="{{.KID}}">
        <input type="date" name="expiry" style="width: 140px;">
        <button type="submit" class="secondary">Set Expiry</button>
      </form>
      {{end}}
    </td>
  </tr>
  {{end}}
</table>
{{else}}
<p style="text-align: center; color: #999; padding: 1.5rem;">
  No API keys yet. Create one above to get started.
</p>
{{end}}
` + htmlFooter))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(w, data)
}

// HandlePOSTCreateKey creates a new key and shows the token once
func (s *Server) HandlePOSTCreateKey(w http.ResponseWriter, r *http.Request) {
	u, ok := s.getUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	// Enforce per-user limit
	existing, err := s.store.GetAPIKeysByUserID(u.ID)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	activeCount := 0
	now := time.Now()
	for _, k := range existing {
		if k.RevokedAt.Valid {
			continue
		}
		if k.ExpiresAt.Valid && now.After(k.ExpiresAt.Time) {
			continue
		}
		activeCount++
	}
	if activeCount >= s.config.MaxKeysPerUser {
		http.Error(w, "max keys reached", http.StatusForbidden)
		return
	}
	// Expiry: default 1 year; clamp to at most 1 year from now
	var expiresAt *time.Time
	expStr := strings.TrimSpace(r.Form.Get("expiry"))
	maxExp := now.Add(365 * 24 * time.Hour)
	if expStr == "" {
		e := maxExp
		expiresAt = &e
	} else {
		// parse yyyy-mm-dd as UTC midnight
		d, err := time.Parse("2006-01-02", expStr)
		if err != nil {
			http.Error(w, "invalid expiry", http.StatusBadRequest)
			return
		}
		e := d.UTC()
		if e.After(maxExp) {
			e = maxExp
		}
		expiresAt = &e
	}
	// Generate token
	kid, err := generateID(8)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	secret, err := generateID(24)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	secretHashBytes, err := bcrypt.GenerateFromPassword([]byte(secret), 12)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	secretHash := string(secretHashBytes)
	id, err := generateID(16)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if _, err := s.store.CreateAPIKey(id, u.ID, kid, secretHash, expiresAt); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	token := "dglk_" + kid + "." + secret
	// Show once page
	t := template.Must(template.New("created").Parse(htmlHeader + `
<h2>API Key Created Successfully</h2>

<div class="alert success">
  <strong>Your API key has been generated.</strong><br>
  Copy it now - it will never be shown again.
</div>

<div class="token-display">{{.Token}}</div>

<div class="alert">
  <strong>How to use:</strong><br>
  Include this header in your requests:<br>
  <code>Authorization: Bearer {{.Token}}</code>
</div>

<div style="margin-top: 1rem;">
  <strong>Example usage:</strong>
  <pre style="background: #f8f9fa; padding: 0.75rem; border-radius: 6px; overflow-x: auto; font-size: 0.85rem;">
curl -H "Authorization: Bearer {{.Token}}" \
  "http://localhost:4420/balance?address=YOUR_DOGE_ADDRESS"</pre>
</div>

<div class="links">
  <a href="/keys">Back to API Keys</a>
</div>
` + htmlFooter))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(w, struct{ Token string }{Token: token})
}

// HandlePOSTRevokeKey revokes a key
func (s *Server) HandlePOSTRevokeKey(w http.ResponseWriter, r *http.Request) {
	u, ok := s.getUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	kid := strings.TrimSpace(r.Form.Get("kid"))
	if kid == "" {
		http.Error(w, "missing kid", http.StatusBadRequest)
		return
	}
	if err := s.store.RevokeAPIKey(u.ID, kid, time.Now()); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/keys", http.StatusFound)
}

// HandlePOSTExpireKey sets expiry for a key
func (s *Server) HandlePOSTExpireKey(w http.ResponseWriter, r *http.Request) {
	u, ok := s.getUserFromRequest(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	kid := strings.TrimSpace(r.Form.Get("kid"))
	expStr := strings.TrimSpace(r.Form.Get("expiry"))
	if kid == "" {
		http.Error(w, "missing kid", http.StatusBadRequest)
		return
	}
	var expiresAt *time.Time
	if expStr != "" {
		d, err := time.Parse("2006-01-02", expStr)
		if err != nil {
			http.Error(w, "invalid expiry", http.StatusBadRequest)
			return
		}
		maxExp := time.Now().Add(365 * 24 * time.Hour)
		e := d.UTC()
		if e.After(maxExp) {
			e = maxExp
		}
		expiresAt = &e
	}
	if err := s.store.UpdateAPIKeyExpiry(u.ID, kid, expiresAt); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/keys", http.StatusFound)
}
