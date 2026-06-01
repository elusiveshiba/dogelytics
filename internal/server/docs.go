package server

import (
	"net/http"
	"strings"
)

// HandleGETDocs renders the public API documentation page.
func (s *Server) HandleGETDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		ShowTipsWidget bool
		TipsAddress    string
	}{
		ShowTipsWidget: s.config.EnableDashboardTips && strings.TrimSpace(s.config.DashboardTipsAddress) != "",
		TipsAddress:    s.config.DashboardTipsAddress,
	}

	renderTemplate(w, htmlHead+`
<style>
body {
  align-items: flex-start;
  overflow-x: hidden;
}
.docs-desktop {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  max-width: 900px;
  width: 100%;
  position: relative;
  z-index: 1;
}
.docs-window {
  width: 100%;
}
.title-coin {
  display: block;
  width: 14px;
  height: 14px;
}
.docs-window p,
.docs-window li {
  line-height: 1.5;
}
.docs-window ul {
  margin-left: 18px;
}
.docs-code {
  background: #fff;
  border-top: 1px solid #a08030;
  border-left: 1px solid #a08030;
  border-right: 1px solid #fff;
  border-bottom: 1px solid #fff;
  font-family: "Courier New", monospace;
  font-size: 11px;
  margin-top: 6px;
  overflow-x: auto;
  padding: 8px;
  white-space: pre;
}
.docs-table {
  border-collapse: collapse;
  width: 100%;
}
.docs-table th,
.docs-table td {
  border: 1px solid #a08030;
  padding: 5px;
  text-align: left;
  vertical-align: top;
}
.docs-table th {
  background: #e8d89c;
  color: #000;
}
.docs-nav {
  margin-top: 4px;
}
.dashboard-button {
  display: inline-block;
  background: #f0e6c8;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #3a3020;
  border-bottom: 2px solid #3a3020;
  box-shadow: inset 1px 1px 0 #f8f0d8, inset -1px -1px 0 #a08830;
  color: #000;
  font-weight: bold;
  padding: 5px 8px;
  text-decoration: none;
}
.dashboard-button:active {
  border-top-color: #3a3020;
  border-left-color: #3a3020;
  border-right-color: #fff;
  border-bottom-color: #fff;
}
.dashboard-actions {
  display: contents;
}
.tips-widget {
  position: fixed;
  right: 16px;
  bottom: 16px;
  z-index: 20;
  width: 175px;
  background: #f0e6c8;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #3a3020;
  border-bottom: 2px solid #3a3020;
  box-shadow: inset 1px 1px 0 #f8f0d8, inset -1px -1px 0 #a08830;
  padding: 2px;
}
.tips-widget.minimised {
  width: auto;
}
.tips-title {
  background: linear-gradient(to right, #c8a951, #e6c85a);
  color: #000;
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
  padding: 0;
  margin: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
}
.tips-toggle:active {
  width: 20px;
  min-width: 20px;
  height: 18px;
  padding: 0;
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
  border-top: 1px solid #a08030;
  border-left: 1px solid #a08030;
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
  border-top: 1px solid #a08030;
  border-left: 1px solid #a08030;
  border-right: 1px solid #fff;
  border-bottom: 1px solid #fff;
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
  display: block;
}
.tips-copy-button {
  flex: 0 0 18px;
  width: 18px;
  min-width: 18px;
  height: 18px;
  margin: 0;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.tips-copy-button:active {
  width: 18px;
  min-width: 18px;
  height: 18px;
  margin: 0;
  padding: 0;
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
.dashboard-status {
  min-height: 18px;
  margin-top: 8px;
  color: #8b6914;
}
.dashboard-status.error {
  color: #8b0000;
}
@media (max-width: 640px) {
  .dashboard-actions {
    display: flex;
    align-items: flex-end;
    justify-content: center;
    margin-top: 12px;
    padding-bottom: 12px;
  }
  .tips-widget {
    position: static;
    flex: 0 1 auto;
    width: auto;
    margin: 0 auto;
  }
}
</style>
<body>
<div class="docs-desktop">
  <div class="container docs-window">
    <div class="window-title">
      <span>Dogelytics API Docs</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <h1>Dogelytics API Docs</h1>
      <p style="text-align: center;">A small public API for Dogecoin wallet balances and conversion rates.</p>
      <div class="links">
        <a class="dashboard-button" href="/">dashboard</a>
      </div>
    </div>
  </div>

  <div class="container docs-window">
    <div class="window-title">
      <span>Base URL</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <p>Use the public API host for external requests:</p>
      <pre class="docs-code"><code>https://api.dogelytics.com</code></pre>
    </div>
  </div>

  <div class="container docs-window">
    <div class="window-title">
      <span>API Keys</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <p>API keys are optional. Requests without a key are still allowed, but a valid key gets the higher API-key rate limit.</p>
      <pre class="docs-code"><code>curl \
  -H "Authorization: Bearer YOUR_API_KEY" \
  "https://api.dogelytics.com/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8"</code></pre>
      <p>You can also send the key with <code>X-Api-Key: YOUR_API_KEY</code>.</p>
    </div>
  </div>

  <div class="container docs-window">
    <div class="window-title">
      <span>GET /balance</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <p>Returns the balance for a Dogecoin address. Values are decimal DOGE strings.</p>
      <table class="docs-table">
        <tr><th>Parameter</th><th>Required</th><th>Description</th></tr>
        <tr><td><code>address</code></td><td>yes</td><td>Dogecoin address to check.</td></tr>
      </table>
      <pre class="docs-code"><code>curl "https://api.dogelytics.com/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8"</code></pre>
      <pre class="docs-code"><code>{
  "incoming": "0.00000000",
  "available": "123.45678900",
  "outgoing": "0.00000000",
  "current": "123.45678900"
}</code></pre>
    </div>
  </div>

  <div class="container docs-window">
    <div class="window-title">
      <span>GET /conversion</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <p>Returns the current DOGE conversion rate for one target currency code.</p>
      <table class="docs-table">
        <tr><th>Parameter</th><th>Required</th><th>Description</th></tr>
        <tr><td><code>currency</code></td><td>yes</td><td>Lowercase or uppercase currency code such as <code>usd</code>, <code>aud</code>, or <code>eur</code>.</td></tr>
      </table>
      <pre class="docs-code"><code>curl "https://api.dogelytics.com/conversion?currency=usd"</code></pre>
      <pre class="docs-code"><code>{
  "currency": "usd",
  "rate": "0.23543210",
  "source": "coingecko",
  "cached": false,
  "fetched_at": "2026-05-28T01:44:00Z",
  "coingecko_updated_at": "2026-05-28T01:43:20Z"
}</code></pre>
      <p>Rates are sourced from CoinGecko and cached locally for one hour per currency. Dogelytics only refreshes from CoinGecko when a currency is missing from the cache or the cached row is older than one hour. The dashboard uses the same conversion data through its same-origin <code>/api/conversion</code> route.</p>
    </div>
  </div>

  <div class="container docs-window">
    <div class="window-title">
      <span>GET /health</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <p>Returns service status and the raw sync heights used to compare the indexer, local Dogecoin Core node, and best known blockchain height.</p>
      <pre class="docs-code"><code>curl "https://api.dogelytics.com/health"</code></pre>
      <pre class="docs-code"><code>{
  "ok": true,
  "indexer_height": 5900000,
  "core_height": 6224976,
  "blockchain_height": 6224976
}</code></pre>
      <p><code>indexer_height</code> is the last block indexed by Dogelytics, <code>core_height</code> is the local Dogecoin Core block height, and <code>blockchain_height</code> is Core's best known chain tip from headers.</p>
      <p>If Dogecoin Core RPC is not configured or is unavailable, this endpoint still returns <code>ok</code> and <code>indexer_height</code>, but omits <code>core_height</code> and <code>blockchain_height</code>.</p>
    </div>
  </div>

  <div class="container docs-window">
    <div class="window-title">
      <span>Dashboard Helper APIs</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <p>The dashboard host exposes same-origin helper routes for browser use. <code>/api/balance</code> and <code>/api/conversion</code> mirror the public API responses, while <code>/api/dashboard-stats</code> returns public usage counters and sync metrics for the dashboard.</p>
      <pre class="docs-code"><code>GET /api/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8
GET /api/conversion?currency=usd
GET /api/dashboard-stats</code></pre>
    </div>
  </div>

  <div class="container docs-window">
    <div class="window-title">
      <span>Errors and Limits</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <table class="docs-table">
        <tr><th>Status</th><th>Error</th><th>Meaning</th></tr>
        <tr><td>400</td><td><code>missing-parameter</code></td><td>The <code>address</code> query parameter is missing.</td></tr>
        <tr><td>400</td><td><code>invalid-address</code></td><td>The address is not a valid Dogecoin address.</td></tr>
        <tr><td>400</td><td><code>unsupported-address</code></td><td>The address type is not supported.</td></tr>
        <tr><td>400</td><td><code>invalid-currency</code></td><td>The currency code format is invalid.</td></tr>
        <tr><td>400</td><td><code>unsupported-currency</code></td><td>The requested currency is not available from CoinGecko.</td></tr>
        <tr><td>401</td><td><code>invalid-api-key</code></td><td>The supplied API key is invalid, expired, or revoked.</td></tr>
        <tr><td>429</td><td><code>rate-limit-exceeded</code></td><td>The rate limit was exceeded. The response message includes when to try again.</td></tr>
        <tr><td>500</td><td><code>database-error</code></td><td>Dogelytics could not read indexer data.</td></tr>
        <tr><td>502</td><td><code>conversion-source-error</code></td><td>Dogelytics could not refresh a conversion rate from CoinGecko.</td></tr>
        <tr><td>503</td><td><code>conversion-cache-unavailable</code></td><td>The local conversion cache is unavailable.</td></tr>
      </table>
    </div>
  </div>
<div class="dashboard-actions">
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
    <div id="tips-copy-row" class="tips-address-row">
      <code id="tips-address" class="tips-address">{{.TipsAddress}}</code>
      <button id="tips-copy-button" type="button" class="tips-copy-button" title="Copy tip address" aria-label="Copy tip address">
        <span class="tips-copy-icon" aria-hidden="true"></span>
      </button>
    </div>
    <div id="tips-status" class="dashboard-status" aria-live="polite"></div>
  </div>
</div>
{{end}}
</div>
</div>
{{if .ShowTipsWidget}}
<script>
  (function() {
    const widget = document.getElementById('tips-widget');
    const toggle = document.getElementById('tips-toggle');
    const copyButton = document.getElementById('tips-copy-button');
    const address = document.getElementById('tips-address');
    const status = document.getElementById('tips-status');
    if (!widget || !toggle || !copyButton || !address || !status) return;

    function scrollTipsIntoView() {
      const rect = widget.getBoundingClientRect();
      const viewportHeight = window.innerHeight || document.documentElement.clientHeight;
      if (rect.bottom <= viewportHeight) return;
      window.scrollBy({
        top: rect.bottom - viewportHeight + 16,
        behavior: 'smooth',
      });
    }

    function applyMinimised(minimised) {
      widget.classList.toggle('minimised', minimised);
      toggle.textContent = minimised ? '+' : '-';
      toggle.setAttribute('aria-label', minimised ? 'Expand tips' : 'Minimise tips');
      try {
        localStorage.setItem('dogelytics_tips_minimised', minimised ? 'true' : 'false');
      } catch (_) {}
      if (!minimised) {
        requestAnimationFrame(scrollTipsIntoView);
      }
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

    copyButton.addEventListener('click', copyTipAddress);
  })();
</script>
{{end}}
`+memeNumberAssets+`
</body>
</html>
`, data)
}
