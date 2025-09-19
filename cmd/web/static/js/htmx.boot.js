// HTMX Boot Configuration for BuyOrBye
// Implements CSRF protection and other HTMX configuration as per CLAUDE.md

document.body.addEventListener('htmx:configRequest', (e) => {
    // CSRF Token injection
    const token = document.querySelector('meta[name="csrf-token"]')?.content
    if (token) e.detail.headers['X-CSRF-Token'] = token

    // Add request ID for tracing
    e.detail.headers['X-Request-ID'] = crypto.randomUUID()
})

// Global error handling with better styling
document.body.addEventListener('htmx:responseError', (e) => {
    const target = e.detail.target
    target.innerHTML = `
        <div class="rounded-md bg-red-50 dark:bg-red-900/20 p-4 border border-red-200 dark:border-red-800">
            <div class="flex">
                <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd"/>
                    </svg>
                </div>
                <div class="ml-3">
                    <h3 class="text-sm font-medium text-red-800 dark:text-red-200">
                        Something went wrong
                    </h3>
                    <div class="mt-2 text-sm text-red-700 dark:text-red-300">
                        Please try again. If the problem persists, contact support.
                    </div>
                </div>
            </div>
        </div>
    `
})

// Loading states - improved to handle multiple indicators
document.body.addEventListener('htmx:beforeRequest', (e) => {
    // Show global loading indicator
    const globalIndicator = document.querySelector('#global-loading')
    if (globalIndicator) globalIndicator.classList.remove('hidden')

    // Show specific form loading indicators
    const form = e.detail.elt.closest('form')
    if (form) {
        const indicators = form.querySelectorAll('.htmx-indicator')
        indicators.forEach(indicator => indicator.classList.remove('htmx-indicator'))
    }
})

document.body.addEventListener('htmx:afterRequest', (e) => {
    // Hide global loading indicator
    const globalIndicator = document.querySelector('#global-loading')
    if (globalIndicator) globalIndicator.classList.add('hidden')

    // Hide specific form loading indicators
    const form = e.detail.elt.closest('form')
    if (form) {
        const indicators = form.querySelectorAll('.htmx-indicator-show')
        indicators.forEach(indicator => indicator.classList.add('htmx-indicator'))
    }
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
    target.innerHTML = `
        <div class="rounded-md bg-yellow-50 dark:bg-yellow-900/20 p-4 border border-yellow-200 dark:border-yellow-800">
            <div class="flex">
                <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd"/>
                    </svg>
                </div>
                <div class="ml-3">
                    <h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                        Network error
                    </h3>
                    <div class="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
                        Please check your connection and try again.
                    </div>
                </div>
            </div>
        </div>
    `
})

// Handle timeout errors
document.body.addEventListener('htmx:timeout', (e) => {
    const target = e.detail.target
    target.innerHTML = `
        <div class="rounded-md bg-yellow-50 dark:bg-yellow-900/20 p-4 border border-yellow-200 dark:border-yellow-800">
            <div class="flex">
                <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm-1-13a1 1 0 112 0v4a1 1 0 11-2 0V5zm1 8a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"/>
                    </svg>
                </div>
                <div class="ml-3">
                    <h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                        Request timed out
                    </h3>
                    <div class="mt-2 text-sm text-yellow-700 dark:text-yellow-300">
                        The server took too long to respond. Please try again.
                    </div>
                </div>
            </div>
        </div>
    `
})

// Smooth transitions for content swaps
document.body.addEventListener('htmx:beforeSwap', (e) => {
    // Add fade-out class for smooth transition
    if (e.detail.target) {
        e.detail.target.style.opacity = '0.5'
        e.detail.target.style.transition = 'opacity 0.2s ease-in-out'
    }
})

document.body.addEventListener('htmx:afterSwap', (e) => {
    // Restore opacity with fade-in effect
    if (e.detail.target) {
        setTimeout(() => {
            e.detail.target.style.opacity = '1'
        }, 10) // Small delay to ensure DOM is updated
    }
})

// Add loading state to form elements
document.body.addEventListener('htmx:beforeRequest', (e) => {
    const form = e.detail.elt.closest('form')
    if (form) {
        // Disable form inputs during submission
        const inputs = form.querySelectorAll('input, button, select, textarea')
        inputs.forEach(input => {
            if (!input.disabled) {
                input.setAttribute('data-was-enabled', 'true')
                input.disabled = true
            }
        })

        // Add visual loading state to submit button
        const submitBtn = form.querySelector('button[type="submit"]')
        if (submitBtn) {
            submitBtn.classList.add('opacity-75', 'cursor-not-allowed')
        }
    }
})

document.body.addEventListener('htmx:afterRequest', (e) => {
    const form = e.detail.elt.closest('form')
    if (form) {
        // Re-enable form inputs
        const inputs = form.querySelectorAll('input[data-was-enabled], button[data-was-enabled], select[data-was-enabled], textarea[data-was-enabled]')
        inputs.forEach(input => {
            input.disabled = false
            input.removeAttribute('data-was-enabled')
        })

        // Remove loading state from submit button
        const submitBtn = form.querySelector('button[type="submit"]')
        if (submitBtn) {
            submitBtn.classList.remove('opacity-75', 'cursor-not-allowed')
        }

        // Focus first input with error if validation failed
        if (e.detail.xhr.status >= 400 && e.detail.xhr.status < 500) {
            setTimeout(() => {
                const firstErrorInput = form.querySelector('input.border-red-500, .border-red-500 input')
                if (firstErrorInput) {
                    firstErrorInput.focus()
                }
            }, 100)
        }
    }
})

// Form state preservation
document.body.addEventListener('htmx:beforeSwap', (e) => {
    // If we're swapping form content, preserve input values with Alpine state
    if (e.detail.target.closest('form')) {
        const form = e.detail.target.closest('form')
        const inputs = form.querySelectorAll('input[x-model], select[x-model], textarea[x-model]')

        // Store current input values before swap
        inputs.forEach(input => {
            if (input.type !== 'hidden') {
                input.setAttribute('data-preserved-value', input.value)
            }
        })
    }
})

document.body.addEventListener('htmx:afterSwap', (e) => {
    // Restore preserved values after swap
    if (e.detail.target.closest('form')) {
        const form = e.detail.target.closest('form')
        const inputs = form.querySelectorAll('input[data-preserved-value]')

        inputs.forEach(input => {
            const preservedValue = input.getAttribute('data-preserved-value')
            if (preservedValue && input.value !== preservedValue) {
                input.value = preservedValue
                // Trigger input event to update Alpine state
                input.dispatchEvent(new Event('input', { bubbles: true }))
            }
            input.removeAttribute('data-preserved-value')
        })
    }
})

// Clear form errors on successful submission
document.body.addEventListener('htmx:afterRequest', (e) => {
    if (e.detail.xhr.status >= 200 && e.detail.xhr.status < 300) {
        // Clear error states on success
        const errorContainers = document.querySelectorAll('[id*="-messages"], [id*="-errors"]')
        errorContainers.forEach(container => {
            if (container.innerHTML.includes('text-red-')) {
                container.innerHTML = ''
            }
        })

        // Show success message if provided in headers
        const successMessage = e.detail.xhr.getResponseHeader('X-Success-Message')
        if (successMessage && typeof Alpine !== 'undefined' && Alpine.store('toast')) {
            Alpine.store('toast').success(successMessage)
        }
    }
})

// Prevent form submission when Alpine validation fails
document.body.addEventListener('htmx:beforeRequest', (e) => {
    const form = e.detail.elt.closest('form')
    if (form && form.hasAttribute('x-data')) {
        // Get Alpine component data
        const alpineData = Alpine.$data(form)

        // If there's a canSubmit property, check it
        if (alpineData && typeof alpineData.canSubmit === 'boolean' && !alpineData.canSubmit) {
            e.preventDefault()

            // Show validation error toast
            if (typeof Alpine !== 'undefined' && Alpine.store('toast')) {
                Alpine.store('toast').error('Please fix the errors before submitting')
            }
        }
    }
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