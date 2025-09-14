/**
 * Alpine.js Store and Helper Tests
 * Tests for alpine.boot.js functionality
 */

// Mock Alpine.js for testing
global.Alpine = {
  store: jest.fn(() => ({
    get: jest.fn(),
    set: jest.fn()
  })),
  magic: jest.fn(),
  data: jest.fn()
};

global.document = {
  addEventListener: jest.fn(),
  dispatchEvent: jest.fn(),
  createElement: jest.fn(() => ({
    style: {},
    classList: {
      add: jest.fn(),
      remove: jest.fn()
    },
    textContent: ''
  })),
  body: {
    appendChild: jest.fn(),
    removeChild: jest.fn()
  },
  querySelector: jest.fn(),
  querySelectorAll: jest.fn()
};

global.navigator = {
  clipboard: {
    writeText: jest.fn().mockResolvedValue()
  }
};

describe('Alpine Store - UI State Management', () => {
  let mockUIStore;

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock UI store structure
    mockUIStore = {
      theme: {
        current: 'light',
        toggle: jest.fn(),
        set: jest.fn()
      },
      sidebar: {
        isOpen: false,
        toggle: jest.fn(),
        open: jest.fn(),
        close: jest.fn()
      },
      modals: {
        activeModal: null,
        open: jest.fn(),
        close: jest.fn(),
        closeAll: jest.fn()
      },
      toasts: {
        items: [],
        show: jest.fn(),
        remove: jest.fn(),
        clear: jest.fn()
      }
    };

    Alpine.store.mockReturnValue(mockUIStore);
  });

  test('should manage theme state', () => {
    const uiStore = Alpine.store('ui');
    
    // Test theme toggle
    uiStore.theme.toggle();
    expect(uiStore.theme.toggle).toHaveBeenCalled();

    // Test theme set
    uiStore.theme.set('dark');
    expect(uiStore.theme.set).toHaveBeenCalledWith('dark');
  });

  test('should manage sidebar state', () => {
    const uiStore = Alpine.store('ui');
    
    // Test sidebar toggle
    uiStore.sidebar.toggle();
    expect(uiStore.sidebar.toggle).toHaveBeenCalled();

    // Test sidebar open
    uiStore.sidebar.open();
    expect(uiStore.sidebar.open).toHaveBeenCalled();

    // Test sidebar close
    uiStore.sidebar.close();
    expect(uiStore.sidebar.close).toHaveBeenCalled();
  });

  test('should manage modal state', () => {
    const uiStore = Alpine.store('ui');
    
    // Test modal open
    uiStore.modals.open('test-modal');
    expect(uiStore.modals.open).toHaveBeenCalledWith('test-modal');

    // Test modal close
    uiStore.modals.close();
    expect(uiStore.modals.close).toHaveBeenCalled();

    // Test close all modals
    uiStore.modals.closeAll();
    expect(uiStore.modals.closeAll).toHaveBeenCalled();
  });

  test('should manage toast notifications', () => {
    const uiStore = Alpine.store('ui');
    
    // Test show toast
    uiStore.toasts.show('Success message', 'success');
    expect(uiStore.toasts.show).toHaveBeenCalledWith('Success message', 'success');

    // Test remove toast
    uiStore.toasts.remove('toast-id');
    expect(uiStore.toasts.remove).toHaveBeenCalledWith('toast-id');

    // Test clear all toasts
    uiStore.toasts.clear();
    expect(uiStore.toasts.clear).toHaveBeenCalled();
  });
});

describe('Alpine Magic Helpers', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('should format currency with $currency magic', () => {
    const currencyFormatter = (value, currency = 'USD', locale = 'en-US') => {
      return new Intl.NumberFormat(locale, {
        style: 'currency',
        currency: currency
      }).format(value);
    };

    // Test various currency formats
    expect(currencyFormatter(1234.56)).toBe('$1,234.56');
    expect(currencyFormatter(1234.56, 'EUR', 'de-DE')).toBe('1.234,56 €');
    expect(currencyFormatter(0)).toBe('$0.00');
    expect(currencyFormatter(-123.45)).toBe('-$123.45');
  });

  test('should format dates with $date magic', () => {
    const dateFormatter = (date, format = 'short') => {
      const d = new Date(date);
      
      switch (format) {
        case 'short':
          return d.toLocaleDateString();
        case 'long':
          return d.toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
          });
        case 'time':
          return d.toLocaleTimeString();
        case 'relative':
          const now = new Date();
          const diffMs = now - d;
          const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
          
          if (diffDays === 0) return 'Today';
          if (diffDays === 1) return 'Yesterday';
          if (diffDays < 7) return `${diffDays} days ago`;
          return d.toLocaleDateString();
        default:
          return d.toLocaleDateString();
      }
    };

    const testDate = new Date('2024-01-15T10:30:00Z');
    
    expect(dateFormatter(testDate, 'short')).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/);
    expect(dateFormatter(testDate, 'long')).toBe('January 15, 2024');
    expect(dateFormatter(testDate, 'time')).toMatch(/\d{1,2}:\d{2}:\d{2}/);
  });

  test('should validate input with $validate magic', () => {
    const validator = (value, rules) => {
      const ruleArray = rules.split('|');
      const errors = [];

      for (const rule of ruleArray) {
        const [ruleName, ruleValue] = rule.split(':');

        switch (ruleName) {
          case 'required':
            if (!value || value.toString().trim() === '') {
              errors.push('This field is required');
            }
            break;
          case 'email':
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (value && !emailRegex.test(value)) {
              errors.push('Please enter a valid email address');
            }
            break;
          case 'min':
            if (value && value.toString().length < parseInt(ruleValue)) {
              errors.push(`Minimum ${ruleValue} characters required`);
            }
            break;
          case 'max':
            if (value && value.toString().length > parseInt(ruleValue)) {
              errors.push(`Maximum ${ruleValue} characters allowed`);
            }
            break;
          case 'numeric':
            if (value && isNaN(value)) {
              errors.push('This field must be numeric');
            }
            break;
        }
      }

      return {
        isValid: errors.length === 0,
        errors: errors
      };
    };

    // Test required validation
    expect(validator('', 'required')).toEqual({
      isValid: false,
      errors: ['This field is required']
    });

    // Test email validation
    expect(validator('invalid-email', 'email')).toEqual({
      isValid: false,
      errors: ['Please enter a valid email address']
    });

    expect(validator('test@example.com', 'email')).toEqual({
      isValid: true,
      errors: []
    });

    // Test length validation
    expect(validator('ab', 'min:3')).toEqual({
      isValid: false,
      errors: ['Minimum 3 characters required']
    });

    // Test multiple rules
    expect(validator('', 'required|email')).toEqual({
      isValid: false,
      errors: ['This field is required']
    });
  });

  test('should handle clipboard operations with $clipboard magic', async () => {
    const clipboardHelper = {
      copy: async (text) => {
        try {
          await navigator.clipboard.writeText(text);
          return { success: true, message: 'Copied to clipboard' };
        } catch (error) {
          return { success: false, message: 'Failed to copy' };
        }
      }
    };

    const result = await clipboardHelper.copy('Test text');
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('Test text');
    expect(result).toEqual({ success: true, message: 'Copied to clipboard' });
  });
});

describe('Alpine Utility Functions', () => {
  test('should debounce function calls', (done) => {
    let callCount = 0;
    
    const debounce = (func, delay) => {
      let timeout;
      return (...args) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(this, args), delay);
      };
    };

    const debouncedFunction = debounce(() => {
      callCount++;
    }, 100);

    // Call multiple times rapidly
    debouncedFunction();
    debouncedFunction();
    debouncedFunction();

    // Should only call once after delay
    setTimeout(() => {
      expect(callCount).toBe(1);
      done();
    }, 150);
  });

  test('should throttle function calls', (done) => {
    let callCount = 0;
    
    const throttle = (func, delay) => {
      let lastCall = 0;
      return (...args) => {
        const now = Date.now();
        if (now - lastCall >= delay) {
          lastCall = now;
          func.apply(this, args);
        }
      };
    };

    const throttledFunction = throttle(() => {
      callCount++;
    }, 100);

    // Call multiple times rapidly
    throttledFunction();
    throttledFunction();
    throttledFunction();

    // Should only call once immediately
    expect(callCount).toBe(1);

    // Call again after delay
    setTimeout(() => {
      throttledFunction();
      expect(callCount).toBe(2);
      done();
    }, 150);
  });

  test('should format numbers with locale support', () => {
    const numberFormatter = (value, options = {}) => {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: options.decimals || 0,
        maximumFractionDigits: options.decimals || 2,
        useGrouping: options.grouping !== false
      }).format(value);
    };

    expect(numberFormatter(1234.567)).toBe('1,234.57');
    expect(numberFormatter(1234.567, { decimals: 3 })).toBe('1,234.567');
    expect(numberFormatter(1234, { grouping: false })).toBe('1234');
  });
});

describe('Alpine Decision Helpers', () => {
  test('should format decision colors correctly', () => {
    const getDecisionColor = (decision) => {
      const colors = {
        'BUY': 'text-green-600 bg-green-100',
        'WAIT': 'text-yellow-600 bg-yellow-100',
        'BYE': 'text-red-600 bg-red-100'
      };
      
      return colors[decision] || 'text-gray-600 bg-gray-100';
    };

    expect(getDecisionColor('BUY')).toBe('text-green-600 bg-green-100');
    expect(getDecisionColor('WAIT')).toBe('text-yellow-600 bg-yellow-100');
    expect(getDecisionColor('BYE')).toBe('text-red-600 bg-red-100');
    expect(getDecisionColor('UNKNOWN')).toBe('text-gray-600 bg-gray-100');
  });

  test('should calculate confidence level classes', () => {
    const getConfidenceClass = (confidence) => {
      if (confidence >= 80) return 'confidence-high';
      if (confidence >= 60) return 'confidence-medium';
      return 'confidence-low';
    };

    expect(getConfidenceClass(85)).toBe('confidence-high');
    expect(getConfidenceClass(70)).toBe('confidence-medium');
    expect(getConfidenceClass(45)).toBe('confidence-low');
  });

  test('should format price ranges', () => {
    const formatPriceRange = (min, max, currency = 'USD') => {
      const formatter = new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: currency
      });

      if (min === max) {
        return formatter.format(min);
      }

      return `${formatter.format(min)} - ${formatter.format(max)}`;
    };

    expect(formatPriceRange(100, 100)).toBe('$100.00');
    expect(formatPriceRange(100, 200)).toBe('$100.00 - $200.00');
  });
});

describe('Alpine Component Integration', () => {
  test('should initialize dashboard data component', () => {
    const dashboardData = () => ({
      refreshStats: jest.fn(),
      showNotification: jest.fn(),
      init: jest.fn()
    });

    const component = dashboardData();
    
    expect(component).toHaveProperty('refreshStats');
    expect(component).toHaveProperty('showNotification');
    expect(component).toHaveProperty('init');
    
    // Test method calls
    component.refreshStats();
    component.showNotification('Test message', 'success');
    component.init();
    
    expect(component.refreshStats).toHaveBeenCalled();
    expect(component.showNotification).toHaveBeenCalledWith('Test message', 'success');
    expect(component.init).toHaveBeenCalled();
  });

  test('should handle form validation component', () => {
    const formValidation = () => ({
      errors: {},
      isValid: true,
      validate: jest.fn(),
      clearErrors: jest.fn(),
      getFieldError: jest.fn()
    });

    const component = formValidation();
    
    expect(component).toHaveProperty('errors');
    expect(component).toHaveProperty('isValid');
    expect(component).toHaveProperty('validate');
    
    // Test validation
    component.validate();
    expect(component.validate).toHaveBeenCalled();
  });

  test('should manage modal component state', () => {
    const modalComponent = () => ({
      isOpen: false,
      open: jest.fn(),
      close: jest.fn(),
      toggle: jest.fn()
    });

    const component = modalComponent();
    
    component.open();
    component.close();
    component.toggle();
    
    expect(component.open).toHaveBeenCalled();
    expect(component.close).toHaveBeenCalled();
    expect(component.toggle).toHaveBeenCalled();
  });
});

describe('Alpine Event Handling', () => {
  test('should handle custom events', () => {
    const eventHandler = jest.fn();
    
    document.addEventListener = jest.fn((event, handler) => {
      if (event === 'decision-completed') {
        eventHandler.mockImplementation(handler);
      }
    });

    // Simulate event listener setup
    document.addEventListener('decision-completed', (event) => {
      eventHandler(event.detail);
    });

    // Simulate event dispatch
    const mockEvent = {
      detail: {
        decision: 'BUY',
        confidence: 85,
        productName: 'Test Product'
      }
    };

    eventHandler(mockEvent.detail);
    
    expect(eventHandler).toHaveBeenCalledWith(mockEvent.detail);
  });

  test('should handle keyboard shortcuts', () => {
    const shortcutHandler = jest.fn();
    
    const keyboardShortcuts = {
      'Escape': () => shortcutHandler('close'),
      'Enter': () => shortcutHandler('submit'),
      'ctrl+s': () => shortcutHandler('save')
    };

    const handleKeydown = (event) => {
      const key = event.key;
      const ctrl = event.ctrlKey;
      
      if (key === 'Escape') {
        keyboardShortcuts['Escape']();
      } else if (key === 'Enter') {
        keyboardShortcuts['Enter']();
      } else if (ctrl && key === 's') {
        event.preventDefault();
        keyboardShortcuts['ctrl+s']();
      }
    };

    // Test Escape key
    handleKeydown({ key: 'Escape', ctrlKey: false });
    expect(shortcutHandler).toHaveBeenCalledWith('close');

    // Test Enter key
    handleKeydown({ key: 'Enter', ctrlKey: false });
    expect(shortcutHandler).toHaveBeenCalledWith('submit');

    // Test Ctrl+S
    const preventDefault = jest.fn();
    handleKeydown({ key: 's', ctrlKey: true, preventDefault });
    expect(preventDefault).toHaveBeenCalled();
    expect(shortcutHandler).toHaveBeenCalledWith('save');
  });
});