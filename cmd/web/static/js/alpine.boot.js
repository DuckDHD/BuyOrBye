/**
 * Alpine.js Boot Configuration for BuyOrBye
 * Provides global stores, magic helpers, and shared component functionality
 */

(function() {
    'use strict';

    // Configuration
    const config = {
        debug: window.location.hostname === 'localhost',
        theme: {
            default: 'system',
            storage: 'buyorbye_theme'
        },
        currency: {
            locale: 'en-US',
            currency: 'USD'
        }
    };

    // Global UI Store
    function createUIStore() {
        return {
            // Sidebar state
            sidebar: {
                isOpen: false,
                isPinned: localStorage.getItem('buyorbye_sidebar_pinned') === 'true',
                
                toggle() {
                    this.isOpen = !this.isOpen;
                },
                
                open() {
                    this.isOpen = true;
                },
                
                close() {
                    this.isOpen = false;
                },
                
                pin() {
                    this.isPinned = true;
                    localStorage.setItem('buyorbye_sidebar_pinned', 'true');
                },
                
                unpin() {
                    this.isPinned = false;
                    localStorage.setItem('buyorbye_sidebar_pinned', 'false');
                }
            },

            // Theme management
            theme: {
                current: localStorage.getItem(config.theme.storage) || config.theme.default,
                
                set(theme) {
                    this.current = theme;
                    localStorage.setItem(config.theme.storage, theme);
                    this.apply();
                },
                
                toggle() {
                    const newTheme = this.current === 'dark' ? 'light' : 'dark';
                    this.set(newTheme);
                },
                
                apply() {
                    const root = document.documentElement;
                    
                    if (this.current === 'dark' || 
                        (this.current === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
                        root.classList.add('dark');
                    } else {
                        root.classList.remove('dark');
                    }
                },
                
                init() {
                    this.apply();
                    
                    // Listen for system theme changes
                    if (this.current === 'system') {
                        window.matchMedia('(prefers-color-scheme: dark)')
                            .addEventListener('change', () => this.apply());
                    }
                }
            },

            // Loading states
            loading: {
                global: false,
                states: new Set(),
                
                start(key = 'global') {
                    this.states.add(key);
                    this.global = this.states.size > 0;
                },
                
                stop(key = 'global') {
                    this.states.delete(key);
                    this.global = this.states.size > 0;
                },
                
                isLoading(key = 'global') {
                    return this.states.has(key);
                }
            },

            // Toast notifications
            toasts: {
                items: [],
                maxToasts: 5,
                defaultDuration: 5000,
                
                show(message, type = 'info', duration = this.defaultDuration) {
                    const toast = {
                        id: Date.now() + Math.random(),
                        message,
                        type,
                        duration,
                        timestamp: new Date()
                    };
                    
                    this.items.unshift(toast);
                    
                    // Remove excess toasts
                    if (this.items.length > this.maxToasts) {
                        this.items = this.items.slice(0, this.maxToasts);
                    }
                    
                    // Auto remove toast
                    if (duration > 0) {
                        setTimeout(() => this.remove(toast.id), duration);
                    }
                    
                    return toast.id;
                },
                
                remove(id) {
                    const index = this.items.findIndex(toast => toast.id === id);
                    if (index > -1) {
                        this.items.splice(index, 1);
                    }
                },
                
                clear() {
                    this.items = [];
                },
                
                success(message, duration) {
                    return this.show(message, 'success', duration);
                },
                
                error(message, duration) {
                    return this.show(message, 'error', duration);
                },
                
                warning(message, duration) {
                    return this.show(message, 'warning', duration);
                },
                
                info(message, duration) {
                    return this.show(message, 'info', duration);
                }
            }
        };
    }

    // Modal Management Store
    function createModalStore() {
        return {
            stack: [],
            
            isOpen(id = null) {
                if (id) {
                    return this.stack.includes(id);
                }
                return this.stack.length > 0;
            },
            
            open(id) {
                if (!this.stack.includes(id)) {
                    this.stack.push(id);
                    document.body.style.overflow = 'hidden';
                }
            },
            
            close(id = null) {
                if (id) {
                    const index = this.stack.indexOf(id);
                    if (index > -1) {
                        this.stack.splice(index, 1);
                    }
                } else {
                    // Close topmost modal
                    this.stack.pop();
                }
                
                if (this.stack.length === 0) {
                    document.body.style.overflow = '';
                }
            },
            
            closeAll() {
                this.stack = [];
                document.body.style.overflow = '';
            },
            
            confirm(id) {
                document.dispatchEvent(new CustomEvent('modal:confirm', {
                    detail: { id }
                }));
                this.close(id);
            }
        };
    }

    // Form Validation Helpers
    function createValidationHelpers() {
        return {
            rules: {
                required: (value) => {
                    return value !== null && value !== undefined && String(value).trim() !== '';
                },
                
                email: (value) => {
                    if (!value) return true; // Optional field
                    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                    return emailRegex.test(value);
                },
                
                min: (value, min) => {
                    if (!value) return true; // Optional field
                    return String(value).length >= min;
                },
                
                max: (value, max) => {
                    if (!value) return true; // Optional field
                    return String(value).length <= max;
                },
                
                number: (value) => {
                    if (!value) return true; // Optional field
                    return !isNaN(value) && isFinite(value);
                },
                
                currency: (value) => {
                    if (!value) return true; // Optional field
                    const cleanValue = String(value).replace(/[$,]/g, '');
                    return !isNaN(cleanValue) && parseFloat(cleanValue) >= 0;
                },
                
                url: (value) => {
                    if (!value) return true; // Optional field
                    try {
                        new URL(value);
                        return true;
                    } catch {
                        return false;
                    }
                }
            },
            
            validate(value, rules) {
                const errors = [];
                
                if (typeof rules === 'string') {
                    rules = rules.split('|');
                }
                
                for (const rule of rules) {
                    const [ruleName, ...params] = rule.split(':');
                    const validator = this.rules[ruleName];
                    
                    if (validator && !validator(value, ...params)) {
                        errors.push(this.getErrorMessage(ruleName, params));
                    }
                }
                
                return {
                    isValid: errors.length === 0,
                    errors
                };
            },
            
            getErrorMessage(rule, params = []) {
                const messages = {
                    required: 'This field is required',
                    email: 'Please enter a valid email address',
                    min: `Must be at least ${params[0]} characters`,
                    max: `Must not exceed ${params[0]} characters`,
                    number: 'Please enter a valid number',
                    currency: 'Please enter a valid amount',
                    url: 'Please enter a valid URL'
                };
                
                return messages[rule] || 'Invalid value';
            }
        };
    }

    // Utility Functions
    function createUtilities() {
        return {
            // Currency formatting
            formatCurrency(amount, options = {}) {
                const defaults = {
                    style: 'currency',
                    currency: config.currency.currency,
                    minimumFractionDigits: 2,
                    maximumFractionDigits: 2
                };
                
                const formatOptions = { ...defaults, ...options };
                
                try {
                    return new Intl.NumberFormat(config.currency.locale, formatOptions)
                        .format(parseFloat(amount) || 0);
                } catch (error) {
                    return `$${parseFloat(amount || 0).toFixed(2)}`;
                }
            },

            // Date formatting
            formatDate(date, options = {}) {
                const defaults = {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric'
                };
                
                const formatOptions = { ...defaults, ...options };
                
                try {
                    const dateObj = typeof date === 'string' ? new Date(date) : date;
                    return new Intl.DateTimeFormat(config.currency.locale, formatOptions)
                        .format(dateObj);
                } catch (error) {
                    return 'Invalid Date';
                }
            },

            // Relative time formatting
            formatRelativeTime(date) {
                const now = new Date();
                const targetDate = typeof date === 'string' ? new Date(date) : date;
                const diffMs = now - targetDate;
                const diffSec = Math.floor(diffMs / 1000);
                const diffMin = Math.floor(diffSec / 60);
                const diffHour = Math.floor(diffMin / 60);
                const diffDay = Math.floor(diffHour / 24);

                if (diffSec < 60) return 'just now';
                if (diffMin < 60) return `${diffMin}m ago`;
                if (diffHour < 24) return `${diffHour}h ago`;
                if (diffDay < 7) return `${diffDay}d ago`;
                
                return this.formatDate(targetDate);
            },

            // Number formatting
            formatNumber(number, options = {}) {
                const defaults = {
                    minimumFractionDigits: 0,
                    maximumFractionDigits: 2
                };
                
                const formatOptions = { ...defaults, ...options };
                
                try {
                    return new Intl.NumberFormat(config.currency.locale, formatOptions)
                        .format(parseFloat(number) || 0);
                } catch (error) {
                    return String(number || 0);
                }
            },

            // Percentage formatting
            formatPercentage(value, decimals = 1) {
                return `${parseFloat(value || 0).toFixed(decimals)}%`;
            },

            // Debounce function
            debounce(func, wait) {
                let timeout;
                return function executedFunction(...args) {
                    const later = () => {
                        clearTimeout(timeout);
                        func(...args);
                    };
                    clearTimeout(timeout);
                    timeout = setTimeout(later, wait);
                };
            },

            // Throttle function
            throttle(func, limit) {
                let inThrottle;
                return function() {
                    const args = arguments;
                    const context = this;
                    if (!inThrottle) {
                        func.apply(context, args);
                        inThrottle = true;
                        setTimeout(() => inThrottle = false, limit);
                    }
                };
            },

            // Copy to clipboard
            async copyToClipboard(text) {
                try {
                    await navigator.clipboard.writeText(text);
                    Alpine.store('ui').toasts.success('Copied to clipboard');
                    return true;
                } catch (error) {
                    console.error('Copy failed:', error);
                    Alpine.store('ui').toasts.error('Failed to copy');
                    return false;
                }
            },

            // Generate random ID
            generateId(prefix = 'id') {
                return `${prefix}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
            }
        };
    }

    // Decision-specific utilities
    function createDecisionHelpers() {
        return {
            getDecisionColor(decision) {
                const colors = {
                    'BUY': 'text-green-600 dark:text-green-400',
                    'WAIT': 'text-yellow-600 dark:text-yellow-400',
                    'BYE': 'text-red-600 dark:text-red-400'
                };
                return colors[decision] || 'text-gray-600 dark:text-gray-400';
            },

            getDecisionBadgeColor(decision) {
                const colors = {
                    'BUY': 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
                    'WAIT': 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
                    'BYE': 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                };
                return colors[decision] || 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400';
            },

            getConfidenceColor(confidence) {
                if (confidence >= 80) return 'text-green-600 dark:text-green-400';
                if (confidence >= 60) return 'text-yellow-600 dark:text-yellow-400';
                return 'text-red-600 dark:text-red-400';
            },

            formatConfidence(confidence) {
                return `${Math.round(confidence)}%`;
            }
        };
    }

    // Initialize Alpine.js Boot System
    function init() {
        // Wait for Alpine to be available
        if (typeof Alpine === 'undefined') {
            setTimeout(init, 100);
            return;
        }

        if (config.debug) {
            console.log('🏔️ Initializing Alpine Boot System');
        }

        // Create and register stores
        Alpine.store('ui', createUIStore());
        Alpine.store('modals', createModalStore());

        // Create utilities
        const utils = createUtilities();
        const validation = createValidationHelpers();
        const decisions = createDecisionHelpers();

        // Register magic helpers
        Alpine.magic('currency', () => utils.formatCurrency);
        Alpine.magic('date', () => utils.formatDate);
        Alpine.magic('relativeTime', () => utils.formatRelativeTime);
        Alpine.magic('number', () => utils.formatNumber);
        Alpine.magic('percentage', () => utils.formatPercentage);
        Alpine.magic('debounce', () => utils.debounce);
        Alpine.magic('throttle', () => utils.throttle);
        Alpine.magic('copy', () => utils.copyToClipboard);
        Alpine.magic('generateId', () => utils.generateId);
        Alpine.magic('validate', () => validation.validate.bind(validation));
        Alpine.magic('decision', () => decisions);

        // Global event listeners
        document.addEventListener('show-toast', (event) => {
            const { message, type } = event.detail;
            Alpine.store('ui').toasts.show(message, type);
        });

        document.addEventListener('keydown', (event) => {
            // Close modal on Escape
            if (event.key === 'Escape' && Alpine.store('modals').isOpen()) {
                Alpine.store('modals').close();
            }
        });

        // Initialize theme
        Alpine.store('ui').theme.init();

        // Initialize on Alpine start
        Alpine.start();

        if (config.debug) {
            console.log('✅ Alpine Boot System initialized');
            
            // Add debug helpers to window
            window.alpineBoot = {
                config,
                utils,
                validation,
                decisions,
                version: '1.0.0'
            };
        }

        // Dispatch ready event
        document.dispatchEvent(new CustomEvent('alpine:boot:ready'));
    }

    // Auto-initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();