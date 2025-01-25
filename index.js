require('dotenv').config();
const express = require('express');
const path = require('path');
const helmet = require('helmet');
const rateLimit = require('express-rate-limit');
const cookieParser = require('cookie-parser');
const session = require('express-session');
const db = require('./app/db/schema');

const app = express();

// Trust proxy - required when behind Nginx
app.set('trust proxy', 1);  // 1 = trust first proxy

// Convert MAX_FILE_SIZE from MB to bytes
const maxFileSize = parseInt(process.env.MAX_FILE_SIZE) * 1024 * 1024;

// Security middleware with custom CSP
app.use(helmet({
    contentSecurityPolicy: {
        directives: {
            defaultSrc: ["'self'"],
            scriptSrc: ["'self'"],
            styleSrc: ["'self'"],
            imgSrc: ["'self'"],
            connectSrc: ["'self'"],
            fontSrc: ["'self'"],
            objectSrc: ["'none'"],
            mediaSrc: ["'self'"],
            frameSrc: ["'none'"],
            frameAncestors: ["'none'"],
            formAction: ["'self'"]
        }
    },
    crossOriginEmbedderPolicy: false
}));

// Body parser middleware with increased limits
app.use(express.json({ limit: maxFileSize }));
app.use(express.urlencoded({ extended: true, limit: maxFileSize }));

app.use(cookieParser(process.env.COOKIE_SECRET));
app.use(session({
    secret: process.env.COOKIE_SECRET,
    resave: false,
    saveUninitialized: false,
    cookie: { 
        secure: process.env.NODE_ENV === 'production',
        httpOnly: true,
        sameSite: 'strict'
    },
    proxy: true // Also required when behind a proxy
}));

// Rate limiting
const uploadLimiter = rateLimit({
    windowMs: 15, // 15 minutes
    max: 1000000, // limit each IP to 10 requests per windowMs
    standardHeaders: true,
    legacyHeaders: false,
    trustProxy: true
});

// Apply rate limiter to upload endpoint
app.use('/upload', uploadLimiter);

// Ensure uploads and data directories exist
const createRequiredDirs = () => {
    const dirs = ['uploads', 'data'];
    dirs.forEach(dir => {
        const dirPath = path.join(__dirname, dir);
        if (!require('fs').existsSync(dirPath)) {
            require('fs').mkdirSync(dirPath);
        }
    });
};
createRequiredDirs();

// Serve static files
app.use(express.static(path.join(__dirname, 'app/public')));

// Routes
app.use('/', require('./app/routes/index'));
app.use('/upload', require('./app/routes/upload'));
app.use('/file', require('./app/routes/file'));

// Error handler for payload too large
app.use((err, req, res, next) => {
    if (err.type === 'entity.too.large') {
        return res.status(413).json({
            error: `File too large. Maximum size is ${process.env.MAX_FILE_SIZE}MB`
        });
    }
    next(err);
});

// General error handler
app.use((err, req, res, next) => {
    console.error(err.stack);
    res.status(500).json({ error: 'Something went wrong!' });
});

// Start cleanup job
require('./app/middleware/cleanup').startCleanupJob();

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
    console.log(`Maximum file size: ${process.env.MAX_FILE_SIZE}MB`);
});
