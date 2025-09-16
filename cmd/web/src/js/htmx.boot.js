// HTMX Boot Configuration for BuyOrBye
// Implements CSRF protection and other HTMX configuration as per CLAUDE.md

document.body.addEventListener('htmx:configRequest', (e) => {
    // CSRF Token injection
    const token = document.querySelector('meta[name="csrf-token"]')?.content
    if (token) e.detail.headers['X-CSRF-Token'] = token

    // Add request ID for tracing
    e.detail.headers['X-Request-ID'] = crypto.randomUUID()
})

// Global error handling
document.body.addEventListener('htmx:responseError', (e) => {
    const target = e.detail.target
    target.innerHTML = '<div class="alert alert-error">Something went wrong. Please try again.</div>'
})

// Loading states
document.body.addEventListener('htmx:beforeRequest', (e) => {
    const indicator = document.querySelector('#loading-indicator')
    if (indicator) indicator.classList.remove('htmx-indicator')
})

document.body.addEventListener('htmx:afterRequest', (e) => {
    const indicator = document.querySelector('#loading-indicator')
    if (indicator) indicator.classList.add('htmx-indicator')
})

// Handle CSRF failures specifically
document.body.addEventListener('htmx:responseError', (e) => {
    if (e.detail.xhr.status === 403) {
        // CSRF token validation failed - redirect to login
        window.location.href = '/auth/login?error=session_expired'
    }
})

// Handle authentication failures
document.body.addEventListener('htmx:responseError', (e) => {
    if (e.detail.xhr.status === 401) {
        // Unauthorized - redirect to login
        window.location.href = '/auth/login?error=unauthorized'
    }
})

// Handle network errors gracefully
document.body.addEventListener('htmx:sendError', (e) => {
    const target = e.detail.target
    target.innerHTML = '<div class="alert alert-error">Network error. Please check your connection and try again.</div>'
})

// Handle timeout errors
document.body.addEventListener('htmx:timeout', (e) => {
    const target = e.detail.target
    target.innerHTML = '<div class="alert alert-error">Request timed out. Please try again.</div>'
})

// Log HTMX events for debugging (only in development)
if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    document.body.addEventListener('htmx:beforeRequest', (e) => {
        console.log('HTMX Request:', {
            method: e.detail.requestConfig.verb,
            url: e.detail.requestConfig.path,
            headers: e.detail.requestConfig.headers,
            target: e.detail.target
        })
    })

    document.body.addEventListener('htmx:afterRequest', (e) => {
        console.log('HTMX Response:', {
            status: e.detail.xhr.status,
            url: e.detail.xhr.responseURL,
            target: e.detail.target
        })
    })
}