document.getElementById('uploadForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const file = document.getElementById('file').files[0];
    const expireHours = parseInt(document.getElementById('expireHours').value);
    const errorDiv = document.getElementById('error');
    const resultDiv = document.getElementById('result');
    const urlDisplay = document.getElementById('urlDisplay');
    const expiryTime = document.getElementById('expiryTime');
    
    if (!file) {
        errorDiv.textContent = 'Please select a file';
        errorDiv.style.display = 'block';
        return;
    }

    // Hide previous results/errors
    errorDiv.style.display = 'none';
    resultDiv.style.display = 'none';

    // Create FormData
    const formData = new FormData();
    formData.append('file', file);
    formData.append('expireHours', expireHours.toString());

    try {
        const response = await fetch('/upload', {
            method: 'POST',
            body: formData
        });

        const data = await response.json();

        if (response.ok) {
            // Show success message and URL
            resultDiv.style.display = 'block';
            urlDisplay.innerHTML = `URL: <a href="${data.url}" target="_blank">${data.url}</a>`;
            expiryTime.textContent = new Date(data.expiresAt).toLocaleString();

            // Clear form
            document.getElementById('uploadForm').reset();

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
        }
    } catch (err) {
        errorDiv.textContent = 'An error occurred. Please try again.';
        errorDiv.style.display = 'block';
        console.error('Upload error:', err);
    }
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