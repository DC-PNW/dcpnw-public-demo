// Vulnerable TypeScript code

import express from "express";
import mysql from "mysql";
import jwt from "jsonwebtoken";
import crypto from "crypto";

const app = express();

// 1. Hardcoded sensitive data
const dbPassword = "hardcodedpassword123";
const jwtSecret = "supersecretkey";

// 2. Insecure MySQL connection (No parameterized queries, SQL Injection possible)
const connection = mysql.createConnection({
    host: "localhost",
    user: "root",
    password: dbPassword,
    database: "testdb",
});

app.use(express.json());

// 3. Vulnerable to SQL Injection
app.post("/login", (req, res) => {
    const { username, password } = req.body;

    const query = `SELECT * FROM users WHERE username = '${username}' AND password = '${password}'`; // SQL Injection
    connection.query(query, (err, results) => {
        if (err) {
            res.status(500).send("Internal Server Error");
            return;
        }

        if (results.length > 0) {
            const token = jwt.sign({ username }, jwtSecret); // 4. Insecure JWT secret
            res.json({ token });
        } else {
            res.status(401).send("Unauthorized");
        }
    });
});

// 5. Weak cryptographic algorithms
app.post("/hash", (req, res) => {
    const { data } = req.body;
    const hash = crypto.createHash("md5").update(data).digest("hex"); // MD5 is insecure
    res.send({ hash });
});

// 6. Command Injection
app.post("/exec", (req, res) => {
    const { command } = req.body;
    const exec = require("child_process").exec;
    exec(command, (error: any, stdout: string) => {
        if (error) {
            res.status(500).send(error.message);
            return;
        }
        res.send(stdout);
    });
});

// 7. Exposing stack traces in production
app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
    res.status(500).send(err.stack); // Stack trace exposure
});

app.listen(3000, () => {
    console.log("Server running on port 3000");
});
