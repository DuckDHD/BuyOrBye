/**
 * Service Worker for BuyOrBye PWA
 * Provides offline caching and background sync capabilities
 */

const CACHE_NAME = 'buyorbye-v1.0.0';
const STATIC_CACHE = 'buyorbye-static-v1.0.0';

// Assets to cache immediately
const PRECACHE_ASSETS = [
    '/',
    '/static/css/output.css',
    '/static/css/chat.css',
    '/static/js/chat.js',
    '/static/js/offline.js',
    '/static/manifest.json',
    '/api/health'
];

// Cache strategies
const RUNTIME_CACHE = {
    // API responses - Network first, cache fallback
    api: {
        pattern: /^\/api\//,
        strategy: 'networkFirst',
        maxAge: 5 * 60, // 5 minutes
        maxEntries: 50
    },

    // Static assets - Cache first
    static: {
        pattern: /\.(js|css|png|jpg|jpeg|svg|ico|woff|woff2)$/,
        strategy: 'cacheFirst',
        maxAge: 7 * 24 * 60 * 60, // 7 days
        maxEntries: 100
    },

    // Pages - Network first with offline fallback
    pages: {
        pattern: /^\/[^\/]*$/,
        strategy: 'networkFirst',
        maxAge: 24 * 60 * 60, // 24 hours
        maxEntries: 20
    }
};

// Install event - precache assets
self.addEventListener('install', event => {
    console.log('SW: Install event');

    event.waitUntil(
        Promise.all([
            // Cache static assets
            caches.open(STATIC_CACHE).then(cache => {
                console.log('SW: Precaching static assets');
                return cache.addAll(PRECACHE_ASSETS);
            }),

            // Skip waiting to activate immediately
            self.skipWaiting()
        ])
    );
});

// Activate event - cleanup old caches
self.addEventListener('activate', event => {
    console.log('SW: Activate event');

    event.waitUntil(
        Promise.all([
            // Clean up old caches
            caches.keys().then(cacheNames => {
                return Promise.all(
                    cacheNames
                        .filter(name => name !== CACHE_NAME && name !== STATIC_CACHE)
                        .map(name => {
                            console.log('SW: Deleting old cache', name);
                            return caches.delete(name);
                        })
                );
            }),

            // Take control of all clients
            self.clients.claim()
        ])
    );
});

// Fetch event - implement caching strategies
self.addEventListener('fetch', event => {
    const { request } = event;
    const url = new URL(request.url);

    // Skip non-GET requests
    if (request.method !== 'GET') {
        return;
    }

    // Skip external requests
    if (url.origin !== self.location.origin) {
        return;
    }

    // Apply caching strategy based on request type
    const strategy = getCacheStrategy(url.pathname);

    if (strategy) {
        event.respondWith(handleRequest(request, strategy));
    }
});

// Background sync for offline chat messages
self.addEventListener('sync', event => {
    console.log('SW: Background sync event', event.tag);

    if (event.tag === 'chat-messages') {
        event.waitUntil(syncChatMessages());
    }
});

// Push notification handling
self.addEventListener('push', event => {
    console.log('SW: Push event', event);

    const options = {
        body: event.data ? event.data.text() : 'New message from BuyOrBye',
        icon: '/static/images/icon-192x192.png',
        badge: '/static/images/badge-72x72.png',
        vibrate: [100, 50, 100],
        data: {
            url: '/'
        },
        actions: [
            {
                action: 'open',
                title: 'Open App',
                icon: '/static/images/action-open.png'
            },
            {
                action: 'close',
                title: 'Close',
                icon: '/static/images/action-close.png'
            }
        ]
    };

    event.waitUntil(
        self.registration.showNotification('BuyOrBye', options)
    );
});

// Notification click handling
self.addEventListener('notificationclick', event => {
    console.log('SW: Notification click', event);

    event.notification.close();

    if (event.action === 'open' || !event.action) {
        event.waitUntil(
            clients.openWindow(event.notification.data.url || '/')
        );
    }
});

// Helper functions
function getCacheStrategy(pathname) {
    for (const [key, config] of Object.entries(RUNTIME_CACHE)) {
        if (config.pattern.test(pathname)) {
            return config;
        }
    }
    return null;
}

async function handleRequest(request, strategy) {
    const cacheName = strategy.strategy === 'cacheFirst' ? STATIC_CACHE : CACHE_NAME;

    try {
        switch (strategy.strategy) {
            case 'cacheFirst':
                return await cacheFirst(request, cacheName, strategy);

            case 'networkFirst':
                return await networkFirst(request, cacheName, strategy);

            default:
                return await fetch(request);
        }
    } catch (error) {
        console.log('SW: Request failed', request.url, error);

        // Return offline fallback for navigation requests
        if (request.destination === 'document') {
            return getOfflineFallback();
        }

        // Return cached version if available
        const cachedResponse = await caches.match(request);
        if (cachedResponse) {
            return cachedResponse;
        }

        // Return generic offline response
        return new Response('Offline', { status: 503 });
    }
}

async function cacheFirst(request, cacheName, strategy) {
    const cachedResponse = await caches.match(request);

    if (cachedResponse) {
        // Update cache in background if not too fresh
        const cacheTime = new Date(cachedResponse.headers.get('sw-cache-time') || 0);
        const maxAge = strategy.maxAge * 1000;

        if (Date.now() - cacheTime.getTime() > maxAge) {
            updateCacheInBackground(request, cacheName, strategy);
        }

        return cachedResponse;
    }

    // Not in cache, fetch from network
    const networkResponse = await fetch(request);
    await updateCache(request, networkResponse.clone(), cacheName, strategy);
    return networkResponse;
}

async function networkFirst(request, cacheName, strategy) {
    try {
        const networkResponse = await fetch(request);
        await updateCache(request, networkResponse.clone(), cacheName, strategy);
        return networkResponse;
    } catch (error) {
        // Network failed, try cache
        const cachedResponse = await caches.match(request);
        if (cachedResponse) {
            return cachedResponse;
        }
        throw error;
    }
}

async function updateCache(request, response, cacheName, strategy) {
    if (!response.ok) return;

    const cache = await caches.open(cacheName);

    // Add timestamp header
    const responseToCache = new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: {
            ...response.headers,
            'sw-cache-time': new Date().toISOString()
        }
    });

    await cache.put(request, responseToCache);

    // Cleanup old entries if needed
    await cleanupCache(cache, strategy.maxEntries);
}

function updateCacheInBackground(request, cacheName, strategy) {
    // Don't await this - run in background
    fetch(request)
        .then(response => response.ok ? updateCache(request, response, cacheName, strategy) : null)
        .catch(() => {
            // Silently ignore background update failures
        });
}

async function cleanupCache(cache, maxEntries) {
    if (!maxEntries) return;

    const keys = await cache.keys();
    if (keys.length > maxEntries) {
        // Delete oldest entries (simple FIFO, could be improved with LRU)
        const deleteCount = keys.length - maxEntries;
        const deletePromises = keys.slice(0, deleteCount).map(key => cache.delete(key));
        await Promise.all(deletePromises);
    }
}

async function getOfflineFallback() {
    // Try to return cached home page
    const cachedHome = await caches.match('/');
    if (cachedHome) {
        return cachedHome;
    }

    // Return basic offline page
    return new Response(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>BuyOrBye - Offline</title>
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <style>
                body { font-family: system-ui, sans-serif; text-align: center; padding: 20px; background: #f8fafc; }
                .offline { max-width: 400px; margin: 100px auto; }
                .offline h1 { color: #3b82f6; margin-bottom: 20px; }
                .offline p { color: #64748b; line-height: 1.5; }
                .retry-btn { background: #3b82f6; color: white; border: none; padding: 12px 24px; border-radius: 8px; margin-top: 20px; cursor: pointer; }
                .retry-btn:hover { background: #2563eb; }
            </style>
        </head>
        <body>
            <div class="offline">
                <h1>📱 BuyOrBye</h1>
                <h2>You're Offline</h2>
                <p>It looks like you've lost your internet connection. Don't worry, your data is safe!</p>
                <p>Check your connection and try again.</p>
                <button class="retry-btn" onclick="location.reload()">Try Again</button>
            </div>
        </body>
        </html>
    `, {
        headers: { 'Content-Type': 'text/html' }
    });
}

async function syncChatMessages() {
    // Get pending messages from IndexedDB or localStorage
    // This would integrate with the offline handler in the main app
    console.log('SW: Syncing chat messages...');

    try {
        // In a real implementation, this would:
        // 1. Get pending messages from storage
        // 2. Send them to the server
        // 3. Remove successfully sent messages
        // 4. Keep failed messages for next sync

        // For now, just log that sync was attempted
        console.log('SW: Chat sync completed');
    } catch (error) {
        console.log('SW: Chat sync failed', error);
        throw error; // Re-throw to retry sync later
    }
}

console.log('SW: Service worker script loaded');