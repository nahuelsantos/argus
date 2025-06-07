// Configuration and state management
window.ArgusApp = {
    config: null,
    currentPage: 'home',
    sidebarCollapsed: false,
    statusCheckInterval: null,
    
    async init() {
        this.setupEventHandlers();
        this.showPage('home');
        this.updateActivitySummary();
        
        // Load configuration first, then start status checking
        await this.loadConfiguration();
        this.startStatusChecking();
    },
    
    async loadConfiguration() {
        try {
            // Use current hostname and port for API calls
            const baseUrl = `${window.location.protocol}//${window.location.host}`;
            const configEndpoints = [
                `${baseUrl}/config`,
            ];
            
            for (const endpoint of configEndpoints) {
                try {
                    const controller = new AbortController();
                    const timeoutId = setTimeout(() => controller.abort(), 3000);
                    
                    const response = await fetch(endpoint, {
                        signal: controller.signal
                    });
                    
                    clearTimeout(timeoutId);
                    
                    if (response.ok) {
                        this.config = await response.json();
                        
                        // Update version and environment in the UI if we have them
                        if (this.config.version) {
                            const versionElement = document.querySelector('.logo-text .version');
                            if (versionElement) {
                                versionElement.textContent = this.config.version.trim();
                            }
                        }
                        
                        // Show dev badge if in development environment
                        if (this.config.environment === 'development') {
                            const envBadge = document.querySelector('.logo-text .env-badge');
                            if (envBadge) {
                                envBadge.textContent = 'dev';
                                envBadge.style.display = 'inline';
                            }
                        }
                        break;
                    }
                } catch (err) {
                    // Silently fail for expected network errors
                    continue;
                }
            }
            
            if (!this.config) {
                this.config = {
                    api_base_url: baseUrl,
                    version: 'v0.0.1',
                    environment: 'fallback'
                };
            }
        } catch (error) {
            // Fallback to current host
            const baseUrl = `${window.location.protocol}//${window.location.host}`;
            this.config = {
                api_base_url: baseUrl,
                version: 'v0.0.1',
                environment: 'fallback'
            };
        }
    },
    
    setupEventHandlers() {
        // Navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const page = item.dataset.page;
                this.showPage(page);
            });
        });
        
        // Sidebar toggle
        document.querySelector('.sidebar-toggle').addEventListener('click', () => {
            this.toggleSidebar();
        });
        
        // Theme toggle
        document.querySelector('.theme-toggle').addEventListener('click', () => {
            this.toggleTheme();
        });
        
        // Button handlers
        document.addEventListener('click', (e) => {
            if (e.target.matches('.test-btn')) {
                this.handleTestButton(e.target);
                this.incrementActivityCounter();
            }
            
            // Copy button handler
            if (e.target.closest('.copy-btn')) {
                const copyBtn = e.target.closest('.copy-btn');
                const commandId = copyBtn.dataset.commandId;
                if (commandId && this.storedCommands && this.storedCommands[commandId]) {
                    const textToCopy = this.storedCommands[commandId];
                    this.copyToClipboard(textToCopy, e);
                }
            }
            
            // Settings handlers
            if (e.target.matches('.test-connection-btn')) {
                this.testConnection(e.target);
            }
            
            if (e.target.matches('#save-settings')) {
                this.saveSettings();
            }
            
            if (e.target.matches('#reset-settings')) {
                this.resetSettings();
            }
        });
    },
    
    toggleSidebar() {
        this.sidebarCollapsed = !this.sidebarCollapsed;
        const sidebar = document.querySelector('.sidebar');
        const toggleIcon = document.querySelector('.sidebar-toggle img');
        
        if (this.sidebarCollapsed) {
            sidebar.classList.add('collapsed');
            toggleIcon.src = 'icons/angle-right.svg';
            toggleIcon.alt = 'expand sidebar';
            // Update theme toggle for collapsed state
            this.updateThemeToggleForCollapsed(true);
        } else {
            sidebar.classList.remove('collapsed');
            toggleIcon.src = 'icons/angle-left.svg';
            toggleIcon.alt = 'collapse sidebar';
            // Update theme toggle for expanded state
            this.updateThemeToggleForCollapsed(false);
        }
    },
    
    updateThemeToggleForCollapsed(isCollapsed) {
        const themeToggle = document.querySelector('.theme-toggle');
        const currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
        
        if (isCollapsed) {
            // Collapsed: show only one icon - moon for light theme, moon-solid for dark theme
            if (currentTheme === 'dark') {
                themeToggle.innerHTML = `
                    <img class="icon" src="icons/moon-solid.svg" alt="switch to light mode">
                `;
            } else {
                themeToggle.innerHTML = `
                    <img class="icon" src="icons/moon.svg" alt="switch to dark mode">
                `;
            }
        } else {
            // Expanded: use moon.svg and sun.svg with toggle switch
            if (currentTheme === 'dark') {
                themeToggle.innerHTML = `
                    <img class="icon" src="icons/moon.svg" alt="dark mode">
                    <div class="toggle-switch">
                        <div class="toggle-knob"></div>
                    </div>
                    <img class="icon" src="icons/sun.svg" alt="light mode">
                `;
            } else {
                themeToggle.innerHTML = `
                    <img class="icon" src="icons/moon.svg" alt="dark mode">
                    <div class="toggle-switch active">
                        <div class="toggle-knob"></div>
                    </div>
                    <img class="icon" src="icons/sun.svg" alt="light mode">
                `;
            }
        }
    },
    
    showPage(pageId) {
        // Update navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-page="${pageId}"]`).classList.add('active');
        
        // Update content
        document.querySelectorAll('.page').forEach(page => {
            page.classList.remove('active');
        });
        document.getElementById(pageId).classList.add('active');
        
        // Load settings when navigating to settings page
        if (pageId === 'settings') {
            this.loadSettings();
        }
        
        this.currentPage = pageId;
    },
    
    startStatusChecking() {
        // Clear any existing interval first
        if (this.statusCheckInterval) {
            clearInterval(this.statusCheckInterval);
            this.statusCheckInterval = null;
        }
        
        // Check immediately
        this.checkLGTMStatus();
        
        // Then check every 30 seconds
        this.statusCheckInterval = setInterval(() => {
            this.checkLGTMStatus();
        }, 30000);
    },
    
    async checkLGTMStatus() {
        // Don't check status if config is not loaded yet
        if (!this.config) {
            return;
        }

        // Use the new backend endpoint that actually checks LGTM services
        try {
            const baseUrl = this.config.api_base_url || `${window.location.protocol}//${window.location.host}`;
            const response = await fetch(`${baseUrl}/lgtm-status`, {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });
            
            if (response.ok) {
                const status = await response.json();
                
                // Update each service status based on actual checks
                this.updateServiceStatus('prometheus-status', status.prometheus || 'offline');
                this.updateServiceStatus('alertmanager-status', status.alertmanager || 'offline');
                this.updateServiceStatus('grafana-status', status.grafana || 'offline');
                this.updateServiceStatus('loki-status', status.loki || 'offline');
                this.updateServiceStatus('tempo-status', status.tempo || 'offline');
            } else {
                // If the endpoint fails, mark all as offline
                const services = ['prometheus-status', 'alertmanager-status', 'grafana-status', 'loki-status', 'tempo-status'];
                services.forEach(service => {
                    this.updateServiceStatus(service, 'offline');
                });
            }
        } catch (error) {
            // If we can't reach our backend, mark all as offline
            const services = ['prometheus-status', 'alertmanager-status', 'grafana-status', 'loki-status', 'tempo-status'];
            services.forEach(service => {
                this.updateServiceStatus(service, 'offline');
            });
        }
    },
    
    updateServiceStatus(elementId, status) {
        const element = document.getElementById(elementId);
        if (element) {
            element.className = `status-indicator status-${status}`;
        }
    },
    
    async handleTestButton(button) {
        const endpoint = button.dataset.endpoint;
        const params = button.dataset.params || '';
        const testName = button.dataset.testName;
        
        // Disable button and show loading
        button.disabled = true;
        button.classList.add('btn-loading');
        
        // Clear previous results
        this.hideResults();
        
        // Show status bar
        this.showStatusBar();
        this.updateStatusBar('Starting test...', 10);
        
        const baseUrl = this.config?.api_base_url || `${window.location.protocol}//${window.location.host}`;
        const fullUrl = `${baseUrl}${endpoint}${params}`;
        const curlCommand = `curl "${fullUrl}"`;
        
        // Use regular fetch for all tests (no more SSE complexity)
        this.handleRegularTest(fullUrl, testName, curlCommand, button);
    },
    
    async handleRegularTest(fullUrl, testName, curlCommand, button) {
        this.updateStatusBar('Sending request...', 30);
        
        try {
            const response = await fetch(fullUrl, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            
            this.updateStatusBar('Processing response...', 70);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const result = await response.json();
            
            this.updateStatusBar('Test completed!', 100);
            
            // Hide status bar after completion
            setTimeout(() => {
                this.hideStatusBar();
                this.showResults(testName, result, curlCommand);
            }, 1000);
            
        } catch (error) {
            this.updateStatusBar(`Request failed: ${error.message}`, 100, true);
            setTimeout(() => {
                this.hideStatusBar();
            }, 3000);
        } finally {
            button.disabled = false;
            button.classList.remove('btn-loading');
        }
    },
    
    showStatusBar() {
        const existingBar = document.querySelector('.status-bar');
        if (existingBar) {
            existingBar.remove();
        }
        
        const statusBarHtml = `
            <div class="status-bar">
                <div class="status-bar-header">test progress</div>
                <div class="status-bar-body">
                    <div class="status-bar-fill" style="width: 0%"></div>
                    <div class="status-bar-text">Starting...</div>
                </div>
            </div>
        `;
        document.querySelector('.page.active').insertAdjacentHTML('beforeend', statusBarHtml);
    },
    
    updateStatusBar(message, percentage, isError = false) {
        const statusBar = document.querySelector('.status-bar');
        if (statusBar) {
            const fill = statusBar.querySelector('.status-bar-fill');
            const text = statusBar.querySelector('.status-bar-text');
            
            if (fill) fill.style.width = `${percentage}%`;
            if (text) text.textContent = message;
            
            if (isError) {
                statusBar.classList.add('status-error');
            }
        }
    },
    
    hideStatusBar() {
        const statusBar = document.querySelector('.status-bar');
        if (statusBar) {
            statusBar.remove();
        }
    },
    
    showResults(testName, result, command) {
        const guidance = this.getTestGuidance(testName, result);
        
        // Store command with unique ID to avoid template literal issues
        const commandId = 'cmd-' + Date.now() + '-' + Math.random().toString(36).slice(2, 11);
        if (!this.storedCommands) {
            this.storedCommands = {};
        }
        this.storedCommands[commandId] = command;
        
        // Determine which fields to show based on the result
        const hasItems = result.items_generated && result.items_generated !== 'N/A' && result.items_generated > 0;
        const hasDuration = result.duration_seconds || result.duration;
        const shouldShowStats = hasItems || (hasDuration && hasDuration !== 'N/A');
        
        let statsHtml = '';
        if (shouldShowStats) {
            statsHtml = '<div class="results-stats">';
            if (hasItems) {
                statsHtml += `<p><strong>Items generated:</strong> ${result.items_generated}</p>`;
            }
            if (hasDuration && hasDuration !== 'N/A') {
                const durationValue = result.duration_seconds || result.duration;
                const durationText = typeof durationValue === 'number' ? `${durationValue.toFixed(2)}s` : durationValue;
                statsHtml += `<p><strong>Duration:</strong> ${durationText}</p>`;
            }
            statsHtml += '</div>';
        }
        
        const resultsHtml = `
            <div class="results-section show">
                <div class="results-header">
                    <h3 class="results-title">Test Completed: ${testName}</h3>
                </div>
                <div class="results-content">
                    <div class="command-section">
                        <strong>Command:</strong> 
                        <div class="command-container">
                            <code class="command-text">${command}</code>
                            <button class="copy-btn" data-command-id="${commandId}">
                                <img src="icons/copy.svg" alt="copy" class="icon">
                            </button>
                        </div>
                    </div>
                    ${statsHtml}
                </div>
                <div class="guidance-section">
                    <div class="guidance-title">Verification Steps:</div>
                    <ul class="guidance-steps">
                        ${guidance.map(step => `<li>${step}</li>`).join('')}
                    </ul>
                </div>
            </div>
        `;
        
        document.querySelector('.page.active').insertAdjacentHTML('beforeend', resultsHtml);
    },
    
    copyToClipboard(text, event) {
        navigator.clipboard.writeText(text).then(() => {
            // Visual feedback
            const copyBtn = event.target.closest('.copy-btn');
            if (copyBtn) {
                const originalHtml = copyBtn.innerHTML;
                copyBtn.innerHTML = '<img src="icons/check.svg" alt="copied" class="icon">';
                setTimeout(() => {
                    copyBtn.innerHTML = originalHtml;
                }, 1000);
            }
        }).catch(err => {
            console.error('Failed to copy to clipboard:', err);
            // Show user-friendly error
            const copyBtn = event.target.closest('.copy-btn');
            if (copyBtn) {
                const originalHtml = copyBtn.innerHTML;
                copyBtn.innerHTML = '<span style="color: var(--error-color); font-size: 12px;">failed</span>';
                setTimeout(() => {
                    copyBtn.innerHTML = originalHtml;
                }, 2000);
            }
        });
    },
    
    hideResults() {
        document.querySelectorAll('.results-section').forEach(section => {
            section.remove();
        });
    },
    
    getTestGuidance(testName, result) {
        const baseGuidance = {
            'LGTM Integration Test': [
                '1. <a href="http://localhost:3000" target="_blank">Open Grafana</a> → Check if datasources are connected',
                '2. <a href="http://localhost:9090" target="_blank">Open Prometheus</a> → Verify targets are up in Status > Targets',
                '3. <a href="http://localhost:3100/ready" target="_blank">Check Loki</a> → Should return "ready"',
                '4. <a href="http://localhost:3200/ready" target="_blank">Check Tempo</a> → Should return "ready"',
                '5. Review test results above for specific component status details'
            ],
            'Grafana Dashboard Test': [
                '1. <a href="http://localhost:3000/d/argus-test-dashboard" target="_blank">Open Argus Testing Dashboard</a> (created by test)',
                '2. Verify all panels are visible: Performance, System Resources, LGTM Health, Metrics, Logs, Test Status',
                '3. Check data is flowing - run some performance/data tests to populate the dashboard',
                '4. <a href="http://localhost:3000/dashboards" target="_blank">View all dashboards</a> to confirm creation'
            ],
            'Alert Rules Test': [
                '1. <a href="http://localhost:9090/api/v1/rules" target="_blank">Check Rules API</a> - Verify Prometheus rules endpoint is accessible',
                '2. <a href="http://localhost:9090/rules" target="_blank">View Rules Configuration</a> - See all loaded rule groups and alerts',
                '3. <a href="http://localhost:9090/alerts" target="_blank">Monitor Active Alerts</a> - Check firing and pending alerts',
                '4. Look for Argus-specific rules in the test results - they should be detected automatically',
                '5. Test rule evaluation by generating system load (CPU/Memory alerts fire at >50% usage)',
                '6. Verify alerts API accessibility and alert state tracking'
            ],
            'Metrics Scale Test': [
                '1. <strong>Important:</strong> Prometheus must be configured to scrape Argus metrics at <code>' + (this.config?.api_base_url || `${window.location.protocol}//${window.location.host}`) + '/metrics</code>',
                '2. Verify metrics are exposed: <a href="' + (this.config?.api_base_url || `${window.location.protocol}//${window.location.host}`) + '/metrics" target="_blank">Check Argus /metrics endpoint</a>',
                '3. <a href="http://localhost:9090" target="_blank">Open Prometheus</a> → Go to Graph tab',
                '4. Try these exact queries (copy and paste):',
                '   • <code>custom_business_metric{type="performance_test"}</code> (main test metric)',
                '   • <code>http_requests_total{method="GET", endpoint="/api/scale-test"}</code> (GET requests)', 
                '   • <code>http_requests_total{endpoint="/api/scale-test"}</code> (all HTTP requests)', 
                '   • <code>rate(http_requests_total{endpoint="/api/scale-test"}[5m])</code> (request rate)',
                '5. Set time range to "Last 5 minutes" or "Last 15 minutes"',
                '6. <strong>If no data:</strong> Add this to your prometheus.yml scrape_configs: <code>- job_name: "argus" static_configs: - targets: ["localhost:3001"]</code>'
            ],
            'Logs Scale Test': [
                '1. <a href="http://localhost:3000" target="_blank">Open Grafana</a>',
                '2. Navigate to Explore > Loki data source',
                '3. Use query: {service="argus"} | json',
                '4. Filter by time range when the test was run to see generated logs'
            ],
            'Traces Scale Test': [
                '1. <a href="http://localhost:3000" target="_blank">Open Grafana</a>',
                '2. Navigate to Explore > Tempo data source',
                '3. Search for service name "argus" or operation names',
                '4. Look for traces generated during the test timeframe'
            ],
            'Generate Metrics': [
                '1. <a href="http://localhost:9090" target="_blank">Open Prometheus</a>',
                '2. Search for "argus_test_metric" in the Graph tab',
                '3. View the metric values and timestamps',
                '4. Verify the metric count matches what was generated'
            ],
            'Generate Logs': [
                '1. <a href="http://localhost:3000" target="_blank">Open Grafana</a>',
                '2. Go to Explore > Loki and query: {service="argus"}',
                '3. See the structured logs generated by the test',
                '4. Verify log count and format match expectations'
            ]
        };
        
        return baseGuidance[testName] || [
            '1. Check the appropriate LGTM stack component for the generated data',
            '2. Use the time range when the test was executed',
            '3. Look for "argus" labels or service names in your queries'
        ];
    },
    
    toggleTheme() {
        const currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', newTheme);
        
        // Update theme toggle based on current sidebar state
        this.updateThemeToggleForCollapsed(this.sidebarCollapsed);
    },
    
    incrementActivityCounter() {
        const counters = ['tests-run', 'metrics-generated', 'logs-generated', 'traces-generated'];
        counters.forEach(id => {
            const element = document.getElementById(id);
            if (element) {
                const current = parseInt(element.textContent) || 0;
                element.textContent = current + 1;
            }
        });
    },
    
    updateActivitySummary() {
        // Get today's date
        const today = new Date().toDateString();
        const storedDate = localStorage.getItem('argus-activity-date');
        
        // Reset counters if it's a new day
        if (storedDate !== today) {
            localStorage.setItem('argus-activity-date', today);
            localStorage.setItem('argus-tests-run', '0');
            localStorage.setItem('argus-metrics-generated', '0');
            localStorage.setItem('argus-logs-generated', '0');
            localStorage.setItem('argus-traces-generated', '0');
        }
        
        // Update display
        document.getElementById('tests-run').textContent = localStorage.getItem('argus-tests-run') || '0';
        document.getElementById('metrics-generated').textContent = localStorage.getItem('argus-metrics-generated') || '0';
        document.getElementById('logs-generated').textContent = localStorage.getItem('argus-logs-generated') || '0';
        document.getElementById('traces-generated').textContent = localStorage.getItem('argus-traces-generated') || '0';
    },

    // Settings Management
    loadSettings() {
        const defaultSettings = {
            grafana: {
                url: 'http://localhost:3000',
                username: 'admin',
                password: ''
            },
            prometheus: {
                url: 'http://localhost:9090',
                username: '',
                password: ''
            },
            loki: {
                url: 'http://localhost:3100'
            },
            tempo: {
                url: 'http://localhost:3200'
            }
        };

        const saved = localStorage.getItem('argus-settings');
        const settings = saved ? JSON.parse(saved) : defaultSettings;
        
        // Populate form fields
        document.getElementById('grafana-url').value = settings.grafana?.url || defaultSettings.grafana.url;
        document.getElementById('grafana-username').value = settings.grafana?.username || defaultSettings.grafana.username;
        document.getElementById('grafana-password').value = settings.grafana?.password || defaultSettings.grafana.password;
        
        document.getElementById('prometheus-url').value = settings.prometheus?.url || defaultSettings.prometheus.url;
        document.getElementById('prometheus-username').value = settings.prometheus?.username || defaultSettings.prometheus.username;
        document.getElementById('prometheus-password').value = settings.prometheus?.password || defaultSettings.prometheus.password;
        
        document.getElementById('loki-url').value = settings.loki?.url || defaultSettings.loki.url;
        document.getElementById('tempo-url').value = settings.tempo?.url || defaultSettings.tempo.url;
        
        return settings;
    },

    saveSettings() {
        const settings = {
            grafana: {
                url: document.getElementById('grafana-url').value,
                username: document.getElementById('grafana-username').value,
                password: document.getElementById('grafana-password').value
            },
            prometheus: {
                url: document.getElementById('prometheus-url').value,
                username: document.getElementById('prometheus-username').value,
                password: document.getElementById('prometheus-password').value
            },
            loki: {
                url: document.getElementById('loki-url').value
            },
            tempo: {
                url: document.getElementById('tempo-url').value
            }
        };

        localStorage.setItem('argus-settings', JSON.stringify(settings));
        
        // Show feedback
        const saveBtn = document.getElementById('save-settings');
        const originalText = saveBtn.textContent;
        saveBtn.textContent = 'saved!';
        saveBtn.disabled = true;
        
        setTimeout(() => {
            saveBtn.textContent = originalText;
            saveBtn.disabled = false;
        }, 2000);

        // Also save to backend
        const baseUrl = this.config?.api_base_url || `${window.location.protocol}//${window.location.host}`;
        fetch(`${baseUrl}/api/settings`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(settings)
        }).catch(err => console.log('Backend save failed:', err));
    },

    resetSettings() {
        localStorage.removeItem('argus-settings');
        this.loadSettings();
    },

    async testConnection(button) {
        const service = button.dataset.service;
        const settings = this.getCurrentSettings();
        
        button.disabled = true;
        button.textContent = 'testing...';
        
        // Remove any existing status indicators
        const existingStatus = button.parentNode.querySelector('.connection-status');
        if (existingStatus) {
            existingStatus.remove();
        }
        
        try {
            const baseUrl = this.config?.api_base_url || `${window.location.protocol}//${window.location.host}`;
            const response = await fetch(`${baseUrl}/api/test-connection/${service}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(settings[service])
            });
            
            const result = await response.json();
            
            // Create status indicator
            const statusEl = document.createElement('span');
            statusEl.className = `connection-status ${result.status === 'success' ? 'success' : 'error'}`;
            statusEl.textContent = result.status === 'success' ? 'connected' : 'failed';
            
            button.parentNode.appendChild(statusEl);
            
        } catch (error) {
            const statusEl = document.createElement('span');
            statusEl.className = 'connection-status error';
            statusEl.textContent = 'error';
            button.parentNode.appendChild(statusEl);
        } finally {
            button.disabled = false;
            button.textContent = 'test connection';
        }
    },

    getCurrentSettings() {
        return {
            grafana: {
                url: document.getElementById('grafana-url').value,
                username: document.getElementById('grafana-username').value,
                password: document.getElementById('grafana-password').value
            },
            prometheus: {
                url: document.getElementById('prometheus-url').value,
                username: document.getElementById('prometheus-username').value,
                password: document.getElementById('prometheus-password').value
            },
            loki: {
                url: document.getElementById('loki-url').value
            },
            tempo: {
                url: document.getElementById('tempo-url').value
            }
        };
    }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
    await window.ArgusApp.init();
}); 