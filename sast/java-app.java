// Filename: VulnerableApp.java

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.ResultSet;
import java.sql.Statement;
import java.io.FileInputStream;
import java.io.ObjectInputStream;
import java.security.MessageDigest;
import javax.crypto.Cipher;
import javax.crypto.spec.SecretKeySpec;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.net.URL;
import java.net.HttpURLConnection;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import org.xml.sax.InputSource;
import java.io.StringReader;

public class VulnerableApp extends HttpServlet {

    // Hardcoded sensitive data (API Key)
    private static final String API_KEY = "12345-ABCDE"; // Sensitive information

    // Weak Cryptography - using DES (weak encryption algorithm)
    public String encrypt(String data, String key) throws Exception {
        Cipher cipher = Cipher.getInstance("DES"); // Weak algorithm
        SecretKeySpec keySpec = new SecretKeySpec(key.getBytes(), "DES");
        cipher.init(Cipher.ENCRYPT_MODE, keySpec);
        byte[] encrypted = cipher.doFinal(data.getBytes());
        return new String(encrypted);
    }

    // SQL Injection vulnerability
    protected void doGet(HttpServletRequest request, HttpServletResponse response) throws IOException {
        String userId = request.getParameter("id");
        try {
            Connection conn = DriverManager.getConnection("jdbc:mysql://localhost:3306/test", "root", "password");
            Statement stmt = conn.createStatement();
            String query = "SELECT * FROM users WHERE id = '" + userId + "'"; // Vulnerable to SQL Injection
            ResultSet rs = stmt.executeQuery(query);
            while (rs.next()) {
                response.getWriter().println("User: " + rs.getString("name"));
            }
            conn.close();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // Command Injection vulnerability
    protected void doPost(HttpServletRequest request, HttpServletResponse response) throws IOException {
        String filename = request.getParameter("filename");
        ProcessBuilder pb = new ProcessBuilder("cat", filename); // Command injection
        try {
            Process p = pb.start();
            BufferedReader reader = new BufferedReader(new InputStreamReader(p.getInputStream()));
            String line;
            while ((line = reader.readLine()) != null) {
                response.getWriter().println(line);
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // Insecure Deserialization vulnerability
    public void deserializeData(String filePath) {
        try {
            FileInputStream fileIn = new FileInputStream(filePath);
            ObjectInputStream in = new ObjectInputStream(fileIn);
            Object obj = in.readObject(); // Insecure deserialization
            in.close();
            fileIn.close();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // Weak Hashing - MD5
    public String hashPassword(String password) throws Exception {
        MessageDigest md = MessageDigest.getInstance("MD5"); // Weak hashing algorithm
        byte[] hash = md.digest(password.getBytes());
        StringBuilder hexString = new StringBuilder();
        for (byte b : hash) {
            hexString.append(String.format("%02x", b));
        }
        return hexString.toString();
    }

    // Directory Traversal vulnerability
    protected void readFile(HttpServletRequest request, HttpServletResponse response) throws IOException {
        String fileName = request.getParameter("file");
        BufferedReader reader = new BufferedReader(new InputStreamReader(new FileInputStream("/var/www/files/" + fileName))); // Directory traversal
        String line;
        while ((line = reader.readLine()) != null) {
            response.getWriter().println(line);
        }
        reader.close();
    }

    // Cross-Site Scripting (XSS) vulnerability
    protected void doGreet(HttpServletRequest request, HttpServletResponse response) throws IOException {
        String name = request.getParameter("name");
        response.getWriter().println("<h1>Hello, " + name + "</h1>"); // XSS vulnerability
    }

    // XML External Entity (XXE) Injection vulnerability - CRITICAL
    protected void parseXml(HttpServletRequest request, HttpServletResponse response) throws IOException {
        String xmlInput = request.getParameter("xml");
        try {
            DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
            // External entity resolution left enabled - allows XXE attacks (file disclosure, SSRF, DoS)
            DocumentBuilder db = dbf.newDocumentBuilder();
            db.parse(new InputSource(new StringReader(xmlInput))); // Vulnerable to XXE
            response.getWriter().println("XML parsed successfully");
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    // Server-Side Request Forgery (SSRF) vulnerability - HIGH
    protected void fetchUrl(HttpServletRequest request, HttpServletResponse response) throws IOException {
        String targetUrl = request.getParameter("url");
        try {
            URL url = new URL(targetUrl); // No validation/allowlist of destination host
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("GET");
            BufferedReader reader = new BufferedReader(new InputStreamReader(conn.getInputStream())); // Vulnerable to SSRF
            String line;
            while ((line = reader.readLine()) != null) {
                response.getWriter().println(line);
            }
            reader.close();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
