const sqlite3 = require('sqlite3').verbose();
const path = require('path');
require('dotenv').config();

const db = new sqlite3.Database(process.env.DB_PATH, (err) => {
    if (err) {
        console.error('Error connecting to database:', err);
        process.exit(1);
    }
    console.log('Connected to SQLite database');
});

// Create tables if they don't exist
const initDb = () => {
    db.serialize(() => {
        // Files table
        db.run(`CREATE TABLE IF NOT EXISTS files (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            original_name TEXT NOT NULL,
            file_path TEXT NOT NULL,
            url_path TEXT NOT NULL UNIQUE,
            mime_type TEXT,
            size INTEGER NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            expires_at DATETIME NOT NULL,
            downloads INTEGER DEFAULT 0
        )`);

        // Access logs table
        db.run(`CREATE TABLE IF NOT EXISTS access_logs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            file_id INTEGER NOT NULL,
            ip_address TEXT,
            user_agent TEXT,
            accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY(file_id) REFERENCES files(id)
        )`);
    });
};

// Initialize database
initDb();

module.exports = db;