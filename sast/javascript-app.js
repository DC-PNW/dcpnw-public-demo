// Filename: vulnerable_script.js

const crypto = require('crypto');
const mysql = require('mysql');
const express = require('express');
const fs = require('fs');
const { exec } = require('child_process');
const app = express();

app.use(express.json());

// Insecure Hashing (weak hashing algorithm)
app.post('/hash', (req, res) => {
    const password = req.body.password;
    const hash = crypto.createHash('md5').update(password).digest('hex'); // Weak hashing algorithm
    res.send(`Hashed password: ${hash}`);
});

// SQL Injection vulnerability
app.get('/user', (req, res) => {
    const userId = req.query.id;
    const connection = mysql.createConnection({
        host: 'localhost',
        user: 'root',
        password: 'password',
        database: 'test'
    });
    connection.connect();
    const query = `SELECT * FROM users WHERE id = '${userId}'`; // SQL injection vulnerability
    connection.query(query, (err, results) => {
        if (err) throw err;
        res.json(results);
    });
    connection.end();
});

// Command Injection vulnerability
app.get('/exec', (req, res) => {
    const filename = req.query.filename;
    exec(`cat ${filename}`, (error, stdout, stderr) => { // Command injection vulnerability
        if (error) {
            console.error(`exec error: ${error}`);
            res.status(500).send('Server Error');
            return;
        }
        res.send(stdout);
    });
});

// Insecure Deserialization
app.post('/deserialize', (req, res) => {
    const data = req.body.data;
    const obj = JSON.parse(data); // Insecure deserialization
    res.json(obj);
});

// Hardcoded sensitive data
const apiKey = "12345-ABCDE"; // Hardcoded sensitive data

// Sensitive data exposure through directory traversal
app.get('/file', (req, res) => {
    const fileName = req.query.file;
    fs.readFile(`/var/www/files/${fileName}`, 'utf8', (err, data) => { // Directory traversal vulnerability
        if (err) {
            res.status(500).send('Error reading file');
            return;
        }
        res.send(data);
    });
});

// Cross-Site Scripting (XSS) vulnerability
app.get('/greet', (req, res) => {
    const name = req.query.name;
    res.send(`<h1>Hello, ${name}</h1>`); // XSS vulnerability
});

// Server setup
app.listen(3000, () => {
    console.log('Server is running on port 3000');
});
