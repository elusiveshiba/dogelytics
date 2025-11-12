package web

// HTML template constants with embedded styles
const htmlStyles = `
<style>
  * {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }
  
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    color: #333;
  }
  
  .container {
    background: white;
    border-radius: 12px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
    padding: 24px;
    max-width: 800px;
    width: 100%;
    animation: slideIn 0.3s ease-out;
  }
  
  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(-20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  
  h1 {
    color: #667eea;
    font-size: 2rem;
    margin-bottom: 0.5rem;
    text-align: center;
  }
  
  h2 {
    color: #667eea;
    font-size: 1.5rem;
    margin-bottom: 1rem;
    border-bottom: 2px solid #667eea;
    padding-bottom: 0.4rem;
  }
  
  h3 {
    color: #764ba2;
    font-size: 1.1rem;
    margin-top: 1.5rem;
    margin-bottom: 0.75rem;
  }
  
  p {
    margin-bottom: 0.75rem;
    line-height: 1.5;
    color: #555;
  }
  
  a {
    color: #667eea;
    text-decoration: none;
    font-weight: 500;
    transition: color 0.2s;
  }
  
  a:hover {
    color: #764ba2;
    text-decoration: underline;
  }
  
  form {
    margin: 1rem 0;
  }
  
  label {
    display: block;
    margin-bottom: 0.75rem;
    font-weight: 500;
    color: #444;
    font-size: 0.95rem;
  }
  
  input[type="email"],
  input[type="password"],
  input[type="date"],
  input[type="text"] {
    width: 100%;
    padding: 8px 12px;
    border: 2px solid #e0e0e0;
    border-radius: 6px;
    font-size: 0.95rem;
    margin-top: 0.25rem;
    transition: border-color 0.2s, box-shadow 0.2s;
  }
  
  input:focus {
    outline: none;
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }
  
  button {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 6px;
    font-size: 0.95rem;
    font-weight: 600;
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
    margin-top: 0.5rem;
  }
  
  button:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  }
  
  button:active {
    transform: translateY(0);
  }
  
  button.secondary {
    background: #6c757d;
    padding: 6px 12px;
    font-size: 0.85rem;
    margin-left: 6px;
  }
  
  button.danger {
    background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
  }
  
  table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 0.75rem;
    background: white;
    border-radius: 6px;
    overflow: hidden;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
    font-size: 0.9rem;
  }
  
  th {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 8px;
    text-align: left;
    font-weight: 600;
    font-size: 0.9rem;
  }
  
  td {
    padding: 8px;
    border-bottom: 1px solid #f0f0f0;
  }
  
  tr:last-child td {
    border-bottom: none;
  }
  
  tr:hover {
    background: #f8f9fa;
  }
  
  .token-display {
    background: #f8f9fa;
    border: 2px dashed #667eea;
    border-radius: 6px;
    padding: 12px;
    margin: 0.75rem 0;
    font-family: "Courier New", monospace;
    font-size: 0.95rem;
    word-break: break-all;
    color: #333;
    box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.05);
  }
  
  .alert {
    background: #fff3cd;
    border-left: 3px solid #ffc107;
    padding: 8px 12px;
    margin: 0.75rem 0;
    border-radius: 4px;
    color: #856404;
    font-size: 0.9rem;
  }
  
  .success {
    background: #d4edda;
    border-left-color: #28a745;
    color: #155724;
  }
  
  .user-badge {
    display: inline-block;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 4px 10px;
    border-radius: 16px;
    font-size: 0.85rem;
    font-weight: 500;
    margin-bottom: 0.5rem;
  }
  
  .links {
    text-align: center;
    margin-top: 1.25rem;
    padding-top: 1rem;
    border-top: 1px solid #e0e0e0;
  }
  
  .links a {
    margin: 0 0.75rem;
    font-size: 0.95rem;
  }
  
  .form-inline {
    display: inline-block;
    margin-left: 8px;
  }
  
  .form-inline input {
    display: inline-block;
    width: auto;
    margin: 0 4px;
  }
  
  @media (max-width: 768px) {
    .container {
      padding: 16px;
    }
    
    h1 {
      font-size: 1.5rem;
    }
    
    h2 {
      font-size: 1.25rem;
    }
    
    table {
      font-size: 0.8rem;
    }
    
    th, td {
      padding: 6px;
    }
  }
</style>
`

const htmlHeader = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Dogelytics - Dogecoin Balance Analytics</title>
` + htmlStyles + `
</head>
<body>
<div class="container">
`

const htmlFooter = `
</div>
</body>
</html>
`

