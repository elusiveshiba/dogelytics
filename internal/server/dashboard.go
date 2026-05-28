package server

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// HandleGETDashboard renders the public dashboard UI.
func (s *Server) HandleGETDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		ShowAdminLink bool
		AdminURL      string
	}{
		ShowAdminLink: s.config.EnableAdminUI,
		AdminURL:      uiURLForPort(r, s.config.AdminUIPort),
	}

	renderTemplate(w, htmlHeader+`
<style>
.dashboard-shell {
  display: grid;
  gap: 12px;
}
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 8px;
}
.dashboard-card {
  background: #c0c0c0;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #808080;
  border-bottom: 2px solid #808080;
  padding: 8px;
}
.dashboard-card h3 {
  margin-top: 0;
}
.dashboard-stat-value {
  font-size: 20px;
  font-weight: bold;
  margin-top: 6px;
}
.dashboard-balance-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 8px;
  margin-top: 10px;
}
.dashboard-balance-item {
  background: #fff;
  border-top: 1px solid #808080;
  border-left: 1px solid #808080;
  border-right: 1px solid #fff;
  border-bottom: 1px solid #fff;
  padding: 8px;
}
.dashboard-balance-label {
  color: #666;
  font-size: 10px;
  margin-bottom: 4px;
}
.dashboard-balance-value {
  font-family: "Courier New", monospace;
  font-size: 12px;
  font-weight: bold;
}
.dashboard-status {
  min-height: 18px;
  margin-top: 8px;
  color: #000080;
}
.dashboard-status.error {
  color: #8b0000;
}
</style>
<div class="dashboard-shell">
  <div>
    <h1>Dogelytics Dashboard</h1>
    <p style="text-align: center;">
      Check a Dogecoin wallet balance and view a quick snapshot of recent activity.
    </p>
    {{if .ShowAdminLink}}
    <div class="links">
      <a href="{{.AdminURL}}">Open admin UI</a>
    </div>
    {{end}}
  </div>

  <div class="dashboard-card">
    <h2>Wallet Checker</h2>
    <form id="wallet-checker-form">
      <label>
        Dogecoin wallet address
        <input type="text" id="wallet-address" name="address" placeholder="D..." autocomplete="off" required>
      </label>
      <button type="submit">Check Balance</button>
    </form>
    <div id="wallet-status" class="dashboard-status" aria-live="polite"></div>
    <div id="wallet-results" class="dashboard-balance-grid" hidden>
      <div class="dashboard-balance-item">
        <div class="dashboard-balance-label">Current</div>
        <div id="balance-current" class="dashboard-balance-value">-</div>
      </div>
      <div class="dashboard-balance-item">
        <div class="dashboard-balance-label">Available</div>
        <div id="balance-available" class="dashboard-balance-value">-</div>
      </div>
      <div class="dashboard-balance-item">
        <div class="dashboard-balance-label">Incoming</div>
        <div id="balance-incoming" class="dashboard-balance-value">-</div>
      </div>
      <div class="dashboard-balance-item">
        <div class="dashboard-balance-label">Outgoing</div>
        <div id="balance-outgoing" class="dashboard-balance-value">-</div>
      </div>
    </div>
  </div>

  <div class="dashboard-card">
    <h2>Server Snapshot</h2>
    <div class="dashboard-grid">
      <div class="dashboard-card">
        <h3>Total Wallet Checks</h3>
        <div id="stat-total-wallets" class="dashboard-stat-value">-</div>
      </div>
      <div class="dashboard-card">
        <h3>Wallet Checks (24h)</h3>
        <div id="stat-wallets-24h" class="dashboard-stat-value">-</div>
      </div>
      <div class="dashboard-card">
        <h3>Unique Wallets (24h)</h3>
        <div id="stat-unique-wallets-24h" class="dashboard-stat-value">-</div>
      </div>
      <div class="dashboard-card">
        <h3>Indexed Height</h3>
        <div id="stat-height" class="dashboard-stat-value">-</div>
      </div>
    </div>
    <div id="stats-status" class="dashboard-status" aria-live="polite"></div>
  </div>
</div>

<script>
  (function() {
    const form = document.getElementById('wallet-checker-form');
    const addressInput = document.getElementById('wallet-address');
    const walletStatus = document.getElementById('wallet-status');
    const walletResults = document.getElementById('wallet-results');
    const statsStatus = document.getElementById('stats-status');

    function setText(id, value) {
      const element = document.getElementById(id);
      if (element) element.textContent = value;
    }

    async function loadDashboardStats() {
      statsStatus.textContent = 'Loading dashboard stats...';
      statsStatus.className = 'dashboard-status';

      try {
        const response = await fetch('/api/dashboard-stats');
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.message || 'Failed to load dashboard stats');
        }

        setText('stat-total-wallets', String(payload.total_wallets_checked));
        setText('stat-wallets-24h', String(payload.wallets_checked_last_24h));
        setText('stat-unique-wallets-24h', String(payload.unique_wallets_last_24h));
        setText('stat-height', String(payload.height));

        if (payload.available) {
          statsStatus.textContent = 'Dashboard stats updated.';
          statsStatus.className = 'dashboard-status';
        } else {
          statsStatus.textContent = payload.message || 'Usage statistics are currently unavailable.';
          statsStatus.className = 'dashboard-status error';
        }
      } catch (error) {
        statsStatus.textContent = error.message || 'Failed to load dashboard stats.';
        statsStatus.className = 'dashboard-status error';
      }
    }

    form.addEventListener('submit', async function(event) {
      event.preventDefault();
      walletResults.hidden = true;
      walletStatus.textContent = 'Checking wallet...';
      walletStatus.className = 'dashboard-status';

      const address = addressInput.value.trim();
      if (!address) {
        walletStatus.textContent = 'Enter a Dogecoin wallet address.';
        walletStatus.className = 'dashboard-status error';
        return;
      }

      try {
        const response = await fetch('/api/balance?address=' + encodeURIComponent(address));
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.message || 'Failed to look up wallet balance');
        }

        setText('balance-current', payload.current);
        setText('balance-available', payload.available);
        setText('balance-incoming', payload.incoming);
        setText('balance-outgoing', payload.outgoing);
        walletResults.hidden = false;
        walletStatus.textContent = 'Balance loaded.';
        walletStatus.className = 'dashboard-status';
      } catch (error) {
        walletStatus.textContent = error.message || 'Failed to look up wallet balance.';
        walletStatus.className = 'dashboard-status error';
      }
    });

    loadDashboardStats();
  })();
</script>
`+htmlFooter, data)
}

// HandleDashboardBalance serves the dashboard balance lookup JSON.
func (s *Server) HandleDashboardBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET is allowed")
		return
	}

	s.serveBalanceJSON(w, r)
}

// HandleDashboardStats serves the public dashboard metrics JSON.
func (s *Server) HandleDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "method-not-allowed", "Only GET is allowed")
		return
	}

	height, err := s.indexerStore.CurrentHeight(r.Context())
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "database-error", "Failed to load indexer height")
		return
	}

	response := DashboardStatsResponse{
		Height:    height,
		Available: false,
		Message:   "Usage statistics are currently unavailable.",
	}

	if !s.hasAuthStore() {
		s.sendJSON(w, http.StatusOK, response)
		return
	}

	stats, err := s.authStore.WithCtx(r.Context()).GetDashboardStats()
	if err != nil {
		log.Printf("[Dogelytics] failed to load dashboard stats: %v", err)
		s.sendJSON(w, http.StatusOK, response)
		return
	}

	response.Available = true
	response.Message = ""
	response.TotalWalletsChecked = stats.TotalWalletsChecked
	response.WalletsCheckedLast24h = stats.WalletsCheckedLast24h
	response.UniqueWalletsLast24h = stats.UniqueWalletsLast24h
	s.sendJSON(w, http.StatusOK, response)
}

func uiURLForPort(r *http.Request, port int) string {
	host := r.Host
	if splitHost, _, err := net.SplitHostPort(r.Host); err == nil {
		host = splitHost
	}

	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}

	return "http://" + net.JoinHostPort(host, portString(port))
}

func portString(port int) string {
	return strconv.Itoa(port)
}
