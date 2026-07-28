package server

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dogeorg/dogelytics/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieName = "dgl_sess"
	sessionTTL        = 7 * 24 * time.Hour
)

var parsedTemplates sync.Map

// generateID returns a URL-safe random ID string with n random bytes
func generateID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hashPassword hashes a plaintext password using bcrypt
func hashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// checkPassword compares a bcrypt hash with a plaintext password
func checkPassword(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// signSession produces an HMAC signature for the session payload
func signSession(secret []byte, userID string, exp int64) string {
	data := fmt.Sprintf("%s|%d", userID, exp)
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// setSessionCookie sets the signed session cookie for a user id
func (s *Server) setSessionCookie(w http.ResponseWriter, r *http.Request, userID string) {
	if s.config.SessionSecret == "" {
		// No session secret configured; skip setting cookie.
		log.Printf("[Dogelytics][auth] session secret is empty; not setting session cookie for user_id=%s", userID)
		return
	}

	secureCookie := s.requestIsHTTPS(r) || strings.HasPrefix(strings.ToLower(s.config.PublicURL), "https://")

	expiresAt := time.Now().Add(sessionTTL)
	exp := expiresAt.Unix()
	sig := signSession([]byte(s.config.SessionSecret), userID, exp)
	val := fmt.Sprintf("%s|%d|%s", userID, exp, sig)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    val,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	log.Printf("[Dogelytics][auth] session cookie set for user_id=%s secure=%t", userID, secureCookie)
}

// clearSessionCookie removes the session cookie
func (s *Server) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.requestIsHTTPS(r) || strings.HasPrefix(strings.ToLower(s.config.PublicURL), "https://"),
		SameSite: http.SameSiteLaxMode,
	})
}

// getSessionUserID validates the cookie and returns the user id if present
func (s *Server) getSessionUserID(r *http.Request) (string, bool) {
	uid, ok, _ := s.getSessionUserIDWithReason(r)
	return uid, ok
}

// getSessionUserIDWithReason validates session cookie and returns a reason on failure.
func (s *Server) getSessionUserIDWithReason(r *http.Request) (string, bool, string) {
	if s.config.SessionSecret == "" {
		return "", false, "session-secret-empty"
	}
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" {
		return "", false, "cookie-missing"
	}
	parts := strings.Split(c.Value, "|")
	if len(parts) != 3 {
		return "", false, "cookie-format-invalid"
	}
	uid := parts[0]
	expStr := parts[1]
	sig := parts[2]

	exp, err := parseInt64(expStr)
	if err != nil {
		return "", false, "cookie-exp-invalid"
	}
	if time.Now().Unix() > exp {
		return "", false, "cookie-expired"
	}
	expected := signSession([]byte(s.config.SessionSecret), uid, exp)
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", false, "cookie-signature-invalid"
	}
	return uid, true, ""
}

func parseInt64(s string) (int64, error) {
	var x int64
	_, err := fmt.Sscanf(s, "%d", &x)
	return x, err
}

// renderTemplate renders a cached HTML template with provided data.
func renderTemplate(w http.ResponseWriter, tmpl string, data any) {
	renderNamedTemplate(w, "page", tmpl, data)
}

func renderNamedTemplate(w http.ResponseWriter, name, source string, data any) {
	cacheKey := name + "\x00" + source
	cached, ok := parsedTemplates.Load(cacheKey)
	if !ok {
		parsed := template.Must(template.New(name).Parse(source))
		cached, _ = parsedTemplates.LoadOrStore(cacheKey, parsed)
	}
	t := cached.(*template.Template)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		log.Printf("[Dogelytics][web] render template %q: %v", name, err)
	}
}

// HandleGETIndex shows a landing page (login form)
func (s *Server) HandleGETIndex(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.getSessionUserID(r)
	if ok && userID != "" {
		http.Redirect(w, r, "/keys", http.StatusFound)
		return
	}

	// Conditionally show register link based on EnableSignups
	registerLink := ""
	if s.config.EnableSignups {
		registerLink = `<div class="links">
	<a href="/register">Don't have an account? Register</a>
</div>`
	}

	renderTemplate(w, htmlHeader+`
<h1>Dogelytics</h1>
<p style="text-align: center; font-size: 0.95rem; margin: 0.5rem 0; color: #666;">
  Dogecoin balance analytics API
</p>
<h2 style="margin-top: 1.5rem;">Login</h2>
<form method="post" action="/login">
	<label>
		Email Address
		<input type="email" name="email" placeholder="your@email.com" autocomplete="email" required>
	</label>
	<label>
		Password
		<input type="password" name="password" placeholder="••••••••••••" autocomplete="current-password" required>
	</label>
	<button type="submit">Login</button>
</form>
`+registerLink+htmlFooter, nil)
}

// HandleGETRegister renders the register form
func (s *Server) HandleGETRegister(w http.ResponseWriter, r *http.Request) {
	// Check if signups are enabled
	if !s.config.EnableSignups {
		renderTemplate(w, htmlHeader+`
<h1>Dogelytics</h1>
<p style="text-align: center; font-size: 0.95rem; margin: 0.5rem 0; color: #666;">
  Dogecoin balance analytics API
</p>
<h2 style="margin-top: 1.5rem;">Registration Disabled</h2>
<p style="text-align: center; color: #666;">
  User registration is currently disabled. Please contact an administrator to create an account.
</p>
<div class="links">
	<a href="/">Back to Login</a>
</div>
`+htmlFooter, nil)
		return
	}

	renderTemplate(w, htmlHeader+`
<h1>Dogelytics</h1>
<p style="text-align: center; font-size: 0.95rem; margin: 0.5rem 0; color: #666;">
  Dogecoin balance analytics API
</p>
<h2 style="margin-top: 1.5rem;">Create Account</h2>
<form method="post" action="/register">
	<label>
		Email Address
		<input type="email" name="email" placeholder="your@email.com" autocomplete="email" required>
	</label>
	<label>
		Password (min 12 characters)
		<input type="password" name="password" minlength="12" placeholder="••••••••••••" autocomplete="new-password" required>
	</label>
	<button type="submit">Create Account</button>
</form>
<div class="links">
	<a href="/">Already have an account? Login</a>
</div>
`+htmlFooter, nil)
}

// HandlePOSTRegister creates a new user
func (s *Server) HandlePOSTRegister(w http.ResponseWriter, r *http.Request) {
	// Check if signups are enabled
	if !s.config.EnableSignups {
		http.Error(w, "registration is disabled", http.StatusForbidden)
		return
	}

	if err := parseLimitedForm(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(r.Form.Get("email")))
	password := r.Form.Get("password")
	if email == "" || password == "" || len(password) < 12 {
		http.Error(w, "password must be at least 12 characters", http.StatusBadRequest)
		return
	}
	// Check if exists
	if existing, _, err := s.authStore.GetUserByEmail(r.Context(), email); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	} else if existing.ID != "" {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}
	// Create
	id, err := generateID(16)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	hash, err := hashPassword(password)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	u, err := s.authStore.CreateUser(r.Context(), id, email, hash)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	s.setSessionCookie(w, r, u.ID)
	log.Printf("[Dogelytics][auth] registration success email=%s user_id=%s", email, u.ID)
	http.Redirect(w, r, "/keys", http.StatusFound)
}

// HandleGETLogin renders the login form (same as index)
func (s *Server) HandleGETLogin(w http.ResponseWriter, r *http.Request) {
	// Redirect to index for consistent experience
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandlePOSTLogin authenticates a user
func (s *Server) HandlePOSTLogin(w http.ResponseWriter, r *http.Request) {
	if err := parseLimitedForm(w, r); err != nil {
		log.Printf("[Dogelytics][auth] login failed: parse form error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(strings.ToLower(r.Form.Get("email")))
	password := r.Form.Get("password")
	if email == "" || password == "" {
		log.Printf("[Dogelytics][auth] login failed: missing email or password")
		http.Error(w, "invalid credentials", http.StatusBadRequest)
		return
	}
	u, hash, err := s.authStore.GetUserByEmail(r.Context(), email)
	if err != nil {
		log.Printf("[Dogelytics][auth] login failed for email=%s: store error: %v", email, err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	if u.ID == "" || checkPassword(hash, password) != nil {
		log.Printf("[Dogelytics][auth] login failed for email=%s: invalid credentials", email)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	s.setSessionCookie(w, r, u.ID)
	log.Printf("[Dogelytics][auth] login success email=%s user_id=%s", email, u.ID)
	http.Redirect(w, r, "/keys", http.StatusFound)
}

// HandlePOSTLogout logs out the session
func (s *Server) HandlePOSTLogout(w http.ResponseWriter, r *http.Request) {
	s.clearSessionCookie(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

// requireLogin ensures the request has a valid session, otherwise redirect
func (s *Server) RequireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok, reason := s.getSessionUserIDWithReason(r)
		if ok && uid != "" {
			next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), uid)))
			return
		}
		log.Printf("[Dogelytics][auth] unauthenticated request path=%s reason=%s", r.URL.Path, reason)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

// getUserFromRequest fetches the user if session is valid
func (s *Server) getUserFromRequest(r *http.Request) (store.User, bool) {
	uid, ok := s.getSessionUserID(r)
	if !ok || uid == "" {
		return store.User{}, false
	}
	u, err := s.authStore.GetUserByID(r.Context(), uid)
	if err != nil || u.ID == "" {
		return store.User{}, false
	}
	return u, true
}
