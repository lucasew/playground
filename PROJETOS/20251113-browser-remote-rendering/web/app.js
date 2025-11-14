class RemoteBrowser {
    constructor() {
        this.userID = null; // Will be set after successful login
        this.currentTabID = null;
        this.htmlStream = null;
        this.tabsStream = null;
        this.tabs = [];
        this.password = null; // No longer needed with cookie auth

        this.init();
    }

    async init() {
        try {
            const response = await fetch('/api/check-session', { credentials: 'include' });
            if (response.ok) {
                const data = await response.json(); // Parse JSON response
                this.userID = data.userID; // Get userID from response
                sessionStorage.setItem('username', this.userID); // Store in sessionStorage
                this.startApp();
            } else {
                this.showLogin();
            }
        } catch (error) {
            console.error('Session check failed:', error);
            this.showLogin();
        }
    }

    showLogin() {
        document.getElementById('main-app-container').style.display = 'none';
        document.getElementById('login-container').style.display = 'flex'; // Use flex for centering
        document.getElementById('login-btn').addEventListener('click', () => this.handleLogin());
    }

    async handleLogin() {
        const usernameInput = document.getElementById('login-username');
        const passwordInput = document.getElementById('login-password');
        const errorDisplay = document.getElementById('login-error');

        const username = usernameInput.value;
        const password = passwordInput.value;

        errorDisplay.textContent = ''; // Clear previous errors

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
                credentials: 'include'
            });

            if (response.ok) {
                sessionStorage.setItem('username', username); // Store username in sessionStorage
                this.userID = username; // Set userID after successful login
                this.startApp();
            } else {
                const errorText = await response.text();
                errorDisplay.textContent = `Login failed: ${errorText}`;
            }
        } catch (error) {
            console.error('Login error:', error);
            errorDisplay.textContent = 'An error occurred during login.';
        }
    }

    startApp() {
        document.getElementById('login-container').style.display = 'none';
        document.getElementById('main-app-container').style.display = 'flex'; // Assuming flex for app-container

        // Set username display
        document.getElementById('username').textContent = this.userID;

        // Setup event listeners
        document.getElementById('new-tab-btn').addEventListener('click', () => this.newTab());
        document.getElementById('navigate-btn').addEventListener('click', () => this.navigateCurrentTab());
        document.getElementById('url-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.navigateCurrentTab();
        });

        // Setup browser content interactions
        this.setupBrowserInteractions();

        // Initialize tabs stream
        this.initTabsStream();
    }

    initTabsStream() {
        // Cookies will handle authentication, no query params needed
        this.tabsStream = new EventSource(`/api/tabs/${this.userID}`);

        this.tabsStream.addEventListener('tabs', (e) => {
            const data = JSON.parse(e.data);
            this.updateTabList(data.tabs);

            // If active tab changed, reconnect HTML stream
            if (data.activeTabID !== this.currentTabID) {
                this.switchToTab(data.activeTabID);
            }
        });

        this.tabsStream.onerror = (e) => {
            console.error('Tabs stream error:', e);
            this.setStatus('Connection error - retrying...');
            // EventSource will auto-reconnect
        };
    }

    switchToTab(tabID) {
        if (!tabID) {
            this.currentTabID = null;
            this.showWelcome();
            return;
        }

        // Close previous stream
        if (this.htmlStream) {
            this.htmlStream.close();
        }

        this.currentTabID = tabID;

        // Update URL input
        const tab = this.tabs.find(t => t.id === tabID);
        if (tab) {
            document.getElementById('url-input').value = tab.url;
        }

        // Open new stream
        // Cookies will handle authentication, no query params needed
        this.htmlStream = new EventSource(`/api/stream/${this.userID}/${tabID}`);

        this.htmlStream.addEventListener('html', (e) => {
            const data = JSON.parse(e.data);
            this.displayHTML(data.content);
        });

        this.htmlStream.onerror = (e) => {
            console.error('HTML stream error:', e);
        };

        this.setStatus('Connected');
    }

    updateTabList(tabs) {
        this.tabs = tabs;
        const tabsList = document.getElementById('tabs-list');
        tabsList.innerHTML = '';

        tabs.forEach(tab => {
            const tabEl = document.createElement('div');
            tabEl.className = 'tab';
            if (tab.id === this.currentTabID) {
                tabEl.classList.add('active');
            }

            const titleEl = document.createElement('span');
            titleEl.className = 'tab-title';
            titleEl.textContent = tab.title || 'New Tab';
            titleEl.onclick = () => this.sendAction({ action: 'switchTab', tabId: tab.id });

            const closeBtn = document.createElement('button');
            closeBtn.className = 'tab-close';
            closeBtn.textContent = '×';
            closeBtn.onclick = (e) => {
                e.stopPropagation();
                this.closeTab(tab.id);
            };

            tabEl.appendChild(titleEl);
            tabEl.appendChild(closeBtn);
            tabsList.appendChild(tabEl);
        });
    }

    setupBrowserInteractions() {
        const iframe = document.getElementById('browser-iframe');

        const attachListeners = () => {
            if (!iframe || !iframe.contentDocument) return;

            const contentDocument = iframe.contentDocument;
            const contentWindow = iframe.contentWindow;

            // Disable all links and forms to prevent navigation
            contentDocument.querySelectorAll('a').forEach(a => {
                a.addEventListener('click', (e) => {
                    e.preventDefault();
                    const href = a.getAttribute('href');
                    if (href && !href.startsWith('#')) {
                        this.navigate(new URL(href, document.getElementById('url-input').value).href);
                    }
                });
            });

            contentDocument.querySelectorAll('form').forEach(form => {
                form.addEventListener('submit', (e) => e.preventDefault());
            });

            // Capture clicks
            contentDocument.addEventListener('click', (e) => {
                if (!this.currentTabID) return;

                // Get coordinates relative to the iframe's content
                const x = e.clientX;
                const y = e.clientY;

                this.sendTabAction({ action: 'click', x, y });
            });

            // Capture keyboard input
            contentDocument.addEventListener('keydown', (e) => {
                if (!this.currentTabID) return;

                // Send the key
                this.sendTabAction({ action: 'type', text: e.key });
            });

            // Capture scroll
            let scrollTimeout;
            contentWindow.addEventListener('scroll', (e) => {
                if (!this.currentTabID) return;

                clearTimeout(scrollTimeout);
                scrollTimeout = setTimeout(() => {
                    this.sendTabAction({
                        action: 'scrollTo',
                        x: contentWindow.scrollX,
                        y: contentWindow.scrollY
                    });
                }, 500);
            });
        };

        // Attach listeners initially and whenever the iframe content loads
        iframe.onload = attachListeners;
        attachListeners(); // Call once in case content is already loaded (e.g., from showWelcome)
    }

    displayHTML(html) {
        const iframe = document.getElementById('browser-iframe');
        if (iframe && iframe.contentDocument) {
            iframe.contentDocument.open();
            iframe.contentDocument.write(html);
            iframe.contentDocument.close();
            // The onload event will trigger attachListeners
        }
    }

    async newTab() {
        const url = prompt('Enter URL for new tab:', 'https://example.com');
        if (!url) return;

        try {
            const result = await this.sendAction({ action: 'newTab', url });
            this.setStatus('New tab created');
        } catch (err) {
            console.error('Failed to create tab:', err);
            this.setStatus('Failed to create tab');
        }
    }

    async closeTab(tabID) {
        try {
            await this.sendAction({ action: 'closeTab', tabId: tabID });
        } catch (err) {
            console.error('Failed to close tab:', err);
        }
    }

    async navigateCurrentTab() {
        if (!this.currentTabID) {
            alert('No active tab. Create a new tab first.');
            return;
        }

        const url = document.getElementById('url-input').value;
        if (!url) return;

        await this.navigate(url);
    }

    async navigate(url) {
        if (!this.currentTabID) return;

        try {
            await this.sendTabAction({ action: 'navigate', url });
            this.setStatus('Navigating...');
        } catch (err) {
            console.error('Navigation failed:', err);
            this.setStatus('Navigation failed');
        }
    }

    async sendAction(action) {
        const url = `/api/action/${this.userID}`; // Re-adding this line
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(action),
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }

        return response.json();
    }

    async sendTabAction(action) {
        if (!this.currentTabID) return;

        const url = `/api/action/${this.userID}/${this.currentTabID}`;
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(action),
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }

        return response.json();
    }

    setStatus(message) {
        document.getElementById('status').textContent = message;
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.browser = new RemoteBrowser();
});
