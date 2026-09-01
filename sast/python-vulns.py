# Filename: python-vulns.py
# SAST demo file with intentionally vulnerable code patterns for scanner testing.

import os
import re
import ssl
import yaml
import xml.etree.ElementTree as ET
from flask import Flask, request, redirect

app = Flask(__name__)

# Hardcoded credentials
DB_PASSWORD = "SuperSecret123!"
AWS_SECRET_KEY = "AKIAABCDEFGHIJKLMNOP"

# Server-Side Request Forgery (SSRF)
@app.route('/fetch', methods=['GET'])
def fetch_url():
    import requests
    target = request.args.get('url')
    resp = requests.get(target)  # unvalidated user-controlled URL
    return resp.text

# XML External Entity (XXE) Injection
@app.route('/parse-xml', methods=['POST'])
def parse_xml():
    xml_data = request.data
    tree = ET.fromstring(xml_data)  # default parser resolves external entities
    return str(tree)

# Insecure YAML deserialization
@app.route('/load-config', methods=['POST'])
def load_config():
    raw = request.form['config']
    config = yaml.load(raw, Loader=yaml.UnsafeLoader)  # unsafe loader allows arbitrary code execution
    return str(config)

# Open Redirect
@app.route('/redirect', methods=['GET'])
def open_redirect():
    next_url = request.args.get('next')
    return redirect(next_url)  # unvalidated redirect target

# Path Traversal
@app.route('/download', methods=['GET'])
def download_file():
    filename = request.args.get('file')
    path = os.path.join('/var/app/uploads', filename)  # no sanitization of '../' sequences
    with open(path, 'rb') as f:
        return f.read()

# ReDoS-prone regular expression
def validate_email(email):
    pattern = r'^([a-zA-Z0-9_\.\-])+@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$'
    return re.match(pattern, email) is not None

# Insecure TLS configuration
def create_insecure_context():
    ctx = ssl.SSLContext(ssl.PROTOCOL_TLSv1)  # deprecated, insecure protocol
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE  # disables certificate verification
    return ctx

# Use of eval on user input
@app.route('/calc', methods=['GET'])
def calculate():
    expression = request.args.get('expr')
    result = eval(expression)  # arbitrary code execution
    return str(result)

# Debug mode enabled in production
if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0")
