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
    background: #008080;
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    color: #000;
    font-size: 11px;
  }
  
  .container {
    background: #c0c0c0;
    border-top: 2px solid #fff;
    border-left: 2px solid #fff;
    border-right: 2px solid #000;
    border-bottom: 2px solid #000;
    box-shadow: inset 1px 1px 0 #dfdfdf, inset -1px -1px 0 #808080;
    padding: 2px;
    max-width: 900px;
    width: 100%;
  }
  
  .window-title {
    background: linear-gradient(to right, #000080, #1084d0);
    color: #fff;
    padding: 3px 5px;
    font-weight: bold;
    font-size: 11px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2px;
  }
  
  .window-content {
    background: #c0c0c0;
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
    color: #fff;
    font-size: 13px;
    font-weight: bold;
    margin-bottom: 8px;
    padding: 3px 5px;
    background: #000080;
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
    color: #00f;
    text-decoration: underline;
  }
  
  a:hover {
    color: #f0f;
  }
  
  a:visited {
    color: #800080;
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
    border-top: 1px solid #808080;
    border-left: 1px solid #808080;
    border-right: 1px solid #fff;
    border-bottom: 1px solid #fff;
    box-shadow: inset -1px -1px 0 #c0c0c0, inset 1px 1px 0 #000;
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
    background: #c0c0c0;
    color: #000;
    border-top: 2px solid #fff;
    border-left: 2px solid #fff;
    border-right: 2px solid #000;
    border-bottom: 2px solid #000;
    box-shadow: inset 1px 1px 0 #dfdfdf, inset -1px -1px 0 #808080;
    padding: 4px 12px;
    font-size: 11px;
    font-weight: normal;
    cursor: pointer;
    font-family: inherit;
    min-width: 75px;
    margin-top: 6px;
  }
  
  button:hover {
    background: #c0c0c0;
  }
  
  button:active {
    border-top: 2px solid #000;
    border-left: 2px solid #000;
    border-right: 2px solid #fff;
    border-bottom: 2px solid #fff;
    box-shadow: inset -1px -1px 0 #dfdfdf, inset 1px 1px 0 #808080;
    padding: 5px 11px 3px 13px;
  }
  
  button.secondary {
    background: #c0c0c0;
    padding: 3px 8px;
    font-size: 11px;
    margin-left: 6px;
    min-width: 60px;
  }
  
  button.danger {
    background: #c0c0c0;
  }
  
  table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 8px;
    background: #fff;
    border-top: 1px solid #808080;
    border-left: 1px solid #808080;
    border-right: 1px solid #fff;
    border-bottom: 1px solid #fff;
    font-size: 11px;
  }
  
  th {
    background: #000080;
    color: #fff;
    padding: 4px;
    text-align: left;
    font-weight: bold;
    font-size: 11px;
  }
  
  td {
    padding: 4px;
    border-bottom: 1px solid #c0c0c0;
  }
  
  tr:last-child td {
    border-bottom: none;
  }
  
  tr:hover {
    background: #00007f;
    color: #fff;
  }
  
  tr:hover code {
    background: #00007f;
    color: #fff;
  }
  
  code {
    font-family: "Courier New", monospace;
    font-size: 10px;
    background: #fff;
    padding: 2px 4px;
  }
  
  .token-display {
    background: #fff;
    border-top: 2px solid #808080;
    border-left: 2px solid #808080;
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
    color: #000080;
    font-size: 10px;
  }
  
  .alert {
    background: #ffffe1;
    border-top: 2px solid #fff;
    border-left: 2px solid #fff;
    border-right: 2px solid #808080;
    border-bottom: 2px solid #808080;
    padding: 8px;
    margin: 8px 0;
    font-size: 11px;
  }
  
  .success {
    background: #c6ffc6;
  }
  
  .user-badge {
    display: inline-block;
    background: #c0c0c0;
    color: #000;
    padding: 2px 8px;
    border-top: 1px solid #fff;
    border-left: 1px solid #fff;
    border-right: 1px solid #808080;
    border-bottom: 1px solid #808080;
    font-size: 11px;
    font-weight: bold;
  }
  
  .links {
    text-align: center;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 2px groove #808080;
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

const htmlHeader = `<!doctype html>
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
<body>
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
