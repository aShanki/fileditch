const express = require('express');
const router = express.Router();
const path = require('path');
const { verifyPassword } = require('../middleware/auth');

// Serve login page
router.get('/', (req, res) => {
    if (req.session.isAuthenticated) {
        res.sendFile(path.resolve(__dirname, '../public/upload.html'));
    } else {
        res.sendFile(path.resolve(__dirname, '../public/login.html'));
    }
});

// Handle login
router.post('/login', verifyPassword);

// Handle logout
router.post('/logout', (req, res) => {
    req.session.destroy();
    res.json({ success: true });
});

module.exports = router;