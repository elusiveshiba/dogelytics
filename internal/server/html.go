package server

// HTML template constants with embedded styles
const htmlStyles = `
<style>
  * {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }
  
  body {
    font-family: "MS Sans Serif", Tahoma, Arial, sans-serif;
    background: #4a003f;
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    color: #000;
    font-size: 11px;
  }
  
  .container {
    background: #80b1b3;
    border-top: 2px solid #fff;
    border-left: 2px solid #fff;
    border-right: 2px solid #19392c;
    border-bottom: 2px solid #19392c;
    box-shadow: inset 1px 1px 0 #ecf9f2, inset -1px -1px 0 #517b7a;
    padding: 2px;
    max-width: 900px;
    width: 100%;
  }
  
  .window-title {
    background: linear-gradient(to right, #446a65, #80b1b3);
    color: #000;
    padding: 3px 5px;
    font-weight: bold;
    font-size: 11px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2px;
  }
  
  .window-content {
    background: #80b1b3;
    padding: 12px;
  }
  
  h1 {
    color: #000;
    font-size: 16px;
    font-weight: bold;
    margin-bottom: 12px;
    text-align: center;
  }
  
  h2 {
    color: #000;
    font-size: 13px;
    font-weight: bold;
    margin-bottom: 8px;
    padding: 3px 5px;
    background: #5c9082;
  }
  
  h3 {
    color: #000;
    font-size: 11px;
    font-weight: bold;
    margin-top: 12px;
    margin-bottom: 6px;
  }
  
  p {
    margin-bottom: 8px;
    line-height: 1.4;
    color: #000;
    font-size: 11px;
  }
  
  a {
    color: #19392c;
    text-decoration: underline;
  }
  
  a:hover {
    color: #446a65;
  }
  
  a:visited {
    color: #11300d;
  }
  
  form {
    margin: 8px 0;
  }
  
  label {
    display: block;
    margin-bottom: 6px;
    color: #000;
    font-size: 11px;
  }
  
  input[type="email"],
  input[type="password"],
  input[type="date"],
  input[type="text"] {
    width: 100%;
    padding: 3px 4px;
    border-top: 1px solid #446a65;
    border-left: 1px solid #446a65;
    border-right: 1px solid #fff;
    border-bottom: 1px solid #fff;
    box-shadow: inset -1px -1px 0 #80b1b3, inset 1px 1px 0 #000;
    font-size: 11px;
    font-family: inherit;
    background: #fff;
    margin-top: 2px;
  }
  
  input:focus {
    outline: 1px dotted #000;
    outline-offset: -3px;
  }
  
  button {
    background: #80b1b3;
    color: #000;
    border-top: 2px solid #fff;
    border-left: 2px solid #fff;
    border-right: 2px solid #19392c;
    border-bottom: 2px solid #19392c;
    box-shadow: inset 1px 1px 0 #ecf9f2, inset -1px -1px 0 #517b7a;
    padding: 4px 12px;
    font-size: 11px;
    font-weight: normal;
    cursor: pointer;
    font-family: inherit;
    min-width: 75px;
    margin-top: 6px;
  }
  
  button:hover {
    background: #80b1b3;
  }
  
  button:active {
    border-top: 2px solid #19392c;
    border-left: 2px solid #19392c;
    border-right: 2px solid #fff;
    border-bottom: 2px solid #fff;
    box-shadow: inset -1px -1px 0 #ecf9f2, inset 1px 1px 0 #517b7a;
    padding: 5px 11px 3px 13px;
  }
  
  button.secondary {
    background: #80b1b3;
    padding: 3px 8px;
    font-size: 11px;
    margin-left: 6px;
    min-width: 60px;
  }
  
  button.danger {
    background: #80b1b3;
  }
  
  table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 8px;
    background: #fff;
    border-top: 1px solid #446a65;
    border-left: 1px solid #446a65;
    border-right: 1px solid #fff;
    border-bottom: 1px solid #fff;
    font-size: 11px;
  }
  
  th {
    background: #5c9082;
    color: #000;
    padding: 4px;
    text-align: left;
    font-weight: bold;
    font-size: 11px;
  }
  
  td {
    padding: 4px;
    border-bottom: 1px solid #80b1b3;
  }
  
  tr:last-child td {
    border-bottom: none;
  }
  
  code {
    font-family: "Courier New", monospace;
    font-size: 10px;
    background: #fff;
    padding: 2px 4px;
  }

  p code {
    background: #dfdfdf;
    border: 1px solid #808080;
  }
  
  .token-display {
    background: #fff;
    border-top: 2px solid #446a65;
    border-left: 2px solid #446a65;
    border-right: 2px solid #fff;
    border-bottom: 2px solid #fff;
    padding: 8px;
    margin: 8px 0;
    font-family: "Courier New", monospace;
    font-size: 11px;
    word-break: break-all;
    color: #000;
  }

  .token-display-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .token-value {
    flex: 1;
    min-width: 0;
    word-break: break-all;
  }

  .copy-token-btn {
    flex-shrink: 0;
    min-width: 28px;
    padding: 2px 6px;
    margin: 0;
    line-height: 1.2;
  }

  .copy-token-btn:active {
    padding: 2px 6px;
    margin: 0;
  }

  .copy-status {
    margin-top: 4px;
    color: #19392c;
    font-size: 10px;
  }
  
  .alert {
    background: #d2f1e1;
    border-top: 2px solid #fff;
    border-left: 2px solid #fff;
    border-right: 2px solid #446a65;
    border-bottom: 2px solid #446a65;
    padding: 8px;
    margin: 8px 0;
    font-size: 11px;
  }
  
  .success {
    background: #d2f1e1;
  }
  
  .user-badge {
    display: inline-block;
    background: #80b1b3;
    color: #000;
    padding: 2px 8px;
    border-top: 1px solid #fff;
    border-left: 1px solid #fff;
    border-right: 1px solid #446a65;
    border-bottom: 1px solid #446a65;
    font-size: 11px;
    font-weight: bold;
  }
  
  .links {
    text-align: center;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 2px groove #446a65;
  }
  
  .links a {
    margin: 0 8px;
    font-size: 11px;
  }
  
  .form-inline {
    display: inline-block;
    margin-left: 4px;
  }
  
  .form-inline input {
    display: inline-block;
    width: auto;
    margin: 0 4px;
  }
  
  @media (max-width: 768px) {
    .container {
      padding: 2px;
    }
    
    table {
      font-size: 10px;
    }
    
    th, td {
      padding: 3px;
    }
  }
</style>
`

const htmlHead = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Dogelytics</title>
  <link rel="icon" href="/favicons/favicon.ico?v=20260528" sizes="any">
  <link rel="icon" type="image/png" sizes="32x32" href="/favicons/favicon-32x32.png?v=20260528">
  <link rel="icon" type="image/png" sizes="16x16" href="/favicons/favicon-16x16.png?v=20260528">
  <link rel="apple-touch-icon" href="/favicons/apple-touch-icon.png?v=20260528">
  <link rel="manifest" href="/favicons/site.webmanifest?v=20260528">
` + htmlStyles + `
</head>
`

const htmlHeader = htmlHead + `<body>
<div class="container">
  <div class="window-title">
    <span>Dogelytics</span>
    <span></span>
  </div>
  <div class="window-content">
`

const htmlFooter = `
  </div>
</div>
</body>
</html>
`

// Shared assets for the public meme-number easter egg.
const memeNumberAssets = `
<style>
  .meme-highlight {
    background: #a9dec2;
    color: #000;
    cursor: pointer;
    border-radius: 2px;
    padding: 0 1px;
  }

  .meme-highlight:focus {
    outline: 1px dotted #000;
    outline-offset: 1px;
  }
</style>
<script src="https://cdn.jsdelivr.net/npm/canvas-confetti@1.9.4/dist/confetti.browser.min.js"></script>
<script>
  (function() {
    var matchPattern = /4(?:20|[.,-]20|2[.,-]0)|6[.,-]?9/g;
    var highlightClass = "meme-highlight";
    var scanQueued = false;
    var skippedTags = {
      A: true,
      BUTTON: true,
      INPUT: true,
      SCRIPT: true,
      STYLE: true,
      TEXTAREA: true
    };

    function getMemeType(value) {
      if (value.charAt(0) === "4") {
        return "420";
      }
      return "69";
    }

    function isIgnoredTextNode(node) {
      var current = node.parentNode;
      while (current && current !== document.body) {
        if (current.nodeType === Node.ELEMENT_NODE) {
          if (current.classList && current.classList.contains(highlightClass)) {
            return true;
          }
          if (skippedTags[current.tagName] || current.isContentEditable) {
            return true;
          }
        }
        current = current.parentNode;
      }
      return false;
    }

    function highlightTextNode(node) {
      var text = node.nodeValue;
      var fragment = document.createDocumentFragment();
      var lastIndex = 0;
      var match;

      matchPattern.lastIndex = 0;
      while ((match = matchPattern.exec(text)) !== null) {
        if (match.index > lastIndex) {
          fragment.appendChild(document.createTextNode(text.slice(lastIndex, match.index)));
        }

        var highlight = document.createElement("span");
        highlight.className = highlightClass;
        highlight.setAttribute("data-meme", getMemeType(match[0]));
        highlight.setAttribute("role", "button");
        highlight.setAttribute("tabindex", "0");
        highlight.textContent = match[0];
        fragment.appendChild(highlight);

        lastIndex = match.index + match[0].length;
      }

      if (lastIndex < text.length) {
        fragment.appendChild(document.createTextNode(text.slice(lastIndex)));
      }

      if (node.parentNode) {
        node.parentNode.replaceChild(fragment, node);
      }
    }

    function scan(root) {
      var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
        acceptNode: function(node) {
          if (!node.nodeValue || !node.nodeValue.trim() || isIgnoredTextNode(node)) {
            return NodeFilter.FILTER_REJECT;
          }

          matchPattern.lastIndex = 0;
          if (!matchPattern.test(node.nodeValue)) {
            return NodeFilter.FILTER_REJECT;
          }

          return NodeFilter.FILTER_ACCEPT;
        }
      });
      var nodes = [];
      var current;

      while ((current = walker.nextNode()) !== null) {
        nodes.push(current);
      }

      for (var i = 0; i < nodes.length; i++) {
        highlightTextNode(nodes[i]);
      }
    }

    function queueScan() {
      if (scanQueued) {
        return;
      }

      scanQueued = true;
      (window.requestAnimationFrame || window.setTimeout)(function() {
        scanQueued = false;
        if (document.body) {
          scan(document.body);
        }
      }, 16);
    }

    function normaliseCoordinate(value, maximum) {
      if (maximum <= 0) {
        return 0.5;
      }

      var coordinate = value / maximum;
      if (coordinate < 0) {
        return 0;
      }
      if (coordinate > 1) {
        return 1;
      }
      return coordinate;
    }

    function fireConfetti(element, memeType) {
      if (typeof window.confetti !== "function") {
        return;
      }

      var rect = element.getBoundingClientRect();
      var origin = {
        x: normaliseCoordinate(rect.left + (rect.width / 2), window.innerWidth || document.documentElement.clientWidth),
        y: normaliseCoordinate(rect.top + (rect.height / 2), window.innerHeight || document.documentElement.clientHeight)
      };
      var colours = memeType === "420"
        ? ["#1b7a1b", "#2ecc40", "#7cfc00", "#3a5f0b", "#9acd32", "#006400"]
        : ["#ff0000", "#9dd2c3", "#1e90ff", "#ff69b4", "#00cc00"];
      var particleCount = memeType === "420" ? 60 : 45;
      var options = {
        colors: colours,
        disableForReducedMotion: true,
        origin: origin,
        scalar: 0.9,
        spread: 90,
        startVelocity: 32,
        zIndex: 2000
      };

      window.confetti(Object.assign({}, options, {
        angle: 60,
        drift: -0.2,
        particleCount: particleCount
      }));
      window.confetti(Object.assign({}, options, {
        angle: 120,
        drift: 0.2,
        particleCount: particleCount
      }));
    }

    function findHighlightTarget(target) {
      if (!target || target.nodeType !== Node.ELEMENT_NODE) {
        return null;
      }
      return target.closest("." + highlightClass);
    }

    function shouldRescan(mutations) {
      for (var i = 0; i < mutations.length; i++) {
        var mutation = mutations[i];

        if (mutation.type === "characterData") {
          if (!isIgnoredTextNode(mutation.target)) {
            return true;
          }
          continue;
        }

        if (mutation.type !== "childList") {
          continue;
        }

        for (var j = 0; j < mutation.addedNodes.length; j++) {
          var addedNode = mutation.addedNodes[j];

          if (addedNode.nodeType === Node.TEXT_NODE) {
            if (!isIgnoredTextNode(addedNode)) {
              return true;
            }
            continue;
          }

          if (addedNode.nodeType === Node.ELEMENT_NODE) {
            if (addedNode.classList && addedNode.classList.contains(highlightClass)) {
              continue;
            }
            if (addedNode.closest && addedNode.closest("." + highlightClass)) {
              continue;
            }
            return true;
          }
        }
      }

      return false;
    }

    function init() {
      if (!document.body) {
        return;
      }

      scan(document.body);

      var observer = new MutationObserver(function(mutations) {
        if (shouldRescan(mutations)) {
          queueScan();
        }
      });
      observer.observe(document.body, {
        childList: true,
        subtree: true,
        characterData: true
      });

      document.addEventListener("click", function(event) {
        var target = findHighlightTarget(event.target);
        if (target) {
          fireConfetti(target, target.getAttribute("data-meme"));
        }
      });

      document.addEventListener("keydown", function(event) {
        if (event.key !== "Enter" && event.key !== " ") {
          return;
        }

        var target = findHighlightTarget(event.target);
        if (!target) {
          return;
        }

        event.preventDefault();
        fireConfetti(target, target.getAttribute("data-meme"));
      });
    }

    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", init);
      return;
    }

    init();
  })();
</script>
`
