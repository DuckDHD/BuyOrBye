/**
 * Offline Support and Service Worker Registration for BuyOrBye
 * Provides graceful offline experience and app-like functionality
 */

// Service Worker registration
if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
            .then(registration => {
                console.log('SW registered:', registration);

                // Check for updates
                registration.addEventListener('updatefound', () => {
                    const newWorker = registration.installing;
                    newWorker.addEventListener('statechange', () => {
                        if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
                            // Show update available notification
                            showUpdateNotification();
                        }
                    });
                });
            })
            .catch(error => {
                console.log('SW registration failed:', error);
            });
    });
}

// Offline detection and handling
class OfflineHandler {
    constructor() {
        this.isOnline = navigator.onLine;
        this.pendingMessages = JSON.parse(localStorage.getItem('pending_messages') || '[]');
        this.offlineBanner = null;

        this.setupEventListeners();
        this.checkInitialState();
    }

    setupEventListeners() {
        window.addEventListener('online', this.handleOnline.bind(this));
        window.addEventListener('offline', this.handleOffline.bind(this));

        // Listen for failed requests
        window.addEventListener('fetch-failed', this.handleFailedRequest.bind(this));
    }

    checkInitialState() {
        if (!this.isOnline) {
            this.showOfflineMessage();
        }

        // Process any pending messages on load
        if (this.pendingMessages.length > 0) {
            this.showPendingMessagesNotification();
        }
    }

    handleOnline() {
        console.log('Connection restored');
        this.isOnline = true;
        this.removeOfflineMessage();
        this.retryPendingMessages();
        this.showOnlineNotification();
    }

    handleOffline() {
        console.log('Connection lost');
        this.isOnline = false;
        this.showOfflineMessage();
    }

    handleFailedRequest(event) {
        if (event.detail && event.detail.type === 'chat') {
            this.storePendingMessage(event.detail.message);
        }
    }

    showOfflineMessage() {
        if (this.offlineBanner) return;

        this.offlineBanner = document.createElement('div');
        this.offlineBanner.id = 'offline-banner';
        this.offlineBanner.className = 'fixed top-0 left-0 right-0 bg-yellow-500 text-white px-4 py-2 text-sm text-center z-50 transform -translate-y-full transition-transform duration-300';
        this.offlineBanner.innerHTML = `
            <div class="flex items-center justify-center space-x-2">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
                </svg>
                <span>You're offline. Messages will be sent when connection is restored.</span>
            </div>
        `;

        document.body.appendChild(this.offlineBanner);

        // Animate in
        requestAnimationFrame(() => {
            this.offlineBanner.classList.remove('-translate-y-full');
        });
    }

    removeOfflineMessage() {
        if (!this.offlineBanner) return;

        this.offlineBanner.classList.add('-translate-y-full');
        setTimeout(() => {
            if (this.offlineBanner && this.offlineBanner.parentNode) {
                this.offlineBanner.parentNode.removeChild(this.offlineBanner);
                this.offlineBanner = null;
            }
        }, 300);
    }

    showOnlineNotification() {
        this.showNotification('🌐 Back online! Syncing messages...', 'success');
    }

    showPendingMessagesNotification() {
        const count = this.pendingMessages.length;
        this.showNotification(`📝 You have ${count} message${count > 1 ? 's' : ''} waiting to be sent`, 'info');
    }

    storePendingMessage(message) {
        const pendingMessage = {
            id: Date.now(),
            text: message,
            timestamp: new Date().toISOString()
        };

        this.pendingMessages.push(pendingMessage);
        localStorage.setItem('pending_messages', JSON.stringify(this.pendingMessages));

        this.showNotification('💾 Message saved. Will send when online.', 'info');
    }

    async retryPendingMessages() {
        if (this.pendingMessages.length === 0) return;

        const messages = [...this.pendingMessages];
        this.pendingMessages = [];
        localStorage.removeItem('pending_messages');

        let successCount = 0;
        for (const message of messages) {
            try {
                // Try to send the message via the global chat interface
                if (window.chatInterface && typeof window.chatInterface.sendMessage === 'function') {
                    await window.chatInterface.sendMessage(message.text, true);
                    successCount++;
                }
            } catch (error) {
                console.log('Failed to retry message:', error);
                // Re-add failed messages back to pending
                this.pendingMessages.push(message);
            }
        }

        if (this.pendingMessages.length > 0) {
            localStorage.setItem('pending_messages', JSON.stringify(this.pendingMessages));
        }

        if (successCount > 0) {
            this.showNotification(`✅ ${successCount} message${successCount > 1 ? 's' : ''} sent successfully!`, 'success');
        }
    }

    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 z-50 px-4 py-3 rounded-lg text-white text-sm max-w-sm transform translate-x-full transition-transform duration-300 ${this.getNotificationClass(type)}`;
        notification.textContent = message;

        document.body.appendChild(notification);

        // Animate in
        requestAnimationFrame(() => {
            notification.classList.remove('translate-x-full');
        });

        // Auto remove after 4 seconds
        setTimeout(() => {
            notification.classList.add('translate-x-full');
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 300);
        }, 4000);
    }

    getNotificationClass(type) {
        switch (type) {
            case 'success': return 'bg-green-500';
            case 'error': return 'bg-red-500';
            case 'warning': return 'bg-yellow-500';
            default: return 'bg-blue-500';
        }
    }

    // Public API for other scripts
    isOffline() {
        return !this.isOnline;
    }

    getPendingMessageCount() {
        return this.pendingMessages.length;
    }
}

// App update notification
function showUpdateNotification() {
    const notification = document.createElement('div');
    notification.className = 'fixed bottom-4 left-4 right-4 bg-blue-600 text-white p-4 rounded-lg shadow-lg z-50 mx-auto max-w-sm';
    notification.innerHTML = `
        <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
                <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd"></path>
                </svg>
                <span class="font-medium">Update Available</span>
            </div>
            <button onclick="location.reload()" class="bg-white text-blue-600 px-3 py-1 rounded text-sm font-medium hover:bg-blue-50">
                Update
            </button>
        </div>
    `;

    document.body.appendChild(notification);

    // Auto-remove after 10 seconds
    setTimeout(() => {
        if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
        }
    }, 10000);
}

// Connection quality monitoring
class ConnectionQuality {
    constructor() {
        this.connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
        this.setupConnectionMonitoring();
    }

    setupConnectionMonitoring() {
        if (!this.connection) return;

        this.connection.addEventListener('change', () => {
            this.handleConnectionChange();
        });

        // Initial check
        this.handleConnectionChange();
    }

    handleConnectionChange() {
        if (!this.connection) return;

        const { effectiveType, downlink, rtt } = this.connection;

        // Warn about slow connections
        if (effectiveType === 'slow-2g' || effectiveType === '2g') {
            this.showSlowConnectionWarning();
        }

        // Log connection details for debugging
        console.log('Connection changed:', {
            effectiveType,
            downlink: `${downlink} Mbps`,
            rtt: `${rtt} ms`
        });
    }

    showSlowConnectionWarning() {
        // Only show once per session
        if (sessionStorage.getItem('slow_connection_warned')) return;
        sessionStorage.setItem('slow_connection_warned', 'true');

        const warning = document.createElement('div');
        warning.className = 'fixed bottom-4 left-4 right-4 bg-orange-500 text-white p-3 rounded-lg shadow-lg z-50 mx-auto max-w-sm';
        warning.innerHTML = `
            <div class="flex items-center space-x-2">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"></path>
                </svg>
                <span class="text-sm">Slow connection detected. Messages may take longer to send.</span>
            </div>
        `;

        document.body.appendChild(warning);

        setTimeout(() => {
            if (warning.parentNode) {
                warning.parentNode.removeChild(warning);
            }
        }, 5000);
    }
}

// Initialize offline handling and connection monitoring
document.addEventListener('DOMContentLoaded', () => {
    window.offlineHandler = new OfflineHandler();
    window.connectionQuality = new ConnectionQuality();

    // Add global error handler for failed fetch requests
    window.addEventListener('unhandledrejection', (event) => {
        if (event.reason && event.reason.name === 'TypeError' && event.reason.message.includes('fetch')) {
            console.log('Fetch failed, likely due to network issues');
            // Don't prevent default to allow other error handlers to run
        }
    });
});

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { OfflineHandler, ConnectionQuality };
}