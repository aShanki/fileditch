function formatDate(dateString) {
    return new Date(dateString).toLocaleString();
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function copyUrl() {
    const urlText = document.getElementById('urlDisplay').textContent;
    navigator.clipboard.writeText(urlText).then(() => {
        alert('URL copied to clipboard!');
    });
}

async function logout() {
    try {
        await fetch('/logout', { method: 'POST' });
        window.location.href = '/';
    } catch (err) {
        console.error('Logout failed:', err);
    }
}

function updateFileInfo(file) {
    const fileInfo = document.querySelector('.file-info');
    if (file) {
        fileInfo.textContent = `Selected: ${file.name} (${formatFileSize(file.size)})`;
        fileInfo.style.display = 'block';
    } else {
        fileInfo.textContent = 'Maximum size: 5GB';
    }
}

function handleDragOver(e) {
    e.preventDefault();
    e.stopPropagation();
    const label = document.querySelector('.file-input-label');
    label.classList.add('drag-over');
}

function handleDragLeave(e) {
    e.preventDefault();
    e.stopPropagation();
    const label = document.querySelector('.file-input-label');
    label.classList.remove('drag-over');
}

function handleDrop(e) {
    e.preventDefault();
    e.stopPropagation();
    
    const label = document.querySelector('.file-input-label');
    label.classList.remove('drag-over');
    
    const fileInput = document.getElementById('file');
    const files = e.dataTransfer.files;
    
    if (files.length > 0) {
        fileInput.files = files;
        updateFileInfo(files[0]);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    // Display max file size in the UI
    const maxSizeInMB = 5120; // from .env
    const fileInput = document.getElementById('file');
    const fileLabel = document.querySelector('.file-input-label');
    
    fileInput.setAttribute('title', `Maximum file size: ${formatFileSize(maxSizeInMB * 1024 * 1024)}`);

    // File input change handler
    fileInput.addEventListener('change', (e) => {
        const file = e.target.files[0];
        updateFileInfo(file);
    });

    // Drag and drop handlers
    fileLabel.addEventListener('dragover', handleDragOver);
    fileLabel.addEventListener('dragleave', handleDragLeave);
    fileLabel.addEventListener('drop', handleDrop);

    // Logout button handler
    document.getElementById('logoutButton').addEventListener('click', logout);
    
    // Copy URL button handler
    document.getElementById('copyButton').addEventListener('click', copyUrl);
    
    // Upload form handler
    document.getElementById('uploadForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const expireHours = document.getElementById('expireHours');
        const uploadButton = document.getElementById('uploadButton');
        const progressBar = document.getElementById('uploadProgress');
        const uploadBarElem = document.getElementById('uploadBar');
        const progressText = document.getElementById('progressText');
        const errorDiv = document.getElementById('error');
        const resultDiv = document.getElementById('result');
        
        if (!fileInput.files[0]) {
            errorDiv.textContent = 'Please select a file';
            errorDiv.style.display = 'block';
            return;
        }

        // Client-side file size validation
        const maxSize = maxSizeInMB * 1024 * 1024; // Convert MB to bytes
        if (fileInput.files[0].size > maxSize) {
            errorDiv.textContent = `File too large. Maximum size is ${formatFileSize(maxSize)}`;
            errorDiv.style.display = 'block';
            return;
        }

        const formData = new FormData();
        formData.append('file', fileInput.files[0]);
        formData.append('expireHours', expireHours.value);

        uploadButton.disabled = true;
        progressBar.style.display = 'block';
        errorDiv.style.display = 'none';
        resultDiv.style.display = 'none';

        try {
            const xhr = new XMLHttpRequest();
            
            xhr.upload.onprogress = (e) => {
                if (e.lengthComputable) {
                    const percentComplete = (e.loaded / e.total) * 100;
                    uploadBarElem.style.width = percentComplete + '%';
                    progressText.textContent = `${formatFileSize(e.loaded)} / ${formatFileSize(e.total)} (${Math.round(percentComplete)}%)`;
                }
            };

            xhr.onload = function() {
                try {
                    const data = JSON.parse(xhr.responseText);
                    if (xhr.status === 200) {
                        document.getElementById('urlDisplay').textContent = data.url;
                        document.getElementById('expiryTime').textContent = formatDate(data.expiresAt);
                        resultDiv.style.display = 'block';
                        e.target.reset();
                        updateFileInfo(null);
                    } else {
                        errorDiv.textContent = data.error || 'Upload failed';
                        errorDiv.style.display = 'block';
                    }
                } catch (parseError) {
                    // Handle non-JSON responses (like HTML error pages)
                    if (xhr.status === 413) {
                        errorDiv.textContent = `File too large. Maximum size allowed by the server is ${formatFileSize(maxSize)}`;
                    } else {
                        errorDiv.textContent = `Upload failed (${xhr.status}). Please try again.`;
                    }
                    errorDiv.style.display = 'block';
                }
            };

            xhr.onerror = function() {
                errorDiv.textContent = 'Connection error. Please check your internet connection and try again.';
                errorDiv.style.display = 'block';
            };

            xhr.open('POST', '/upload', true);
            xhr.send(formData);
        } catch (err) {
            errorDiv.textContent = 'An error occurred. Please try again.';
            errorDiv.style.display = 'block';
        } finally {
            uploadButton.disabled = false;
        }
    });
});