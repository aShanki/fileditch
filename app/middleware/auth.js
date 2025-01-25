// Authentication middleware
const requireAuth = (req, res, next) => {
    if (req.session.isAuthenticated) {
        return next();
    }
    res.status(401).json({ error: 'Unauthorized' });
};

// Password verification
const verifyPassword = (req, res, next) => {
    const { password } = req.body;
    
    if (password === process.env.SITE_PASSWORD) {
        req.session.isAuthenticated = true;
        res.json({ success: true });
    } else {
        // Log failed attempt
        console.log(`Failed login attempt from IP: ${req.ip}`);
        res.status(401).json({ error: 'Invalid password' });
    }
};

module.exports = {
    requireAuth,
    verifyPassword
};