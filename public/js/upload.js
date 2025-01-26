document.getElementById('uploadForm').addEventListener('submit', (e) => {
    e.preventDefault();
    
    const file = document.getElementById('file').files[0];
    const expireHours = parseInt(document.getElementById('expireHours').value);
    const errorDiv = document.getElementById('error');
    const resultDiv = document.getElementById('result');
    const urlDisplay = document.getElementById('urlDisplay');
    const expiryTime = document.getElementById('expiryTime');
    const uploadProgress = document.getElementById('uploadProgress');
    const uploadBar = document.getElementById('uploadBar');
    const progressText = document.getElementById('progressText');
    
    if (!file) {
        errorDiv.textContent = 'Please select a file';
        errorDiv.style.display = 'block';
        return;
    }

    // Hide previous results/errors and show progress
    errorDiv.style.display = 'none';
    resultDiv.style.display = 'none';
    uploadProgress.style.display = 'block';
    progressText.textContent = 'Preparing upload...';

    // Create FormData
    const formData = new FormData();
    formData.append('file', file);
    formData.append('expireHours', expireHours.toString());

    // Create and configure XMLHttpRequest
    const xhr = new XMLHttpRequest();
    xhr.open('POST', '/upload', true);

    // Upload progress handler
    xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) {
            const percentComplete = (e.loaded / e.total) * 100;
            uploadBar.style.width = percentComplete + '%';
            progressText.textContent = `Uploading: ${Math.round(percentComplete)}%`;
        }
    };

    // Response handler
    xhr.onload = function() {
        try {
            const data = JSON.parse(xhr.responseText);
            
            if (xhr.status === 200) {
                // Show success message and URL
                resultDiv.style.display = 'block';
                urlDisplay.innerHTML = `URL: <a href="${data.url}" target="_blank">${data.url}</a>`;
                expiryTime.textContent = new Date(data.expiresAt).toLocaleString();
                progressText.textContent = 'Upload complete!';

                // Clear form
                document.getElementById('uploadForm').reset();
                
                // Reset progress bar after a delay
                setTimeout(() => {
                    uploadProgress.style.display = 'none';
                    uploadBar.style.width = '0%';
                    progressText.textContent = '';
                }, 2000);

                // Setup copy button
                const copyButton = document.getElementById('copyButton');
                copyButton.onclick = () => {
                    navigator.clipboard.writeText(data.url)
                        .then(() => {
                            copyButton.textContent = 'Copied!';
                            setTimeout(() => {
                                copyButton.textContent = 'Copy URL';
                            }, 2000);
                        })
                        .catch(() => {
                            copyButton.textContent = 'Failed to copy';
                            setTimeout(() => {
                                copyButton.textContent = 'Copy URL';
                            }, 2000);
                        });
                };
            } else {
                errorDiv.textContent = data.error || 'Upload failed';
                errorDiv.style.display = 'block';
                uploadProgress.style.display = 'none';
                uploadBar.style.width = '0%';
                progressText.textContent = '';
            }
        } catch (err) {
            errorDiv.textContent = 'An error occurred while processing the response';
            errorDiv.style.display = 'block';
            uploadProgress.style.display = 'none';
            uploadBar.style.width = '0%';
            progressText.textContent = '';
        }
    };

    // Error handler
    xhr.onerror = function() {
        errorDiv.textContent = 'Network error occurred. Please try again.';
        errorDiv.style.display = 'block';
        uploadProgress.style.display = 'none';
        uploadBar.style.width = '0%';
        progressText.textContent = '';
    };

    // Send the request
    xhr.send(formData);
});

// Drag and drop handling
const dropZone = document.querySelector('.file-input-container');
const fileInput = document.getElementById('file');

['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
    dropZone.addEventListener(eventName, preventDefaults, false);
    document.body.addEventListener(eventName, preventDefaults, false);
});

function preventDefaults (e) {
    e.preventDefault();
    e.stopPropagation();
}

['dragenter', 'dragover'].forEach(eventName => {
    dropZone.addEventListener(eventName, highlight, false);
});

['dragleave', 'drop'].forEach(eventName => {
    dropZone.addEventListener(eventName, unhighlight, false);
});

function highlight(e) {
    dropZone.classList.add('highlight');
}

function unhighlight(e) {
    dropZone.classList.remove('highlight');
}

dropZone.addEventListener('drop', handleDrop, false);

function handleDrop(e) {
    const dt = e.dataTransfer;
    const files = dt.files;

    fileInput.files = files;
}

// Add logout functionality
document.getElementById('logoutButton').addEventListener('click', async () => {
    try {
        const response = await fetch('/logout', {
            method: 'POST'
        });

        const data = await response.json();
        
        if (response.ok) {
            window.location.href = '/';
        } else {
            console.error('Logout failed:', data.error);
        }
    } catch (err) {
        console.error('Logout error:', err);
    }
});