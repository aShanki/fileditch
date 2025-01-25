const express = require('express');
const router = express.Router();
const path = require('path');
const db = require('../db/schema');

router.get('/:filename', (req, res) => {
    const filename = req.params.filename;
    
    db.get(
        `SELECT * FROM files 
         WHERE url_path = ? 
         AND expires_at > datetime('now')`,
        [filename],
        (err, file) => {
            if (err) {
                console.error('Database error:', err);
                return res.status(500).send('Server error');
            }

            if (!file) {
                return res.status(404).send('File not found or expired');
            }

            // Log access
            db.run(
                `INSERT INTO access_logs (file_id, ip_address, user_agent)
                 VALUES (?, ?, ?)`,
                [file.id, req.ip, req.get('User-Agent')]
            );

            // Update download count
            db.run(
                'UPDATE files SET downloads = downloads + 1 WHERE id = ?',
                [file.id]
            );

            // Set content disposition and type
            res.setHeader('Content-Type', file.mime_type || 'application/octet-stream');
            res.setHeader('Content-Disposition', `inline; filename="${file.original_name}"`);

            // Handle client disconnection
            req.on('close', () => {
                if (!res.writableEnded) {
                    console.log('Client disconnected during file transfer');
                }
            });

            // Send file with absolute path and error handling
            res.sendFile(path.resolve(file.file_path), {
                dotfiles: 'deny',
                headers: {
                    'Cache-Control': 'public, max-age=3600',
                }
            }, (err) => {
                if (err) {
                    if (err.code === 'ECONNABORTED') {
                        console.log('Connection aborted during file transfer');
                        return;
                    }
                    console.error('File send error:', err);
                    if (!res.headersSent) {
                        res.status(500).send('Error serving file');
                    }
                }
            });
        }
    );
});

// Get file info (for preview or confirmation)
router.get('/:filename/info', (req, res) => {
    const filename = req.params.filename;
    
    db.get(
        `SELECT original_name, mime_type, size, downloads, created_at, expires_at 
         FROM files 
         WHERE url_path = ? 
         AND expires_at > datetime('now')`,
        [filename],
        (err, file) => {
            if (err) {
                return res.status(500).json({ error: 'Server error' });
            }

            if (!file) {
                return res.status(404).json({ error: 'File not found or expired' });
            }

            res.json(file);
        }
    );
});

module.exports = router;