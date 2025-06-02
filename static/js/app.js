// Configuration and state management
window.ArgusApp = {
    config: null,
    currentPage: 'home',
    sidebarCollapsed: false,
    statusCheckInterval: null,
    
    init() {
        this.loadConfiguration();
        this.setupEventHandlers();
        this.showPage('home');
        this.startStatusChecking();
        this.updateActivitySummary();
    },
    
    async loadConfiguration() {
        try {
            // Always use localhost for internal network
            const configEndpoints = [
                'http://localhost:3001/config',
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
                        break;
                    }
                } catch (err) {
                    // Silently fail for expected network errors
                    continue;
                }
            }
            
            if (!this.config) {
                this.config = {
                    api_base_url: 'http://localhost:3001',
                    version: 'v0.0.1',
                    environment: 'fallback'
                };
            }
            
            this.checkLGTMStatus();
        } catch (error) {
            // Fallback to localhost
            this.config = {
                api_base_url: 'http://localhost:3001',
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
        
        this.currentPage = pageId;
    },
    
    startStatusChecking() {
        // Check immediately
        this.checkLGTMStatus();
        
        // Then check every 30 seconds
        this.statusCheckInterval = setInterval(() => {
            this.checkLGTMStatus();
        }, 30000);
    },
    
    async checkLGTMStatus() {
        // Use the new backend endpoint that actually checks LGTM services
        try {
            const response = await fetch('http://localhost:3001/lgtm-status', {
                method: 'GET',
                headers: { 'Content-Type': 'application/json' }
            });
            
            if (response.ok) {
                const status = await response.json();
                
                // Update each service status based on actual checks
                this.updateServiceStatus('prometheus-status', status.prometheus || 'offline');
                this.updateServiceStatus('grafana-status', status.grafana || 'offline');
                this.updateServiceStatus('loki-status', status.loki || 'offline');
                this.updateServiceStatus('tempo-status', status.tempo || 'offline');
            } else {
                // If the endpoint fails, mark all as offline
                const services = ['prometheus-status', 'grafana-status', 'loki-status', 'tempo-status'];
                services.forEach(service => {
                    this.updateServiceStatus(service, 'offline');
                });
            }
        } catch (error) {
            // If we can't reach our backend, mark all as offline
            const services = ['prometheus-status', 'grafana-status', 'loki-status', 'tempo-status'];
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
        
        // Show terminal and start progress
        this.showTerminal();
        this.addTerminalLine('', 'prompt');
        
        // Force localhost for all requests
        const fullUrl = `http://localhost:3001${endpoint}${params}`;
        const curlCommand = `curl "${fullUrl}"`;
        
        this.addTerminalLine(`$ ${curlCommand}`, 'command');
        this.addTerminalLine('Starting test...', 'success');
        
        try {
            const response = await fetch(fullUrl, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const result = await response.json();
            
            this.addTerminalLine(`Test completed successfully`, 'success');
            this.addTerminalLine(`Generated ${result.items_generated || 'N/A'} items`, 'success');
            this.addTerminalLine(`Duration: ${result.duration_ms || result.duration || 'N/A'}`, 'success');
            
            // Show results and guidance
            this.showResults(testName, result, curlCommand);
        } catch (error) {
            this.addTerminalLine(`Request failed: ${error.message}`, 'error');
        } finally {
            // Re-enable button
            button.disabled = false;
            button.classList.remove('btn-loading');
            this.addTerminalLine('$', 'prompt');
        }
    },
    
    showTerminal() {
        const terminal = document.querySelector('.terminal');
        if (terminal) {
            terminal.style.display = 'block';
        } else {
            // Create terminal if it doesn't exist
            const terminalHtml = `
                <div class="terminal">
                    <div class="terminal-header">
                        argus terminal
                    </div>
                    <div class="terminal-body" id="terminal-output"></div>
                </div>
            `;
            document.querySelector('.page.active').insertAdjacentHTML('beforeend', terminalHtml);
        }
    },
    
    addTerminalLine(text, type = '') {
        const output = document.getElementById('terminal-output');
        const line = document.createElement('div');
        line.className = `terminal-line ${type ? 'terminal-' + type : ''}`;
        line.textContent = text;
        output.appendChild(line);
        output.scrollTop = output.scrollHeight;
    },
    
    showResults(testName, result, command) {
        const guidance = this.getTestGuidance(testName, result);
        
        const resultsHtml = `
            <div class="results-section show">
                <div class="results-header">
                    <h3 class="results-title">Test Completed: ${testName}</h3>
                </div>
                <div class="results-content">
                    <p><strong>Command:</strong> <code>${command}</code></p>
                    <p><strong>Items generated:</strong> ${result.items_generated || 'N/A'}</p>
                    <p><strong>Duration:</strong> ${result.duration_ms || result.duration || 'N/A'}</p>
                </div>
                <div class="guidance-section">
                    <div class="guidance-title">Where to see results:</div>
                    <ul class="guidance-steps">
                        ${guidance.map(step => `<li>${step}</li>`).join('')}
                    </ul>
                </div>
            </div>
        `;
        
        document.querySelector('.page.active').insertAdjacentHTML('beforeend', resultsHtml);
    },
    
    hideResults() {
        document.querySelectorAll('.results-section').forEach(section => {
            section.remove();
        });
    },
    
    getTestGuidance(testName, result) {
        const baseGuidance = {
            'Metrics Scale Test': [
                'Open Prometheus at http://localhost:9090',
                'Go to Graph tab and search for "argus" or "performance_test"',
                'Look for custom metrics like "custom_metric" and "http_requests_total"',
                'Check the time range to see the generated data points'
            ],
            'Logs Scale Test': [
                'Open Grafana at http://localhost:3000',
                'Navigate to Explore > Loki data source',
                'Use query: {service="argus"} | json',
                'Filter by time range when the test was run to see generated logs'
            ],
            'Traces Scale Test': [
                'Open Grafana at http://localhost:3000',
                'Navigate to Explore > Tempo data source',
                'Search for service name "argus" or operation names',
                'Look for traces generated during the test timeframe'
            ],
            'Generate Metrics': [
                'Open Prometheus at http://localhost:9090',
                'Search for "argus_test_metric" in the Graph tab',
                'View the metric values and timestamps'
            ],
            'Generate Logs': [
                'Open Grafana at http://localhost:3000',
                'Go to Explore > Loki and query: {service="argus"}',
                'See the structured logs generated by the test'
            ]
        };
        
        return baseGuidance[testName] || [
            'Check the appropriate LGTM stack component for the generated data',
            'Use the time range when the test was executed',
            'Look for "argus" labels or service names in your queries'
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
    }
};

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.ArgusApp.init();
}); 