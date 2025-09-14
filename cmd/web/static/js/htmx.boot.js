/**
 * HTMX Boot Configuration for BuyOrBye
 * Handles CSRF protection, request logging, error handling, and global configuration
 */

(function() {
    'use strict';

    // Configuration
    const config = {
        debug: window.location.hostname === 'localhost',
        logRequests: true,
        retryAttempts: 3,
        retryDelay: 1000,
        timeoutMs: 30000
    };

    // Request ID generator for tracing
    function generateRequestId() {
        return Date.now().toString(36) + Math.random().toString(36).substr(2, 5);
    }

    // CSRF Token Management
    const csrf = {
        token: null,
        
        init() {
            // Get CSRF token from meta tag
            const meta = document.querySelector('meta[name="csrf-token"]');
            if (meta) {
                this.token = meta.getAttribute('content');
            }
            
            // Get token from data attribute on body
            if (!this.token && document.body.dataset.csrfToken) {
                this.token = document.body.dataset.csrfToken;
            }
            
            if (!this.token && config.debug) {
                console.warn('HTMX Boot: CSRF token not found');
            }
        },
        
        getToken() {
            return this.token;
        }
    };

    // Global Error Handler
    const errorHandler = {
        init() {
            // Handle HTMX response errors
            document.addEventListener('htmx:responseError', (event) => {
                const xhr = event.detail.xhr;
                const target = event.target;
                
                this.handleError({
                    type: 'response_error',
                    status: xhr.status,
                    statusText: xhr.statusText,
                    responseText: xhr.responseText,
                    target: target,
                    requestId: event.detail.requestConfig?.headers?.['X-Request-ID']
                });
            });

            // Handle HTMX send errors
            document.addEventListener('htmx:sendError', (event) => {
                this.handleError({
                    type: 'send_error',
                    error: event.detail.error,
                    target: event.target
                });
            });

            // Handle HTMX timeout
            document.addEventListener('htmx:timeout', (event) => {
                this.handleError({
                    type: 'timeout',
                    target: event.target
                });
            });

            // Handle validation errors
            document.addEventListener('htmx:validation:validate', (event) => {
                const result = this.validateForm(event.target);
                if (!result.isValid) {
                    event.detail.isValid = false;
                    this.showValidationErrors(event.target, result.errors);
                }
            });
        },

        handleError(errorInfo) {
            if (config.debug) {
                console.error('HTMX Error:', errorInfo);
            }

            // Handle specific error types
            switch (errorInfo.status) {
                case 401:
                    this.handleUnauthorized(errorInfo);
                    break;
                case 403:
                    this.handleForbidden(errorInfo);
                    break;
                case 422:
                    this.handleValidationError(errorInfo);
                    break;
                case 429:
                    this.handleRateLimit(errorInfo);
                    break;
                case 500:
                case 502:
                case 503:
                case 504:
                    this.handleServerError(errorInfo);
                    break;
                default:
                    this.handleGenericError(errorInfo);
            }
        },

        handleUnauthorized(errorInfo) {
            this.showToast('Session expired. Please log in again.', 'error');
            setTimeout(() => {
                window.location.href = '/auth/login?redirect=' + encodeURIComponent(window.location.pathname);
            }, 2000);
        },

        handleForbidden(errorInfo) {
            this.showToast('Access denied. You do not have permission for this action.', 'error');
        },

        handleValidationError(errorInfo) {
            try {
                const errors = JSON.parse(errorInfo.responseText);
                if (errors.errors && errorInfo.target) {
                    this.showValidationErrors(errorInfo.target.closest('form') || errorInfo.target, errors.errors);
                }
            } catch (e) {
                this.showToast('Validation failed. Please check your input.', 'error');
            }
        },

        handleRateLimit(errorInfo) {
            this.showToast('Too many requests. Please wait a moment before trying again.', 'warning');
        },

        handleServerError(errorInfo) {
            this.showToast('Server error occurred. Please try again later.', 'error');
            
            // Log server errors for debugging
            if (config.debug && errorInfo.requestId) {
                console.error(`Server Error [${errorInfo.requestId}]:`, {
                    status: errorInfo.status,
                    response: errorInfo.responseText
                });
            }
        },

        handleGenericError(errorInfo) {
            if (errorInfo.type === 'timeout') {
                this.showToast('Request timed out. Please try again.', 'warning');
            } else if (errorInfo.type === 'send_error') {
                this.showToast('Connection error. Please check your internet connection.', 'error');
            } else {
                this.showToast('An unexpected error occurred.', 'error');
            }
        },

        validateForm(form) {
            const errors = {};
            let isValid = true;

            // Clear previous errors
            form.querySelectorAll('.error-message').forEach(el => el.remove());
            form.querySelectorAll('.border-red-500').forEach(el => {
                el.classList.remove('border-red-500');
            });

            // Validate required fields
            form.querySelectorAll('[required]').forEach(input => {
                if (!input.value.trim()) {
                    errors[input.name] = 'This field is required';
                    isValid = false;
                }
            });

            // Validate email fields
            form.querySelectorAll('input[type="email"]').forEach(input => {
                if (input.value && !this.isValidEmail(input.value)) {
                    errors[input.name] = 'Please enter a valid email address';
                    isValid = false;
                }
            });

            // Validate number fields
            form.querySelectorAll('input[type="number"]').forEach(input => {
                if (input.value && isNaN(input.value)) {
                    errors[input.name] = 'Please enter a valid number';
                    isValid = false;
                }
            });

            return { isValid, errors };
        },

        showValidationErrors(form, errors) {
            Object.keys(errors).forEach(fieldName => {
                const field = form.querySelector(`[name="${fieldName}"]`);
                if (field) {
                    field.classList.add('border-red-500');
                    
                    const errorEl = document.createElement('div');
                    errorEl.className = 'error-message text-red-500 text-sm mt-1';
                    errorEl.textContent = errors[fieldName];
                    
                    field.parentNode.appendChild(errorEl);
                }
            });
        },

        isValidEmail(email) {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            return emailRegex.test(email);
        },

        showToast(message, type = 'info') {
            // Dispatch custom event for toast notification
            document.dispatchEvent(new CustomEvent('show-toast', {
                detail: { message, type }
            }));
        }
    };

    // Request Logger
    const logger = {
        init() {
            if (!config.logRequests) return;

            document.addEventListener('htmx:beforeRequest', (event) => {
                const requestId = generateRequestId();
                event.detail.requestConfig.headers['X-Request-ID'] = requestId;
                
                if (config.debug) {
                    console.group(`🚀 HTMX Request [${requestId}]`);
                    console.log('Method:', event.detail.requestConfig.verb);
                    console.log('URL:', event.detail.requestConfig.path);
                    console.log('Target:', event.target.tagName + (event.target.id ? `#${event.target.id}` : ''));
                    console.log('Headers:', event.detail.requestConfig.headers);
                    console.groupEnd();
                }
            });

            document.addEventListener('htmx:afterRequest', (event) => {
                if (config.debug) {
                    const requestId = event.detail.requestConfig?.headers?.['X-Request-ID'];
                    const duration = Date.now() - (event.detail.requestConfig?.startTime || Date.now());
                    
                    console.group(`✅ HTMX Response [${requestId}]`);
                    console.log('Status:', event.detail.xhr.status);
                    console.log('Duration:', `${duration}ms`);
                    console.log('Response Size:', event.detail.xhr.responseText.length, 'chars');
                    console.groupEnd();
                }
            });
        }
    };

    // Request Interceptor for custom headers and retry logic
    const interceptor = {
        init() {
            document.addEventListener('htmx:configRequest', (event) => {
                const config = event.detail;
                
                // Add CSRF token to all non-GET requests
                if (config.verb !== 'get' && csrf.getToken()) {
                    config.headers['X-CSRF-Token'] = csrf.getToken();
                }

                // Add common headers
                config.headers['X-Requested-With'] = 'XMLHttpRequest';
                config.headers['X-App-Version'] = window.APP_VERSION || '1.0.0';
                
                // Add timestamp for cache busting if needed
                if (config.verb === 'get' && event.target.dataset.bustCache === 'true') {
                    const separator = config.path.includes('?') ? '&' : '?';
                    config.path += `${separator}_t=${Date.now()}`;
                }

                // Store start time for duration calculation
                config.startTime = Date.now();
            });

            // Add retry logic for failed requests
            document.addEventListener('htmx:responseError', (event) => {
                const target = event.target;
                const retryCount = parseInt(target.dataset.retryCount || '0');
                
                // Retry on 5xx errors or network issues
                if ((event.detail.xhr.status >= 500 || event.detail.xhr.status === 0) && 
                    retryCount < config.retryAttempts) {
                    
                    target.dataset.retryCount = (retryCount + 1).toString();
                    
                    setTimeout(() => {
                        if (config.debug) {
                            console.log(`Retrying request (attempt ${retryCount + 1}/${config.retryAttempts})`);
                        }
                        htmx.trigger(target, event.detail.requestConfig.triggerSpec.trigger);
                    }, config.retryDelay * Math.pow(2, retryCount)); // Exponential backoff
                    
                    event.preventDefault();
                }
            });
        }
    };

    // Loading States Manager
    const loadingStates = {
        init() {
            // Add loading classes during requests
            document.addEventListener('htmx:beforeRequest', (event) => {
                const target = event.target;
                const indicator = target.dataset.indicator;
                
                // Add loading state to target
                target.classList.add('htmx-loading');
                target.setAttribute('aria-busy', 'true');
                
                // Show custom indicator if specified
                if (indicator) {
                    const indicatorEl = document.querySelector(indicator);
                    if (indicatorEl) {
                        indicatorEl.style.display = 'block';
                        indicatorEl.classList.add('htmx-indicator-show');
                    }
                }
            });

            document.addEventListener('htmx:afterRequest', (event) => {
                const target = event.target;
                const indicator = target.dataset.indicator;
                
                // Remove loading state
                target.classList.remove('htmx-loading');
                target.removeAttribute('aria-busy');
                
                // Hide custom indicator
                if (indicator) {
                    const indicatorEl = document.querySelector(indicator);
                    if (indicatorEl) {
                        indicatorEl.style.display = 'none';
                        indicatorEl.classList.remove('htmx-indicator-show');
                    }
                }
                
                // Reset retry count on success
                if (event.detail.successful) {
                    target.removeAttribute('data-retry-count');
                }
            });
        }
    };

    // Initialize HTMX Boot System
    function init() {
        // Wait for DOM to be ready
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', init);
            return;
        }

        if (config.debug) {
            console.log('🚀 Initializing HTMX Boot System');
        }

        // Initialize components
        csrf.init();
        errorHandler.init();
        logger.init();
        interceptor.init();
        loadingStates.init();

        // Configure HTMX defaults
        htmx.config.timeout = config.timeoutMs;
        htmx.config.scrollBehavior = 'smooth';
        htmx.config.defaultSwapStyle = 'innerHTML';
        htmx.config.defaultSettleDelay = 20;

        // Global HTMX extensions
        if (typeof htmx !== 'undefined') {
            // Add custom HTMX extension for BuyOrBye
            htmx.defineExtension('buyorbye', {
                onEvent: function(name, evt) {
                    if (name === 'htmx:afterSettle') {
                        // Re-initialize any Alpine components in new content
                        if (window.Alpine && evt.target) {
                            Alpine.initTree(evt.target);
                        }
                        
                        // Trigger custom settled event
                        evt.target.dispatchEvent(new CustomEvent('buyorbye:settled', {
                            bubbles: true,
                            detail: { originalEvent: evt }
                        }));
                    }
                }
            });
        }

        if (config.debug) {
            console.log('✅ HTMX Boot System initialized');
            
            // Add debug helpers to window
            window.htmxBoot = {
                config,
                csrf: csrf.getToken(),
                generateRequestId,
                version: '1.0.0'
            };
        }

        // Dispatch ready event
        document.dispatchEvent(new CustomEvent('htmx:boot:ready'));
    }

    // Auto-initialize
    init();

})();