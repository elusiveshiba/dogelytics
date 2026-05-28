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

	renderTemplate(w, htmlHead+`
<style>
body {
  align-items: flex-start;
  overflow-x: hidden;
}
.dashboard-desktop {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  max-width: 900px;
  width: 100%;
  position: relative;
  z-index: 1;
}
.dashboard-window {
  width: 100%;
}
.title-coin {
  display: block;
  width: 14px;
  height: 14px;
}
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 8px;
}
.dashboard-section {
  display: grid;
  gap: 8px;
  margin-top: 10px;
}
.dashboard-section h3 {
  margin: 0;
}
.wallet-checker-form {
  display: grid;
  gap: 10px;
}
.wallet-checker-label {
  display: grid;
  gap: 6px;
  font-weight: bold;
  font-size: 14px;
}
.wallet-checker-input {
  display: block;
  width: 100%;
  box-sizing: border-box;
  min-height: 50px;
  padding: 10px 12px;
  border-width: 2px;
  font-size: 24px;
  line-height: 1.2;
  font-family: "Courier New", monospace;
  font-weight: bold;
}
.wallet-checker-actions {
  display: flex;
  align-items: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}
.wallet-checker-currency-label {
  display: grid;
  gap: 3px;
  margin-bottom: 0;
}
.wallet-checker-select-wrap {
  position: relative;
  display: inline-block;
}
.wallet-checker-currency {
  min-width: 120px;
  padding: 4px 24px 4px 6px;
  border-top: 1px solid #a08030;
  border-left: 1px solid #a08030;
  border-right: 1px solid #fff;
  border-bottom: 1px solid #fff;
  border-radius: 0;
  box-shadow: inset -1px -1px 0 #f0e6c8, inset 1px 1px 0 #000;
  background: #fff;
  color: #000;
  font-size: 11px;
  font-family: inherit;
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
}
.wallet-checker-select-arrow {
  position: absolute;
  top: 1px;
  right: 1px;
  bottom: 1px;
  width: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0e6c8;
  border-top: 1px solid #fff;
  border-left: 1px solid #fff;
  border-right: 1px solid #a08030;
  border-bottom: 1px solid #a08030;
  box-shadow: inset 1px 1px 0 #f8f0d8;
  color: #000;
  font-size: 9px;
  pointer-events: none;
}
.wallet-checker-submit {
  padding: 6px 12px;
}
.dashboard-stat-card {
  background: #fff;
  border-top: 1px solid #a08030;
  border-left: 1px solid #a08030;
  border-right: 1px solid #fff;
  border-bottom: 1px solid #fff;
  padding: 8px;
}
.dashboard-stat-label {
  color: #666;
  font-size: 10px;
  margin-bottom: 4px;
}
.dashboard-stat-value {
  font-size: 16px;
  font-weight: bold;
  line-height: 1.1;
}
.dashboard-stat-progress {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 6px;
}
.dashboard-stat-detail {
  color: #666;
  font-size: 10px;
  font-weight: normal;
}
.dashboard-balance-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 8px;
  margin-top: 10px;
}
.dashboard-balance-item {
  background: #fff;
  border-top: 1px solid #a08030;
  border-left: 1px solid #a08030;
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
  font-size: 14px;
  font-weight: bold;
}
.dashboard-balance-detail {
  color: #666;
  font-size: 10px;
  margin-top: 3px;
}
.dashboard-status {
  min-height: 18px;
  margin-top: 8px;
  color: #8b6914;
}
.dashboard-status.error {
  color: #8b0000;
}
.dashboard-actions {
  display: contents;
}
.docs-button {
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
.docs-button:active {
  border-top-color: #3a3020;
  border-left-color: #3a3020;
  border-right-color: #fff;
  border-bottom-color: #fff;
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
  .dashboard-actions {
    display: flex;
    align-items: flex-end;
    justify-content: flex-end;
    margin-top: 12px;
    padding-bottom: 12px;
  }
  .tips-widget {
    position: static;
    flex: 0 1 auto;
    width: auto;
    margin-left: auto;
  }
}
</style>
<body>
<div class="dashboard-desktop">
  <div class="container dashboard-window">
    <div class="window-title">
      <span>Dogelytics</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <h1>Dogelytics Dashboard</h1>
      <p style="text-align: center;">
        Check a Dogecoin wallet balance.
      </p>
      <div class="links">
        <a class="docs-button" href="/docs">docs</a>
        {{if .ShowAdminLink}}
        <a href="{{.AdminURL}}">Open admin UI</a>
        {{end}}
      </div>
    </div>
  </div>

  <div class="container dashboard-window">
    <div class="window-title">
      <span>Wallet Checker</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <form id="wallet-checker-form" class="wallet-checker-form">
        <label class="wallet-checker-label">
          Dogecoin wallet address
          <input class="wallet-checker-input" type="text" id="wallet-address" name="address" placeholder="D..." autocomplete="off" required>
        </label>
        <div class="wallet-checker-actions">
          <button class="wallet-checker-submit" type="submit">Check Balance</button>
          <label class="wallet-checker-currency-label">
            Currency
            <span class="wallet-checker-select-wrap">
              <select class="wallet-checker-currency" id="wallet-currency" name="currency">
                <option value="aed">AED (AED)</option>
                <option value="aud">AUD ($)</option>
                <option value="brl">BRL (R$)</option>
                <option value="cad">CAD ($)</option>
                <option value="chf">CHF (CHF)</option>
                <option value="cny">CNY (¥)</option>
                <option value="doge">DOGE (Ð)</option>
                <option value="eur">EUR (€)</option>
                <option value="gbp">GBP (£)</option>
                <option value="hkd">HKD ($)</option>
                <option value="idr">IDR (Rp)</option>
                <option value="inr">INR (₹)</option>
                <option value="jpy">JPY (¥)</option>
                <option value="krw">KRW (₩)</option>
                <option value="mxn">MXN ($)</option>
                <option value="nok">NOK (kr)</option>
                <option value="nzd">NZD ($)</option>
                <option value="pln">PLN (zl)</option>
                <option value="sek">SEK (kr)</option>
                <option value="sgd">SGD ($)</option>
                <option value="try">TRY (₺)</option>
                <option value="usd" selected>USD ($)</option>
                <option value="zar">ZAR (R)</option>
              </select>
              <span class="wallet-checker-select-arrow" aria-hidden="true">v</span>
            </span>
          </label>
        </div>
      </form>
      <div id="wallet-status" class="dashboard-status" aria-live="polite"></div>
      <div id="wallet-results" class="dashboard-balance-grid" hidden>
        <div class="dashboard-balance-item">
          <div class="dashboard-balance-label">Current</div>
          <div id="balance-current" class="dashboard-balance-value">-</div>
          <div id="balance-current-converted" class="dashboard-balance-detail"></div>
        </div>
        <div class="dashboard-balance-item">
          <div class="dashboard-balance-label">Available</div>
          <div id="balance-available" class="dashboard-balance-value">-</div>
          <div id="balance-available-converted" class="dashboard-balance-detail"></div>
        </div>
        <div class="dashboard-balance-item">
          <div class="dashboard-balance-label">Incoming</div>
          <div id="balance-incoming" class="dashboard-balance-value">-</div>
          <div id="balance-incoming-converted" class="dashboard-balance-detail"></div>
        </div>
        <div class="dashboard-balance-item">
          <div class="dashboard-balance-label">Outgoing</div>
          <div id="balance-outgoing" class="dashboard-balance-value">-</div>
          <div id="balance-outgoing-converted" class="dashboard-balance-detail"></div>
        </div>
      </div>
    </div>
  </div>

  <div class="container dashboard-window">
    <div class="window-title">
      <span>Stats</span>
      <span><img class="title-coin" src="/img/dogecoin-doge-logo.svg" alt="" aria-hidden="true"></span>
    </div>
    <div class="window-content">
      <div class="dashboard-section">
        <h3>Wallets</h3>
        <div class="dashboard-grid">
          <div class="dashboard-stat-card">
            <div class="dashboard-stat-label">Total wallets checked (24h)</div>
            <div id="stat-wallets-24h" class="dashboard-stat-value">-</div>
          </div>
          <div class="dashboard-stat-card">
            <div class="dashboard-stat-label">Unique wallets checked (24h)</div>
            <div id="stat-unique-wallets-24h" class="dashboard-stat-value">-</div>
          </div>
          <div class="dashboard-stat-card">
            <div class="dashboard-stat-label">Total wallets checked</div>
            <div id="stat-total-wallets" class="dashboard-stat-value">-</div>
          </div>
          <div class="dashboard-stat-card">
            <div class="dashboard-stat-label">Unique wallets checked</div>
            <div id="stat-unique-wallets" class="dashboard-stat-value">-</div>
          </div>
        </div>
      </div>

      <div class="dashboard-section">
        <h3>Blockchain</h3>
        <div class="dashboard-grid">
          <div class="dashboard-stat-card">
            <div class="dashboard-stat-label">Indexed Height</div>
            <div class="dashboard-stat-progress">
              <span id="stat-indexed-percent" class="dashboard-stat-value">-</span>
              <span id="stat-indexed-detail" class="dashboard-stat-detail"></span>
            </div>
          </div>
          <div class="dashboard-stat-card">
            <div class="dashboard-stat-label">Dogecoin Core Height</div>
            <div class="dashboard-stat-progress">
              <span id="stat-core-height" class="dashboard-stat-value">-</span>
              <span id="stat-core-detail" class="dashboard-stat-detail"></span>
            </div>
          </div>
        </div>
      </div>
      <div id="stats-status" class="dashboard-status" aria-live="polite"></div>
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
    <div id="tips-copy-row" class="tips-address-row" role="button" tabindex="0" title="Copy tip address">
      <code id="tips-address" class="tips-address">{{.TipsAddress}}</code>
      <span class="tips-copy-icon" aria-hidden="true"></span>
    </div>
    <div id="tips-status" class="dashboard-status" aria-live="polite"></div>
  </div>
</div>
{{end}}
  </div>
</div>

<script>
  (function() {
    const form = document.getElementById('wallet-checker-form');
    const addressInput = document.getElementById('wallet-address');
    const currencySelect = document.getElementById('wallet-currency');
    const walletStatus = document.getElementById('wallet-status');
    const walletResults = document.getElementById('wallet-results');
    const statsStatus = document.getElementById('stats-status');
    const defaultCurrency = 'usd';
    const currencyStorageKey = 'dogelytics_dashboard_currency';
    const currencyMetadata = {
      aed: { ticker: 'AED', symbol: 'AED' },
      aud: { ticker: 'AUD', symbol: '$' },
      brl: { ticker: 'BRL', symbol: 'R$' },
      cad: { ticker: 'CAD', symbol: '$' },
      chf: { ticker: 'CHF', symbol: 'CHF' },
      cny: { ticker: 'CNY', symbol: '¥' },
      doge: { ticker: 'DOGE', symbol: 'Ð' },
      eur: { ticker: 'EUR', symbol: '€' },
      gbp: { ticker: 'GBP', symbol: '£' },
      hkd: { ticker: 'HKD', symbol: '$' },
      idr: { ticker: 'IDR', symbol: 'Rp' },
      inr: { ticker: 'INR', symbol: '₹' },
      jpy: { ticker: 'JPY', symbol: '¥' },
      krw: { ticker: 'KRW', symbol: '₩' },
      mxn: { ticker: 'MXN', symbol: '$' },
      nok: { ticker: 'NOK', symbol: 'kr' },
      nzd: { ticker: 'NZD', symbol: '$' },
      pln: { ticker: 'PLN', symbol: 'zl' },
      sek: { ticker: 'SEK', symbol: 'kr' },
      sgd: { ticker: 'SGD', symbol: '$' },
      try: { ticker: 'TRY', symbol: '₺' },
      usd: { ticker: 'USD', symbol: '$' },
      zar: { ticker: 'ZAR', symbol: 'R' },
    };
    const regionCurrency = {
      AE: 'aed',
      AT: 'eur',
      AU: 'aud',
      BE: 'eur',
      BR: 'brl',
      CA: 'cad',
      CH: 'chf',
      CN: 'cny',
      CY: 'eur',
      DE: 'eur',
      EE: 'eur',
      ES: 'eur',
      FI: 'eur',
      FR: 'eur',
      GB: 'gbp',
      GR: 'eur',
      HK: 'hkd',
      ID: 'idr',
      IE: 'eur',
      IN: 'inr',
      IT: 'eur',
      JP: 'jpy',
      KR: 'krw',
      LT: 'eur',
      LU: 'eur',
      LV: 'eur',
      MT: 'eur',
      MX: 'mxn',
      NL: 'eur',
      NO: 'nok',
      NZ: 'nzd',
      PL: 'pln',
      PT: 'eur',
      SE: 'sek',
      SG: 'sgd',
      SI: 'eur',
      SK: 'eur',
      TR: 'try',
      US: 'usd',
      ZA: 'zar',
    };

    function setText(id, value) {
      const element = document.getElementById(id);
      if (element) element.textContent = value;
    }

    function setStatProgress(mainId, detailId, mainValue, detailValue) {
      setText(mainId, mainValue);
      setText(detailId, detailValue);
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

    function normaliseCurrency(value) {
      const currency = value ? String(value).trim().toLowerCase() : '';
      return currency || defaultCurrency;
    }

    function currencyInfo(currency) {
      const normalised = normaliseCurrency(currency);
      return currencyMetadata[normalised] || {
        ticker: normalised.toUpperCase(),
        symbol: normalised.toUpperCase(),
      };
    }

    function formatFiatAmount(value) {
      const number = Number(value);
      if (!Number.isFinite(number)) return null;

      const fixed = number.toFixed(2);
      const parts = fixed.split('.');
      return formatInteger(parts[0]) + '.' + parts[1];
    }

    function formatConvertedBalance(balanceValue, rateValue, currency) {
      const balance = Number(balanceValue);
      const rate = Number(rateValue);
      if (!Number.isFinite(balance) || !Number.isFinite(rate)) return null;

      const formatted = formatFiatAmount(balance * rate);
      if (!formatted) return null;
      return currencyInfo(currency).symbol + formatted;
    }

    function conversionUnavailableLine(currency) {
      return currencyInfo(currency).ticker + ' unavailable';
    }

    function formatPercent(value) {
      const number = Number(value);
      if (!Number.isFinite(number)) return '-';
      return number.toFixed(2).replace(/\.?0+$/, '') + '%';
    }

    function formatIndexedProgress(indexedHeight, blockchainHeight) {
      const indexed = Number(indexedHeight);
      const total = Number(blockchainHeight);
      if (!Number.isFinite(indexed) || !Number.isFinite(total) || total <= 0) {
        return { main: '-', detail: '' };
      }

      const percent = Math.min((indexed / total) * 100, 100);
      return {
        main: formatPercent(percent),
        detail: formatInteger(indexedHeight) + '/' + formatInteger(blockchainHeight),
      };
    }

    function selectedCurrency() {
      return normaliseCurrency(currencySelect && currencySelect.value);
    }

    function supportedCurrency(value) {
      const currency = value ? String(value).trim().toLowerCase() : '';
      return currencyMetadata[currency] ? currency : '';
    }

    function regionFromLocale(locale) {
      const text = locale ? String(locale) : '';
      if (!text) return '';

      if (window.Intl && Intl.Locale) {
        try {
          const parsed = new Intl.Locale(text);
          if (parsed.region) return parsed.region.toUpperCase();
        } catch (_) {}
      }

      const parts = text.replace('_', '-').split('-');
      for (let i = parts.length - 1; i >= 1; i--) {
        if (/^[A-Za-z]{2}$/.test(parts[i])) return parts[i].toUpperCase();
      }
      return '';
    }

    function detectedCurrency() {
      const locales = [];
      if (navigator.languages && navigator.languages.length) {
        locales.push.apply(locales, navigator.languages);
      }
      if (navigator.language) locales.push(navigator.language);
      if (window.Intl && Intl.DateTimeFormat) {
        try {
          locales.push(Intl.DateTimeFormat().resolvedOptions().locale);
        } catch (_) {}
      }

      for (let i = 0; i < locales.length; i++) {
        const currency = regionCurrency[regionFromLocale(locales[i])];
        if (supportedCurrency(currency)) return currency;
      }
      return defaultCurrency;
    }

    function restoreCurrencySelection() {
      if (!currencySelect) return;

      let saved = '';
      try {
        saved = supportedCurrency(localStorage.getItem(currencyStorageKey));
      } catch (_) {}

      const currency = saved || detectedCurrency();
      if (supportedCurrency(currency)) currencySelect.value = currency;
      rememberCurrencySelection(currencySelect.value);
    }

    function rememberCurrencySelection(currency) {
      try {
        localStorage.setItem(currencyStorageKey, supportedCurrency(currency) || defaultCurrency);
      } catch (_) {}
    }

    async function fetchConversion(currency) {
      const response = await fetch('/api/conversion?currency=' + encodeURIComponent(currency));
      const payload = await response.json();
      if (!response.ok) {
        throw new Error(payload.message || 'Failed to load conversion rate');
      }
      return payload;
    }

    function setBalanceResult(name, dogeValue, convertedValue) {
      setText('balance-' + name, 'Ð' + formatDogeAmount(dogeValue));
      setText('balance-' + name + '-converted', convertedValue || '');
    }

    function convertedLine(balanceValue, conversion, currency) {
      if (!conversion || typeof conversion.rate === 'undefined') {
        return conversionUnavailableLine(currency);
      }

      return formatConvertedBalance(
        balanceValue,
        conversion.rate,
        conversion.currency || currency
      ) || conversionUnavailableLine(currency);
    }

    let statsRequestInFlight = false;

    async function loadDashboardStats(options) {
      const autoRefresh = !!(options && options.autoRefresh);
      if (statsRequestInFlight) return;

      statsRequestInFlight = true;
      if (!autoRefresh) {
        statsStatus.textContent = 'Loading dashboard stats...';
        statsStatus.className = 'dashboard-status';
      }

      try {
        const response = await fetch('/api/dashboard-stats');
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.message || 'Failed to load dashboard stats');
        }

        setText('stat-wallets-24h', formatInteger(payload.wallets_checked_last_24h));
        setText('stat-unique-wallets-24h', formatInteger(payload.unique_wallets_last_24h));
        setText('stat-total-wallets', formatInteger(payload.total_wallets_checked));
        setText('stat-unique-wallets', formatInteger(payload.unique_wallets_checked));

        const indexedProgress = formatIndexedProgress(payload.height, payload.blockchain_height);
        setStatProgress('stat-indexed-percent', 'stat-indexed-detail', indexedProgress.main, indexedProgress.detail);

        const coreProgress = formatIndexedProgress(payload.core_height, payload.blockchain_height);
        setStatProgress(
          'stat-core-height',
          'stat-core-detail',
          coreProgress.main,
          coreProgress.detail
        );

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
      } finally {
        statsRequestInFlight = false;
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

      const currency = selectedCurrency();
      rememberCurrencySelection(currency);

      try {
        const response = await fetch('/api/balance?address=' + encodeURIComponent(address));
        const payload = await response.json();
        if (!response.ok) {
          throw new Error(payload.message || 'Failed to look up wallet balance');
        }

        let conversion = null;
        let conversionUnavailable = false;
        try {
          conversion = await fetchConversion(currency);
        } catch (_) {
          conversionUnavailable = true;
        }

        setBalanceResult('current', payload.current, convertedLine(payload.current, conversion, currency));
        setBalanceResult('available', payload.available, convertedLine(payload.available, conversion, currency));
        setBalanceResult('incoming', payload.incoming, convertedLine(payload.incoming, conversion, currency));
        setBalanceResult('outgoing', payload.outgoing, convertedLine(payload.outgoing, conversion, currency));
        walletResults.hidden = false;
        walletStatus.textContent = conversionUnavailable ? 'Balance loaded. Conversion unavailable.' : 'Balance loaded.';
        walletStatus.className = conversionUnavailable ? 'dashboard-status error' : 'dashboard-status';
      } catch (error) {
        walletStatus.textContent = error.message || 'Failed to look up wallet balance.';
        walletStatus.className = 'dashboard-status error';
      }
    });

    restoreCurrencySelection();
    if (currencySelect) {
      currencySelect.addEventListener('change', function() {
        rememberCurrencySelection(selectedCurrency());
      });
    }
    loadDashboardStats();
    window.setInterval(function() {
      loadDashboardStats({ autoRefresh: true });
    }, 60000);
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
</body>
</html>
`, data)
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
	s.populateBlockchainProgress(r, &response)

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
	response.UniqueWalletsChecked = stats.UniqueWalletsChecked
	response.UniqueWalletsLast24h = stats.UniqueWalletsLast24h
	s.sendJSON(w, http.StatusOK, response)
}

func (s *Server) populateBlockchainProgress(r *http.Request, response *DashboardStatsResponse) {
	progress, ok := s.loadChainProgress(r.Context(), response.Height)
	if !ok {
		return
	}

	response.CoreHeight = progress.CoreHeight
	response.BlockchainHeight = progress.BlockchainHeight
	response.IndexedPercent = progress.IndexedPercent
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
