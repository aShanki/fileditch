const db = require('../db/schema');
const fs = require('fs').promises;
const path = require('path');

const cleanupExpiredFiles = async () => {
    try {
        // Get all expired files
        db.all(
            'SELECT id, file_path FROM files WHERE expires_at < datetime("now")',
            async (err, files) => {
                if (err) {
                    console.error('Error querying expired files:', err);
                    return;
                }

                for (const file of files) {
                    try {
                        // Delete the physical file
                        await fs.unlink(file.file_path);

                        // Remove from database
                        db.run('DELETE FROM files WHERE id = ?', [file.id]);
                        db.run('DELETE FROM access_logs WHERE file_id = ?', [file.id]);

                        console.log(`Cleaned up expired file: ${file.file_path}`);
                    } catch (error) {
                        console.error(`Error cleaning up file ${file.file_path}:`, error);
                    }
                }
            }
        );
    } catch (error) {
        console.error('Cleanup job error:', error);
    }
};

// Run cleanup every hour
const startCleanupJob = () => {
    // Run immediately on startup
    cleanupExpiredFiles();
    
    // Then schedule for every hour
    setInterval(cleanupExpiredFiles, 60 * 60 * 1000);
};

module.exports = {
    startCleanupJob,
    cleanupExpiredFiles
};