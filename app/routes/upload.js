const express = require('express');
const router = express.Router();
const multer = require('multer');
const path = require('path');
const crypto = require('crypto');
const db = require('../db/schema');
const { requireAuth } = require('../middleware/auth');

// Configure multer for file upload
const storage = multer.diskStorage({
    destination: (req, file, cb) => {
        cb(null, process.env.UPLOAD_DIR);
    },
    filename: (req, file, cb) => {
        generateUniqueFilename(file.originalname)
            .then(filename => cb(null, filename))
            .catch(err => cb(err));
    }
});

// File filter based on allowed types
const fileFilter = (req, file, cb) => {
    const allowedTypes = process.env.ALLOWED_TYPES.split(',');
    if (allowedTypes.includes('all')) {
        return cb(null, true);
    }
    
    const ext = path.extname(file.originalname).toLowerCase().substring(1);
    if (allowedTypes.includes(ext)) {
        cb(null, true);
    } else {
        cb(new Error('File type not allowed'));
    }
};

// Convert MAX_FILE_SIZE from MB to bytes
const maxFileSize = parseInt(process.env.MAX_FILE_SIZE) * 1024 * 1024;

const upload = multer({
    storage,
    fileFilter,
    limits: {
        fileSize: maxFileSize // Use the converted size
    }
});

// Generate unique filename
async function generateUniqueFilename(originalName) {
    const randomString = crypto.randomBytes(parseInt(process.env.RANDOM_STRING_LENGTH) / 2)
        .toString('hex');
    
    const sanitizedName = path.parse(originalName).name
        .replace(/[^a-z0-9]/gi, '-')
        .toLowerCase();
    
    const ext = path.extname(originalName).toLowerCase();
    const filename = `${sanitizedName}_${randomString}${ext}`;
    
    return new Promise((resolve, reject) => {
        // Check for collision
        db.get('SELECT id FROM files WHERE url_path = ?', [filename], (err, row) => {
            if (err) reject(err);
            if (row) {
                // If collision, try again
                generateUniqueFilename(originalName).then(resolve).catch(reject);
            } else {
                resolve(filename);
            }
        });
    });
}

// Error handler middleware
const handleUploadError = (err, req, res, next) => {
    if (err instanceof multer.MulterError) {
        if (err.code === 'LIMIT_FILE_SIZE') {
            return res.status(413).json({
                error: `File too large. Maximum size is ${process.env.MAX_FILE_SIZE}MB`
            });
        }
        return res.status(400).json({ error: err.message });
    }
    next(err);
};

// Handle file upload
router.post('/', requireAuth, upload.single('file'), handleUploadError, (req, res) => {
    if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
    }

    const expiresAt = new Date();
    expiresAt.setHours(expiresAt.getHours() + parseInt(req.body.expireHours || 24));

    const fileData = {
        original_name: req.file.originalname,
        file_path: req.file.path,
        url_path: req.file.filename,
        mime_type: req.file.mimetype,
        size: req.file.size,
        expires_at: expiresAt.toISOString()
    };

    db.run(`
        INSERT INTO files (original_name, file_path, url_path, mime_type, size, expires_at)
        VALUES (?, ?, ?, ?, ?, ?)
    `, [
        fileData.original_name,
        fileData.file_path,
        fileData.url_path,
        fileData.mime_type,
        fileData.size,
        fileData.expires_at
    ], function(err) {
        if (err) {
            console.error('Database error:', err);
            return res.status(500).json({ error: 'Failed to save file metadata' });
        }

        res.json({
            url: `${process.env.DOMAIN}/file/${fileData.url_path}`,
            expiresAt: fileData.expires_at
        });
    });
});

module.exports = router;