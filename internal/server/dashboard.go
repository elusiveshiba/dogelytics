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
		ShowAdminLink  bool
		AdminURL       string
		ShowTipsWidget bool
		TipsAddress    string
	}{
		ShowAdminLink:  s.config.EnableAdminUI,
		AdminURL:       uiURLForPort(r, s.config.AdminUIPort),
		ShowTipsWidget: s.config.EnableDashboardTips && strings.TrimSpace(s.config.DashboardTipsAddress) != "",
		TipsAddress:    s.config.DashboardTipsAddress,
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
.docs-button {
  position: fixed;
  left: 16px;
  bottom: 16px;
  z-index: 20;
  display: inline-block;
  background: #c0c0c0;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #000;
  border-bottom: 2px solid #000;
  box-shadow: inset 1px 1px 0 #dfdfdf, inset -1px -1px 0 #808080;
  color: #000;
  font-weight: bold;
  padding: 5px 8px;
  text-decoration: none;
}
.docs-button:active {
  border-top-color: #000;
  border-left-color: #000;
  border-right-color: #fff;
  border-bottom-color: #fff;
}
.tips-widget {
  position: fixed;
  right: 16px;
  bottom: 16px;
  z-index: 20;
  width: 175px;
  background: #c0c0c0;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #000;
  border-bottom: 2px solid #000;
  box-shadow: inset 1px 1px 0 #dfdfdf, inset -1px -1px 0 #808080;
  padding: 2px;
}
.tips-widget.minimised {
  width: auto;
}
.tips-title {
  background: linear-gradient(to right, #000080, #1084d0);
  color: #fff;
  padding: 2px 4px;
  font-weight: bold;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}
.tips-toggle {
  width: 20px;
  min-width: 20px;
  height: 18px;
  padding: 0 5px;
  margin: 0;
  line-height: 14px;
}
.tips-toggle:active {
  width: 20px;
  min-width: 20px;
  height: 18px;
  padding: 0 5px;
}
.tips-body {
  padding: 6px;
}
.tips-body p {
  margin-bottom: 5px;
}
.tips-qr {
  display: flex;
  justify-content: center;
  background: #fff;
  border-top: 1px solid #808080;
  border-left: 1px solid #808080;
  border-right: 1px solid #fff;
  border-bottom: 1px solid #fff;
  margin-bottom: 5px;
  padding: 4px;
}
.tips-address-row {
  display: flex;
  align-items: center;
  gap: 4px;
  background: #fff;
  border-top: 1px solid #808080;
  border-left: 1px solid #808080;
  border-right: 1px solid #fff;
  border-bottom: 1px solid #fff;
  cursor: pointer;
  padding: 4px;
}
.tips-widget.minimised .tips-body {
  display: none;
}
.tips-address {
  flex: 1;
  min-width: 0;
  background: transparent;
  padding: 0;
  word-break: break-all;
  font-size: 10px;
}
.tips-copy-icon {
  position: relative;
  flex: 0 0 13px;
  width: 13px;
  height: 13px;
}
.tips-copy-icon::before,
.tips-copy-icon::after {
  content: "";
  position: absolute;
  width: 8px;
  height: 8px;
  background: #fff;
  border: 1px solid #000;
}
.tips-copy-icon::before {
  left: 1px;
  top: 4px;
}
.tips-copy-icon::after {
  left: 4px;
  top: 1px;
}
@media (max-width: 640px) {
  .tips-widget {
    left: 12px;
    right: 12px;
    bottom: 12px;
    width: auto;
  }
  .docs-button {
    left: 12px;
    bottom: 12px;
  }
}
</style>
<div class="dashboard-shell">
  <div>
    <h1>Dogelytics Dashboard</h1>
    <p style="text-align: center;">
      Check a Dogecoin wallet balance.
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
    <h2>Stats</h2>
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

<a class="docs-button" href="/docs">docs</a>

{{if .ShowTipsWidget}}
<script type="module" src="https://fetch.dogecoin.org/doge-qr.js"></script>
<div id="tips-widget" class="tips-widget">
  <div class="tips-title">
    <span>Such coffee?</span>
    <button id="tips-toggle" type="button" class="tips-toggle" aria-label="Minimise tips">-</button>
  </div>
  <div class="tips-body">
    <p>Enjoying Dogelytics? Send a tip to:</p>
    <div class="tips-qr">
      <doge-qr address="{{.TipsAddress}}" size="sm" background="#fff" fill="#000"></doge-qr>
    </div>
    <div id="tips-copy-row" class="tips-address-row" role="button" tabindex="0" title="Copy tip address">
      <code id="tips-address" class="tips-address">{{.TipsAddress}}</code>
      <span class="tips-copy-icon" aria-hidden="true"></span>
    </div>
    <div id="tips-status" class="dashboard-status" aria-live="polite"></div>
  </div>
</div>
{{end}}

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

    function formatInteger(value) {
      const text = String(value);
      if (!/^-?\d+$/.test(text)) return text;
      const sign = text.startsWith('-') ? '-' : '';
      const digits = sign ? text.slice(1) : text;
      return sign + digits.replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    }

    function formatDogeAmount(value) {
      const text = String(value);
      const parts = text.split('.');
      const whole = formatInteger(parts[0]);
      if (parts.length === 1) return whole;

      const trimmedFraction = parts[1].replace(/0+$/, '');
      const fraction = (trimmedFraction || '00').padEnd(2, '0');
      return whole + '.' + fraction;
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

        setText('stat-total-wallets', formatInteger(payload.total_wallets_checked));
        setText('stat-wallets-24h', formatInteger(payload.wallets_checked_last_24h));
        setText('stat-unique-wallets-24h', formatInteger(payload.unique_wallets_last_24h));
        setText('stat-height', formatInteger(payload.height));

        if (payload.available) {
          statsStatus.textContent = '';
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

        setText('balance-current', 'Ð' + formatDogeAmount(payload.current));
        setText('balance-available', 'Ð' + formatDogeAmount(payload.available));
        setText('balance-incoming', 'Ð' + formatDogeAmount(payload.incoming));
        setText('balance-outgoing', 'Ð' + formatDogeAmount(payload.outgoing));
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
{{if .ShowTipsWidget}}
<script>
  (function() {
    const widget = document.getElementById('tips-widget');
    const toggle = document.getElementById('tips-toggle');
    const copyRow = document.getElementById('tips-copy-row');
    const address = document.getElementById('tips-address');
    const status = document.getElementById('tips-status');
    if (!widget || !toggle || !copyRow || !address || !status) return;

    function applyMinimised(minimised) {
      widget.classList.toggle('minimised', minimised);
      toggle.textContent = minimised ? '+' : '-';
      toggle.setAttribute('aria-label', minimised ? 'Expand tips' : 'Minimise tips');
      try {
        localStorage.setItem('dogelytics_tips_minimised', minimised ? 'true' : 'false');
      } catch (_) {}
    }

    applyMinimised(true);

    toggle.addEventListener('click', function() {
      applyMinimised(!widget.classList.contains('minimised'));
    });

    async function copyTipAddress() {
      const text = address.textContent || '';
      try {
        await navigator.clipboard.writeText(text);
        status.textContent = 'Copied tip address.';
        status.className = 'dashboard-status';
      } catch (_) {
        status.textContent = 'Copy failed. Select the address manually.';
        status.className = 'dashboard-status error';
      }
    }

    copyRow.addEventListener('click', copyTipAddress);
    copyRow.addEventListener('keydown', function(event) {
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        copyTipAddress();
      }
    });
  })();
</script>
{{end}}
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
