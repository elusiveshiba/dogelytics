package server

import "net/http"

// HandleGETDocs renders the public API documentation page.
func (s *Server) HandleGETDocs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	renderTemplate(w, htmlHeader+`
<style>
.docs-shell {
  display: grid;
  gap: 12px;
}
.docs-card {
  background: #c0c0c0;
  border-top: 2px solid #fff;
  border-left: 2px solid #fff;
  border-right: 2px solid #808080;
  border-bottom: 2px solid #808080;
  padding: 8px;
}
.docs-card p,
.docs-card li {
  line-height: 1.5;
}
.docs-card ul {
  margin-left: 18px;
}
.docs-code {
  background: #fff;
  border-top: 1px solid #808080;
  border-left: 1px solid #808080;
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
  border: 1px solid #808080;
  padding: 5px;
  text-align: left;
  vertical-align: top;
}
.docs-table th {
  background: #000080;
  color: #fff;
}
.docs-nav {
  margin-top: 4px;
}
.dashboard-button {
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
.dashboard-button:active {
  border-top-color: #000;
  border-left-color: #000;
  border-right-color: #fff;
  border-bottom-color: #fff;
}
@media (max-width: 640px) {
  .dashboard-button {
    left: 12px;
    bottom: 12px;
  }
}
</style>
<div class="docs-shell">
  <div>
    <h1>Dogelytics API Docs</h1>
    <p style="text-align: center;">A small public API for Dogecoin wallet balances and conversion rates.</p>
  </div>

  <div class="docs-card">
    <h2>Base URL</h2>
    <p>Use the public API host for external requests:</p>
    <pre class="docs-code"><code>https://api.dogelytics.com</code></pre>
  </div>

  <div class="docs-card">
    <h2>GET /balance</h2>
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

  <div class="docs-card">
    <h2>GET /conversion</h2>
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
    <p>Rates are sourced from CoinGecko and cached locally for one hour per currency. Dogelytics only refreshes from CoinGecko when a currency is missing from the cache or the cached row is older than one hour.</p>
  </div>

  <div class="docs-card">
    <h2>API Keys</h2>
    <p>API keys are optional. Requests without a key are still allowed, but a valid key gets the higher API-key rate limit.</p>
    <pre class="docs-code"><code>curl \
  -H "Authorization: Bearer YOUR_API_KEY" \
  "https://api.dogelytics.com/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8"</code></pre>
    <p>You can also send the key with <code>X-Api-Key: YOUR_API_KEY</code>.</p>
  </div>

  <div class="docs-card">
    <h2>GET /health</h2>
    <p>Returns service status and the current indexed block height.</p>
    <pre class="docs-code"><code>curl "https://api.dogelytics.com/health"</code></pre>
    <pre class="docs-code"><code>{
  "ok": true,
  "height": 5900000
}</code></pre>
  </div>

  <div class="docs-card">
    <h2>Errors and Limits</h2>
    <table class="docs-table">
      <tr><th>Status</th><th>Error</th><th>Meaning</th></tr>
      <tr><td>400</td><td><code>missing-parameter</code></td><td>The <code>address</code> query parameter is missing.</td></tr>
      <tr><td>400</td><td><code>invalid-address</code></td><td>The address is not a valid Dogecoin address.</td></tr>
      <tr><td>400</td><td><code>invalid-currency</code></td><td>The currency code format is invalid.</td></tr>
      <tr><td>400</td><td><code>unsupported-currency</code></td><td>The requested currency is not available from CoinGecko.</td></tr>
      <tr><td>401</td><td><code>invalid-api-key</code></td><td>The supplied API key is invalid, expired, or revoked.</td></tr>
      <tr><td>429</td><td><code>rate-limit-exceeded</code></td><td>The rate limit was exceeded. The response message includes when to try again.</td></tr>
      <tr><td>502</td><td><code>conversion-source-error</code></td><td>Dogelytics could not refresh a conversion rate from CoinGecko.</td></tr>
      <tr><td>503</td><td><code>conversion-cache-unavailable</code></td><td>The local conversion cache is unavailable.</td></tr>
    </table>
  </div>
</div>
<a class="dashboard-button" href="/">dashboard</a>
`+htmlFooter, nil)
}
