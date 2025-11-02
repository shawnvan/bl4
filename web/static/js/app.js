// BL4 Item Serial Code Codec - Frontend Application

class BL4Codec {
    constructor() {
        this.apiBase = '/api/v1';
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.checkAPIStatus();
    }

    setupEventListeners() {
        // Decode form
        document.getElementById('decodeForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleDecode();
        });

        // Encode form
        document.getElementById('encodeForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleEncode();
        });

        // Batch form
        document.getElementById('batchForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleBatch();
        });

        // Validate form
        document.getElementById('validateForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleValidate();
        });

        // Batch file upload
        document.getElementById('batchFile').addEventListener('change', (e) => {
            this.handleFileUpload(e);
        });
    }

    async checkAPIStatus() {
        try {
            const response = await fetch('/health');
            if (response.ok) {
                this.updateStatus('API connected', 'success');
                document.getElementById('apiStatus').textContent = '🟢 API Connected';
            } else {
                throw new Error('API not responding correctly');
            }
        } catch (error) {
            this.updateStatus('API connection failed', 'error');
            document.getElementById('apiStatus').textContent = '🔴 API Disconnected';
        }
    }

    showTab(tabName) {
        // Hide all tabs
        document.querySelectorAll('.tab-content').forEach(tab => {
            tab.classList.remove('active');
        });

        // Remove active class from all buttons
        document.querySelectorAll('.tab-button').forEach(button => {
            button.classList.remove('active');
        });

        // Show selected tab
        document.getElementById(tabName).classList.add('active');
        event.target.classList.add('active');
    }

    async handleDecode() {
        const code = document.getElementById('decodeCode').value.trim();
        const includeBitstream = document.getElementById('includeBitstream').checked;
        const includeTokens = document.getElementById('includeTokens').checked;

        if (!code) {
            this.showError('decodeResult', 'Please enter a serial code');
            return;
        }

        this.showLoading('decodeResult');
        this.updateStatus('Decoding serial code...');

        try {
            const response = await fetch(`${this.apiBase}/items/decode`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    serial_code: code,
                    options: {
                        include_bitstream: includeBitstream,
                        include_tokens: includeTokens,
                        format: 'json'
                    }
                })
            });

            const result = await response.json();

            if (response.ok) {
                this.showDecodeResult(result);
                this.updateStatus('Decode completed successfully', 'success');
            } else {
                this.showError('decodeResult', result.message || 'Decode failed');
                this.updateStatus('Decode failed', 'error');
            }
        } catch (error) {
            this.showError('decodeResult', `Network error: ${error.message}`);
            this.updateStatus('Network error', 'error');
        }
    }

    async handleEncode() {
        const itemData = {
            level: parseInt(document.getElementById('itemLevel').value),
            type: document.getElementById('itemType').value,
            manufacturer: document.getElementById('manufacturer').value,
        };

        const itemName = document.getElementById('itemName').value.trim();
        if (itemName) {
            itemData.name = itemName;
        }

        this.showLoading('encodeResult');
        this.updateStatus('Encoding item data...');

        try {
            const response = await fetch(`${this.apiBase}/items/encode`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    item_data: itemData,
                    options: {
                        format: 'json'
                    }
                })
            });

            const result = await response.json();

            if (response.ok) {
                this.showEncodeResult(result);
                this.updateStatus('Encode completed successfully', 'success');
            } else {
                this.showError('encodeResult', result.message || 'Encode failed');
                this.updateStatus('Encode failed', 'error');
            }
        } catch (error) {
            this.showError('encodeResult', `Network error: ${error.message}`);
            this.updateStatus('Network error', 'error');
        }
    }

    async handleBatch() {
        const codesText = document.getElementById('batchCodes').value.trim();

        if (!codesText) {
            this.showError('batchResults', 'Please enter serial codes or upload a file');
            return;
        }

        const codes = codesText.split('\n')
            .map(code => code.trim())
            .filter(code => code.length > 0);

        if (codes.length === 0) {
            this.showError('batchResults', 'No valid serial codes found');
            return;
        }

        this.showBatchProgress();
        this.updateStatus(`Processing batch of ${codes.length} codes...`);

        try {
            const response = await fetch(`${this.apiBase}/items/batch/decode`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    serial_codes: codes,
                    options: {
                        format: 'json'
                    }
                })
            });

            const result = await response.json();

            if (response.ok) {
                this.showBatchResults(result);
                this.updateStatus('Batch processing completed', 'success');
            } else {
                this.showError('batchResults', result.message || 'Batch processing failed');
                this.updateStatus('Batch processing failed', 'error');
            }
        } catch (error) {
            this.showError('batchResults', `Network error: ${error.message}`);
            this.updateStatus('Network error', 'error');
        }

        this.hideBatchProgress();
    }

    async handleValidate() {
        const code = document.getElementById('validateCode').value.trim();

        if (!code) {
            this.showError('validateResult', 'Please enter a serial code');
            return;
        }

        this.showLoading('validateResult');
        this.updateStatus('Validating serial code...');

        try {
            const response = await fetch(`${this.apiBase}/items/validate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    serial_code: code
                })
            });

            const result = await response.json();

            if (response.ok) {
                this.showValidateResult(result);
                this.updateStatus('Validation completed', 'success');
            } else {
                this.showError('validateResult', result.message || 'Validation failed');
                this.updateStatus('Validation failed', 'error');
            }
        } catch (error) {
            this.showError('validateResult', `Network error: ${error.message}`);
            this.updateStatus('Network error', 'error');
        }
    }

    handleFileUpload(event) {
        const file = event.target.files[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (e) => {
            const content = e.target.result;
            document.getElementById('batchCodes').value = content;
            this.updateStatus(`Loaded ${file.name}`, 'info');
        };
        reader.readAsText(file);
    }

    showDecodeResult(result) {
        const resultBox = document.getElementById('decodeResult');
        resultBox.className = 'result-box success';
        resultBox.style.display = 'block';

        let html = '<h3>✅ Decode Successful</h3>';

        // Debug: Log the result structure
        console.log('Decode Result:', result);

        // Fix: Access serial_code from result.data.serial_code
        const serialCode = result.data?.serial_code || 'Unknown';
        html += `<p><strong>Serial Code:</strong> <code style="background: #f8f9fa; padding: 4px 8px; border-radius: 4px; font-family: monospace;">${serialCode}</code></p>`;

        // Add reference format if available
        if (result.data?.reference_format) {
            html += `<p><strong>Reference Format:</strong> <code style="background: #e9ecef; padding: 4px 8px; border-radius: 4px; font-family: monospace; display: block; margin-top: 4px; word-break: break-all;">${result.data.reference_format}</code></p>`;
        }

        if (result.data?.item_data) {
            const itemData = result.data.item_data;
            html += '<h4>📦 Item Data:</h4>';
            html += `<p><strong>Level:</strong> ${itemData.level}</p>`;
            html += `<p><strong>Type:</strong> ${itemData.type}</p>`;
            html += `<p><strong>Manufacturer:</strong> ${itemData.manufacturer}</p>`;

            if (itemData.name) {
                html += `<p><strong>Name:</strong> ${itemData.name}</p>`;
            }

            if (itemData.rarity) {
                html += `<p><strong>Rarity:</strong> ${itemData.rarity}</p>`;
            }

            if (itemData.parts && itemData.parts.length > 0) {
                html += '<h5>🔧 Parts:</h5>';
                html += '<ul style="margin: 10px 0; padding-left: 20px;">';
                itemData.parts.forEach((part, index) => {
                    html += `<li><strong>Part ${index + 1}:</strong> ${part.type} (Value: ${part.value})</li>`;
                });
                html += '</ul>';
            }
        }

        // Add bitstream info if available
        if (result.data?.bitstream) {
            const bitstream = result.data.bitstream;
            html += '<h4>🔢 Bitstream Info:</h4>';
            html += `<p><strong>Size:</strong> ${bitstream.size} bits (${bitstream.byte_size} bytes)</p>`;
            html += `<p><strong>Format:</strong> ${bitstream.format}</p>`;
            if (bitstream.statistics) {
                html += `<p><strong>Token Count:</strong> ${bitstream.statistics.token_count}</p>`;
            }
        }

        if (result.metadata?.duration) {
            html += `<p><em>⏱️ Processing time: ${result.metadata.duration}μs</em></p>`;
        }

        // Show raw JSON in collapsible section
        html += '<details style="margin-top: 20px;"><summary>📋 Raw JSON Response</summary>';
        html += `<pre style="background: #f8f9fa; padding: 10px; border-radius: 4px; overflow-x: auto; font-size: 12px;">${JSON.stringify(result, null, 2)}</pre></details>`;

        resultBox.innerHTML = html;
    }

    showEncodeResult(result) {
        const resultBox = document.getElementById('encodeResult');
        resultBox.className = 'result-box success';
        resultBox.style.display = 'block';

        let html = '<h3>✅ Encode Successful</h3>';

        // Fix: Access serial_code from result.data.serial_code
        const serialCode = result.data?.serial_code || 'Unknown';
        html += `<p><strong>Generated Serial Code:</strong></p>`;
        html += `<p style="font-family: monospace; font-size: 1.1em; background: #f0f0f0; padding: 10px; border-radius: 4px;">${serialCode}</p>`;

        if (result.data?.item_data) {
            const itemData = result.data.item_data;
            html += '<h4>Encoded Item Data:</h4>';
            html += `<p><strong>Level:</strong> ${itemData.level}</p>`;
            html += `<p><strong>Type:</strong> ${itemData.type}</p>`;
            html += `<p><strong>Manufacturer:</strong> ${itemData.manufacturer}</p>`;
        }

        if (result.metadata?.duration) {
            html += `<p><em>Processing time: ${result.metadata.duration}μs</em></p>`;
        }

        // Copy button
        html += `<button onclick="navigator.clipboard.writeText('${serialCode}')">Copy Code</button>`;

        resultBox.innerHTML = html;
    }

    showBatchResults(result) {
        const resultBox = document.getElementById('batchResults');
        resultBox.className = 'result-box success';
        resultBox.style.display = 'block';

        // Debug: Log the result structure
        console.log('Batch Results:', result);

        let html = `<h3>✅ Batch Processing Complete</h3>`;

        // Show statistics if available
        if (result.statistics) {
            const stats = result.statistics;
            html += `<p>Processed ${stats.total || 0} codes (${stats.successful || 0} successful, ${stats.failed || 0} failed)</p>`;
            if (stats.success_rate !== undefined) {
                html += `<p>Success Rate: ${(stats.success_rate * 100).toFixed(1)}%</p>`;
            }
        } else {
            html += `<p>Processed ${result.total_codes || 0} codes</p>`;
        }

        if (result.results && result.results.length > 0) {
            html += '<table class="batch-results-table">';
            html += '<thead><tr><th>#</th><th>Serial Code</th><th>Status</th><th>Item Type</th><th>Level</th><th>Manufacturer</th></tr></thead>';
            html += '<tbody>';

            result.results.forEach((item, index) => {
                const statusClass = item.success ? 'status-success' : 'status-error';
                const status = item.success ? 'Success' : 'Error';
                const itemType = item.data?.item_data?.type || 'N/A';
                const level = item.data?.item_data?.level || 'N/A';
                const manufacturer = item.data?.item_data?.manufacturer || 'N/A';

                html += `<tr>
                    <td>${index + 1}</td>
                    <td style="font-family: monospace; font-size: 0.9em;">${item.serial_code || 'N/A'}</td>
                    <td class="${statusClass}">${status}</td>
                    <td>${itemType}</td>
                    <td>${level}</td>
                    <td>${manufacturer}</td>
                </tr>`;
            });

            html += '</tbody></table>';
        }

        // Show processing time from metadata if available
        if (result.metadata?.duration) {
            html += `<p><em>Total processing time: ${result.metadata.duration}ms</em></p>`;
        } else if (result.duration_ms) {
            html += `<p><em>Total processing time: ${result.duration_ms}ms</em></p>`;
        }

        resultBox.innerHTML = html;
    }

    showValidateResult(result) {
        const resultBox = document.getElementById('validateResult');

        // Debug: Log the result structure
        console.log('Validate Result:', result);

        // Check if result is valid based on the API response structure
        const isValid = result.valid === true;
        const serialCode = result.serial_code || 'Unknown';
        const error = result.error || null;
        const errorType = result.error_type || null;
        const issues = result.issues || [];
        const length = result.length || 0;
        const base85Length = result.base85_length || 0;

        if (isValid) {
            resultBox.className = 'result-box success';
            let html = '<h3>✅ Valid Serial Code</h3>';
            html += `<p><strong>Code:</strong> <code style="background: #f8f9fa; padding: 4px 8px; border-radius: 4px; font-family: monospace;">${serialCode}</code></p>`;
            html += `<p><strong>Length:</strong> ${length} characters</p>`;
            html += `<p><strong>Base85 Length:</strong> ${base85Length} characters</p>`;
            html += `<p><strong>Format:</strong> ✅ Valid BL4 serial code</p>`;
            resultBox.innerHTML = html;
            this.updateStatus('Code is valid', 'success');
        } else {
            resultBox.className = 'result-box error';
            let html = '<h3>❌ Invalid Serial Code</h3>';
            html += `<p><strong>Code:</strong> <code style="background: #f8f9fa; padding: 4px 8px; border-radius: 4px; font-family: monospace;">${serialCode}</code></p>`;

            if (error) {
                html += `<p><strong>Error:</strong> ${error}</p>`;
            }

            if (errorType) {
                html += `<p><strong>Error Type:</strong> ${errorType}</p>`;
            }

            if (issues && issues.length > 0) {
                html += '<p><strong>Issues:</strong></p>';
                html += '<ul style="margin: 10px 0; padding-left: 20px;">';
                issues.forEach(issue => {
                    html += `<li>${issue}</li>`;
                });
                html += '</ul>';
            }

            html += `<p><strong>Length:</strong> ${length} characters</p>`;
            html += `<p><strong>Base85 Length:</strong> ${base85Length} characters</p>`;

            resultBox.innerHTML = html;
            this.updateStatus('Code is invalid', 'error');
        }

        resultBox.style.display = 'block';
    }

    showLoading(elementId) {
        const element = document.getElementById(elementId);
        element.className = 'result-box info';
        element.style.display = 'block';
        element.innerHTML = '<div class="loading"></div> Processing...';
    }

    showError(elementId, message) {
        const element = document.getElementById(elementId);
        element.className = 'result-box error';
        element.style.display = 'block';
        element.innerHTML = `<h3>❌ Error</h3><p>${message}</p>`;
    }

    showBatchProgress() {
        document.getElementById('batchProgress').style.display = 'block';
        document.getElementById('batchResults').style.display = 'none';
    }

    hideBatchProgress() {
        document.getElementById('batchProgress').style.display = 'none';
    }

    updateStatus(message, type = 'info') {
        const statusElement = document.getElementById('status');
        statusElement.textContent = message;

        // Add color coding
        statusElement.className = '';
        if (type === 'success') {
            statusElement.style.color = '#28a745';
        } else if (type === 'error') {
            statusElement.style.color = '#dc3545';
        } else {
            statusElement.style.color = '#f8f9fa';
        }

        // Auto-clear after 5 seconds
        setTimeout(() => {
            if (statusElement.textContent === message) {
                statusElement.textContent = 'Ready';
                statusElement.style.color = '#f8f9fa';
            }
        }, 5000);
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.bl4Codec = new BL4Codec();
});

// Global functions for onclick handlers
function showTab(tabName) {
    window.bl4Codec.showTab(tabName);
}