// Alpine.js stores for BuyOrBye
document.addEventListener('alpine:init', () => {
    Alpine.store('modals', {
        openModals: new Set(),
        
        isOpen(id) {
            return this.openModals.has(id);
        },
        
        open(id) {
            this.openModals.add(id);
        },
        
        close(id) {
            this.openModals.delete(id);
        },
        
        confirm(id) {
            // Emit custom event for confirmation
            window.dispatchEvent(new CustomEvent('modal:confirmed', { detail: { modalId: id } }));
            this.close(id);
        },
        
        closeAll() {
            this.openModals.clear();
        }
    });

    Alpine.store('toast', {
        toasts: [],
        nextId: 1,
        
        show(message, type = 'info', duration = 5000) {
            const id = this.nextId++;
            const toast = { id, message, type, duration };
            this.toasts.push(toast);
            
            // Auto remove after duration
            setTimeout(() => {
                this.remove(id);
            }, duration);
            
            return id;
        },
        
        remove(id) {
            const index = this.toasts.findIndex(toast => toast.id === id);
            if (index > -1) {
                this.toasts.splice(index, 1);
            }
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
    });

    Alpine.store('sidebar', {
        open: false,
        
        toggle() {
            this.open = !this.open;
        },
        
        close() {
            this.open = false;
        }
    });
});

// HTMX custom events
document.body.addEventListener('htmx:afterRequest', function(evt) {
    // Handle response headers for toast messages
    const xhr = evt.detail.xhr;
    const toastMessage = xhr.getResponseHeader('X-Toast-Message');
    const toastType = xhr.getResponseHeader('X-Toast-Type') || 'info';
    
    if (toastMessage) {
        Alpine.store('toast').show(toastMessage, toastType);
    }
});

document.body.addEventListener('htmx:responseError', function(evt) {
    Alpine.store('toast').error('An error occurred. Please try again.');
});

document.body.addEventListener('htmx:sendError', function(evt) {
    Alpine.store('toast').error('Network error. Please check your connection.');
});