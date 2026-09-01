# Filename: vulnerable_app.py

import hashlib
import sqlite3
import subprocess
import json
import pickle  # Used here to demonstrate insecure deserialization
from flask import Flask, request, jsonify

app = Flask(__name__)

# Weak hashing algorithm (MD5)
@app.route('/hash', methods=['POST'])
def hash_password():
    password = request.form['password']
    hashed = hashlib.md5(password.encode()).hexdigest()  # MD5 is a weak hashing algorithm
    return jsonify({"hashed_password": hashed})

# SQL Injection
@app.route('/user', methods=['GET'])
def get_user():
    user_id = request.args.get('id')
    if user_id is None:
        return jsonify({"error": "Missing id parameter"}), 400
    try:
        user_id_int = int(user_id, 10)
    except (TypeError, ValueError):
        return jsonify({"error": "Invalid id parameter"}), 400
    if user_id_int < 0:
        return jsonify({"error": "Invalid id parameter"}), 400
    conn = sqlite3.connect('test.db')
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM users WHERE id = ?", (user_id_int,))
    user = cursor.fetchone()
    conn.close()
    if user:
        return jsonify({"user": user})
    else:
        return jsonify({"error": "User not found"}), 404

# Command Injection
@app.route('/exec', methods=['GET'])
def execute_command():
    filename = request.args.get('filename')
    command = f"cat {filename}"  # Command injection vulnerability
    result = subprocess.check_output(command, shell=True)  # shell=True allows command injection
    return result

# Insecure Deserialization
@app.route('/deserialize', methods=['POST'])
def deserialize_data():
    data = request.data
    obj = pickle.loads(data)  # Insecure deserialization vulnerability
    return jsonify({"deserialized_object": str(obj)})

# Hardcoded sensitive data
API_KEY = "12345-ABCDE"  # Hardcoded sensitive data

# Directory Traversal vulnerability
@app.route('/file', methods=['GET'])
def read_file():
    filename = request.args.get('file')
    with open(f"/var/www/files/{filename}", "r") as f:  # Directory traversal vulnerability
        content = f.read()
    return content

# Cross-Site Scripting (XSS)
@app.route('/greet', methods=['GET'])
def greet_user():
    name = request.args.get('name')
    return f"<h1>Hello, {name}</h1>"  # XSS vulnerability

if __name__ == '__main__':
    app.run(port=5000)
