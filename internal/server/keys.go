package server

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
	keys, err := s.authStore.GetAPIKeysByUserID(u.ID)
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
		Expired   bool
	}
	var list []keyView
	for _, k := range keys {
		kv := keyView{
			KID:       k.KID,
			CreatedAt: k.CreatedAt,
			ExpiresAt: "",
			Revoked:   k.RevokedAt.Valid,
			Expired:   false,
		}
		if k.ExpiresAt.Valid {
			kv.ExpiresAt = k.ExpiresAt.Time.UTC().Format(time.RFC3339)
			kv.Expired = time.Now().After(k.ExpiresAt.Time)
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
<style>
.stats-widget {
  background: #c0c0c0;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #808080;
  border-bottom: 2px solid #808080;
  padding: 12px;
  margin-bottom: 12px;
  color: #000;
}
.stats-widget h3 {
  color: #fff;
  margin: 0 0 8px 0;
  font-size: 11px;
  background: #000080;
  padding: 3px 5px;
  font-weight: bold;
}
.stats-controls {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}
.control-group {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: wrap;
}
.control-label {
  font-size: 11px;
  min-width: 65px;
  font-weight: bold;
}
.stats-timeframe button,
.stats-filter button {
  padding: 3px 8px;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #000;
  border-bottom: 2px solid #000;
  box-shadow: inset 1px 1px 0 #dfdfdf, inset -1px -1px 0 #808080;
  background: #c0c0c0;
  color: #000;
  cursor: pointer;
  font-size: 11px;
  font-family: inherit;
  min-width: 50px;
}
.stats-timeframe button:hover,
.stats-filter button:hover {
  background: #c0c0c0;
}
.stats-timeframe button.active,
.stats-filter button.active {
  border-top: 2px solid #000;
  border-left: 2px solid #000;
  border-right: 2px solid #fff;
  border-bottom: 2px solid #fff;
  box-shadow: inset -1px -1px 0 #dfdfdf, inset 1px 1px 0 #808080;
  padding: 4px 7px 2px 9px;
}
.stats-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  margin-bottom: 12px;
  max-width: 250px;
}
.stat-box {
  background: #c0c0c0;
  border-top: 2px solid #808080;
  border-left: 2px solid #808080;
  border-right: 2px solid #fff;
  border-bottom: 2px solid #fff;
  box-shadow: inset -1px -1px 0 #c0c0c0, inset 1px 1px 0 #000;
  padding: 8px;
  text-align: center;
}
.stat-value {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 4px;
  color: #000;
}
.stat-label {
  font-size: 10px;
  color: #000;
}
.chart-container {
  background: #fff;
  border-top: 2px solid #808080;
  border-left: 2px solid #808080;
  border-right: 2px solid #fff;
  border-bottom: 2px solid #fff;
  padding: 8px;
  margin-top: 8px;
}
.chart-canvas {
  width: 100%;
  height: 220px;
}
.update-indicator {
  font-size: 10px;
  text-align: right;
  margin-top: 4px;
  color: #000;
}
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .stats-controls {
    flex-direction: column;
  }
}
#keys-table td {
  vertical-align: middle;
  height: 40px;
  padding: 4px;
}
#keys-table td form {
  margin: 0;
  display: inline-block;
  vertical-align: middle;
}
#keys-table td button {
  padding: 3px 10px;
  font-size: 10px;
  margin: 0;
  min-width: 60px;
}
</style>

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
<div id="keys-container">
  <table id="keys-table" style="table-layout: fixed;">
    <thead>
      <tr>
        <th>Key ID</th>
        <th>Created</th>
        <th>Expires</th>
        <th>Status</th>
        <th>Actions</th>
      </tr>
    </thead>
    <tbody id="keys-tbody">
      {{range .Keys}}
      <tr data-revoked="{{.Revoked}}" data-expired="{{if .ExpiresAt}}{{if .Expired}}true{{else}}false{{end}}{{else}}false{{end}}">
        <td><code>{{.KID}}</code></td>
        <td>{{.CreatedAt.Format "2006-01-02 15:04"}}</td>
        <td>{{if .ExpiresAt}}{{.ExpiresAt}}{{else}}<em>Never</em>{{end}}</td>
        <td>{{if .Revoked}}<span style="color: #e74c3c;">Revoked</span>{{else if .Expired}}<span style="color: #e67e22;">Expired</span>{{else}}<span style="color: #27ae60;">Active</span>{{end}}</td>
        <td>
          {{if not .Revoked}}
          <form method="post" action="/keys/revoke" class="form-inline">
            <input type="hidden" name="kid" value="{{.KID}}">
            <button type="submit" class="danger">Revoke</button>
          </form>
          {{end}}
        </td>
      </tr>
      {{end}}
    </tbody>
  </table>

  <div id="pagination" style="margin-top: 12px; text-align: center; min-height: 35px; display: flex; justify-content: center; align-items: flex-start;"></div>
</div>

<div style="margin-top: 12px; margin-bottom: 12px;">
  <label style="display: flex; align-items: center; gap: 6px; cursor: pointer;">
    <input type="checkbox" id="show-revoked" style="width: auto; margin: 0;">
    <span>Show expired/revoked keys</span>
  </label>
</div>

{{else}}
<p style="text-align: center; color: #999; padding: 1.5rem;">
  No API keys yet. Create one above to get started.
</p>
{{end}}

<script>
// Pagination and filtering for API keys
const KEYS_PER_PAGE = 5;
let currentPage = 1;
let showRevoked = false;

function filterAndPaginateKeys() {
  const tbody = document.getElementById('keys-tbody');
  if (!tbody) return;
  
  const rows = Array.from(tbody.querySelectorAll('tr:not(.placeholder-row)'));
  
  // Filter rows based on revoked/expired status
  const filteredRows = rows.filter(row => {
    const isRevoked = row.dataset.revoked === 'true';
    const isExpired = row.dataset.expired === 'true';
    return showRevoked || (!isRevoked && !isExpired);
  });
  
  // Hide all rows first
  rows.forEach(row => row.style.display = 'none');
  
  // Calculate pagination
  const totalPages = Math.ceil(filteredRows.length / KEYS_PER_PAGE);
  const startIdx = (currentPage - 1) * KEYS_PER_PAGE;
  const endIdx = startIdx + KEYS_PER_PAGE;
  
  // Show only the current page
  const visibleRows = filteredRows.slice(startIdx, endIdx);
  visibleRows.forEach(row => row.style.display = '');
  
  // Add placeholder rows to maintain height
  const placeholdersNeeded = KEYS_PER_PAGE - visibleRows.length;
  
  // Remove existing placeholders
  tbody.querySelectorAll('.placeholder-row').forEach(p => p.remove());
  
  // Add new placeholders if needed
  for (let i = 0; i < placeholdersNeeded; i++) {
    const placeholder = document.createElement('tr');
    placeholder.className = 'placeholder-row';
    placeholder.innerHTML = '<td colspan="5" style="height: 40px; border: none; background: transparent;"></td>';
    tbody.appendChild(placeholder);
  }
  
  // Update pagination controls
  updatePagination(totalPages, filteredRows.length);
}

function updatePagination(totalPages, totalKeys) {
  const pagination = document.getElementById('pagination');
  if (!pagination) return;
  
  if (totalKeys === 0) {
    pagination.innerHTML = '<p style="color: #808080; margin: 0;">No keys to display</p>';
    return;
  }
  
  if (totalPages <= 1) {
    pagination.innerHTML = '<div style="height: 26px;"></div>'; // Reserve space even when no pagination
    return;
  }
  
  let html = '<div style="display: flex; gap: 4px; align-items: center; height: 26px;">';
  
  // Previous button
  if (currentPage > 1) {
    html += '<button onclick="goToPage(' + (currentPage - 1) + ')">Previous</button>';
  } else {
    html += '<button disabled style="visibility: hidden;">Previous</button>';
  }
  
  // Page numbers - show max 7 buttons with ellipsis
  const maxButtons = 7;
  if (totalPages <= maxButtons) {
    // Show all pages
    for (let i = 1; i <= totalPages; i++) {
      if (i === currentPage) {
        html += '<button class="active" style="font-weight: bold;">' + i + '</button>';
      } else {
        html += '<button onclick="goToPage(' + i + ')">' + i + '</button>';
      }
    }
  } else {
    // Show first, last, and pages around current
    for (let i = 1; i <= totalPages; i++) {
      if (i === 1 || i === totalPages || (i >= currentPage - 1 && i <= currentPage + 1)) {
        if (i === currentPage) {
          html += '<button class="active" style="font-weight: bold;">' + i + '</button>';
        } else {
          html += '<button onclick="goToPage(' + i + ')">' + i + '</button>';
        }
      } else if (i === currentPage - 2 || i === currentPage + 2) {
        html += '<span style="padding: 4px 8px;">...</span>';
      }
    }
  }
  
  // Next button
  if (currentPage < totalPages) {
    html += '<button onclick="goToPage(' + (currentPage + 1) + ')">Next</button>';
  } else {
    html += '<button disabled style="visibility: hidden;">Next</button>';
  }
  
  html += '</div>';
  pagination.innerHTML = html;
}

function goToPage(page) {
  currentPage = page;
  filterAndPaginateKeys();
}

// Toggle revoked keys
document.addEventListener('DOMContentLoaded', function() {
  const checkbox = document.getElementById('show-revoked');
  if (checkbox) {
    checkbox.addEventListener('change', function() {
      showRevoked = this.checked;
      currentPage = 1; // Reset to first page
      filterAndPaginateKeys();
    });
    
    // Initial filter and pagination
    filterAndPaginateKeys();
  }
});
</script>

<div class="stats-widget" style="margin-top: 2rem;">
  <h3>Usage Statistics</h3>
  <div class="stats-controls">
    <div class="control-group">
      <span class="control-label">Filter:</span>
      <div class="stats-filter">
        <button onclick="setFilter('overall')" id="filter-overall" class="active">Overall</button>
        <button onclick="setFilter('keys')" id="filter-keys">My Keys</button>
      </div>
    </div>
    <div class="control-group">
      <span class="control-label">Timeframe:</span>
      <div class="stats-timeframe">
        <button onclick="setTimeframe('hour')" id="btn-hour">Hour</button>
        <button onclick="setTimeframe('day')" id="btn-day" class="active">Day</button>
        <button onclick="setTimeframe('week')" id="btn-week">Week</button>
        <button onclick="setTimeframe('month')" id="btn-month">Month</button>
        <button onclick="setTimeframe('year')" id="btn-year">Year</button>
      </div>
    </div>
  </div>
  <div class="stats-grid">
    <div class="stat-box">
      <div class="stat-value" id="stat-wallets">-</div>
      <div class="stat-label">Wallets Checked</div>
    </div>
  </div>
  <div class="chart-container">
    <canvas id="stats-chart" class="chart-canvas"></canvas>
  </div>
  <div class="update-indicator" id="update-time">Loading...</div>
</div>

<script>
// Load saved preferences or use defaults
let currentTimeframe = localStorage.getItem('dogelytics_timeframe') || 'day';
let currentFilter = localStorage.getItem('dogelytics_filter') || 'overall';
let chart = null;
let refreshTimer = null;
let refreshInFlight = false;

function setTimeframe(timeframe) {
  currentTimeframe = timeframe;
  localStorage.setItem('dogelytics_timeframe', timeframe);
  document.querySelectorAll('.stats-timeframe button').forEach(btn => btn.classList.remove('active'));
  document.getElementById('btn-' + timeframe).classList.add('active');
  loadStatsOnce();
}

function setFilter(filter) {
  currentFilter = filter;
  localStorage.setItem('dogelytics_filter', filter);
  document.querySelectorAll('.stats-filter button').forEach(btn => btn.classList.remove('active'));
  document.getElementById('filter-' + filter).classList.add('active');
  loadStatsOnce();
}

// Restore active states on page load
document.addEventListener('DOMContentLoaded', function() {
  document.querySelectorAll('.stats-timeframe button').forEach(btn => btn.classList.remove('active'));
  document.getElementById('btn-' + currentTimeframe).classList.add('active');
  document.querySelectorAll('.stats-filter button').forEach(btn => btn.classList.remove('active'));
  document.getElementById('filter-' + currentFilter).classList.add('active');
});

async function loadStatsOnce() {
  if (refreshInFlight) return;
  refreshInFlight = true;

  const filterParam = currentFilter === 'overall' ? '' : '&filter=' + currentFilter;
  
  try {
    const [usageResp, tsResp] = await Promise.all([
      fetch('/api/stats/usage?timeframe=' + currentTimeframe + filterParam),
      fetch('/api/stats/timeseries?timeframe=' + currentTimeframe + filterParam),
    ]);

    const usage = await usageResp.json();
    document.getElementById('stat-wallets').textContent = usage.wallets_checked.toLocaleString();
    document.getElementById('update-time').textContent = 'Updated: ' + new Date().toLocaleTimeString();

    const ts = await tsResp.json();
    updateChart(ts);
  } catch (error) {
    console.error('Error loading stats:', error);
    document.getElementById('stat-wallets').textContent = 'Error';
  } finally {
    refreshInFlight = false;
  }
}

function updateChart(data) {
  const canvas = document.getElementById('stats-chart');
  const ctx = canvas.getContext('2d');
  
  // Set canvas size
  // Avoid resizing every refresh (resize triggers expensive reflow + clears state).
  // If the canvas width changes (responsive layout), we resize on demand.
  const desiredWidth = canvas.clientWidth || canvas.offsetWidth;
  const desiredHeight = 220;
  if (canvas.width !== desiredWidth) canvas.width = desiredWidth;
  if (canvas.height !== desiredHeight) canvas.height = desiredHeight;
  
  if (!data || data.length === 0) {
    ctx.fillStyle = '#808080';
    ctx.font = '11px "MS Sans Serif", Tahoma, Arial, sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('No data available', canvas.width / 2, canvas.height / 2);
    return;
  }
  
  // Helper function to calculate nice rounded max value
  function getNiceMax(value) {
    if (value === 0) return 10;
    
    const magnitude = Math.pow(10, Math.floor(Math.log10(value)));
    const normalized = value / magnitude;
    
    let nice;
    if (normalized <= 1) nice = 1;
    else if (normalized <= 2) nice = 2;
    else if (normalized <= 5) nice = 5;
    else nice = 10;
    
    return nice * magnitude;
  }
  
  // Helper function to format Y-axis labels
  function formatYLabel(value) {
    if (value >= 1000000) return (value / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
    if (value >= 1000) return (value / 1000).toFixed(1).replace(/\.0$/, '') + 'k';
    return value.toString();
  }
  
  // Downsample to keep rendering fast for large timeframes.
  // Canvas charts don't benefit from thousands of points in a few hundred pixels.
  let series = data;
  const MAX_POINTS = 300;
  if (series.length > MAX_POINTS) {
    const step = Math.ceil(series.length / MAX_POINTS);
    const sampled = [];
    for (let i = 0; i < series.length; i += step) sampled.push(series[i]);
    const last = series[series.length - 1];
    if (sampled[sampled.length - 1] !== last) sampled.push(last);
    series = sampled;
  }

  // Extract data + max without spreading (spread can be costly for large arrays).
  const wallets = new Array(series.length);
  let dataMax = 1;
  for (let i = 0; i < series.length; i++) {
    const v = (series[i] && typeof series[i].wallets_checked === 'number') ? series[i].wallets_checked : 0;
    wallets[i] = v;
    if (v > dataMax) dataMax = v;
  }
  const maxValue = getNiceMax(dataMax);
  
  // Clear canvas
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  
  // Draw grid and Y-axis labels (Windows 95 style)
  const leftPadding = 45;
  const rightPadding = 10;
  const topPadding = 15;
  const bottomPadding = 15;
  
  ctx.strokeStyle = '#c0c0c0';
  ctx.lineWidth = 1;
  ctx.fillStyle = '#000';
  ctx.font = '10px "MS Sans Serif", Tahoma, Arial, sans-serif';
  ctx.textAlign = 'right';
  
  for (let i = 0; i <= 4; i++) {
    const y = topPadding + ((canvas.height - topPadding - bottomPadding) / 4) * i;
    const value = maxValue * (1 - i / 4);
    
    // Draw grid line
    ctx.beginPath();
    ctx.moveTo(leftPadding, y);
    ctx.lineTo(canvas.width - rightPadding, y);
    ctx.stroke();
    
    // Draw Y-axis label
    ctx.fillText(formatYLabel(value), leftPadding - 5, y + 4);
  }
  
  // Draw line
  const chartWidth = canvas.width - leftPadding - rightPadding;
  const chartHeight = canvas.height - topPadding - bottomPadding;
  const pointSpacing = chartWidth / Math.max(wallets.length - 1, 1);
  
  ctx.strokeStyle = '#000';
  ctx.lineWidth = 2;
  ctx.beginPath();
  let started = false;
  wallets.forEach((value, index) => {
    const x = leftPadding + index * pointSpacing;
    const y = topPadding + chartHeight - (value / maxValue) * chartHeight;
    if (!started) {
      ctx.moveTo(x, y);
      started = true;
    } else {
      ctx.lineTo(x, y);
    }
  });
  ctx.stroke();
}

function startAutoRefresh() {
  const tick = async () => {
    if (!document.hidden) {
      await loadStatsOnce();
    }
    refreshTimer = window.setTimeout(tick, 3000);
  };
  tick();
}

// Load initial stats + auto-refresh
startAutoRefresh();

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
  if (refreshTimer) window.clearTimeout(refreshTimer);
});
</script>

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
	existing, err := s.authStore.GetAPIKeysByUserID(u.ID)
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
	if _, err := s.authStore.CreateAPIKey(id, u.ID, kid, secretHash, expiresAt); err != nil {
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

<div class="token-display">
  <div class="token-display-row">
    <span id="api-token" class="token-value">{{.Token}}</span>
    <button type="button" id="copy-token-btn" class="copy-token-btn" aria-label="Copy API key" title="Copy API key">&#128203;</button>
  </div>
  <div id="copy-status" class="copy-status" aria-live="polite"></div>
</div>

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

<script>
  (function() {
    function copyWithExecCommand(text) {
      var temp = document.createElement("textarea");
      temp.value = text;
      temp.setAttribute("readonly", "");
      temp.style.position = "absolute";
      temp.style.left = "-9999px";
      document.body.appendChild(temp);
      temp.select();
      var ok = false;
      try {
        ok = document.execCommand("copy");
      } catch (_) {
        ok = false;
      }
      document.body.removeChild(temp);
      return ok;
    }

    var button = document.getElementById("copy-token-btn");
    var tokenEl = document.getElementById("api-token");
    var statusEl = document.getElementById("copy-status");
    if (!button || !tokenEl || !statusEl) {
      return;
    }
    button.addEventListener("click", function() {
      var token = tokenEl.textContent || "";
      if (!token) {
        statusEl.textContent = "No API key to copy.";
        return;
      }
      if (!navigator.clipboard || typeof navigator.clipboard.writeText !== "function") {
        if (copyWithExecCommand(token)) {
          statusEl.textContent = "Copied API key to clipboard.";
        } else {
          statusEl.textContent = "Failed to copy API key.";
        }
        return;
      }
      navigator.clipboard.writeText(token).then(function() {
        statusEl.textContent = "Copied API key to clipboard.";
      }).catch(function() {
        if (copyWithExecCommand(token)) {
          statusEl.textContent = "Copied API key to clipboard.";
        } else {
          statusEl.textContent = "Failed to copy API key.";
        }
      });
    });
  })();
</script>
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
	if err := s.authStore.RevokeAPIKey(u.ID, kid, time.Now()); err != nil {
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
	if err := s.authStore.UpdateAPIKeyExpiry(u.ID, kid, expiresAt); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/keys", http.StatusFound)
}
