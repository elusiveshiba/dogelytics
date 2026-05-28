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
    <h2>API Keys</h2>
    <p>API keys are optional. Requests without a key are still allowed, but a valid key gets the higher API-key rate limit.</p>
    <pre class="docs-code"><code>curl \
  -H "Authorization: Bearer YOUR_API_KEY" \
  "https://api.dogelytics.com/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8"</code></pre>
    <p>You can also send the key with <code>X-Api-Key: YOUR_API_KEY</code>.</p>
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
    <p>Rates are sourced from CoinGecko and cached locally for one hour per currency. Dogelytics only refreshes from CoinGecko when a currency is missing from the cache or the cached row is older than one hour. The dashboard uses the same conversion data through its same-origin <code>/api/conversion</code> route.</p>
  </div>

  <div class="docs-card">
    <h2>GET /health</h2>
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

  <div class="docs-card">
    <h2>Dashboard Helper APIs</h2>
    <p>The dashboard host exposes same-origin helper routes for browser use. <code>/api/balance</code> and <code>/api/conversion</code> mirror the public API responses, while <code>/api/dashboard-stats</code> returns public usage counters and sync metrics for the dashboard.</p>
    <pre class="docs-code"><code>GET /api/balance?address=DLAznsPDLDRgsVcTFWRMYMG5uH6GddDtv8
GET /api/conversion?currency=usd
GET /api/dashboard-stats</code></pre>
  </div>

  <div class="docs-card">
    <h2>Errors and Limits</h2>
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
<a class="dashboard-button" href="/">dashboard</a>
`+htmlFooter, nil)
}
