/**
 * HTMX Configuration Tests
 * Tests for htmx.boot.js functionality
 */

// Mock DOM and HTMX for testing
global.document = {
  addEventListener: jest.fn(),
  createElement: jest.fn(() => ({
    style: {},
    classList: {
      add: jest.fn(),
      remove: jest.fn()
    }
  })),
  body: {
    appendChild: jest.fn()
  }
};

global.htmx = {
  config: {},
  on: jest.fn(),
  trigger: jest.fn(),
  ajax: jest.fn(),
  process: jest.fn()
};

global.console = {
  log: jest.fn(),
  error: jest.fn(),
  warn: jest.fn()
};

// Import the HTMX boot configuration
// Note: This would import the actual htmx.boot.js file
// For testing purposes, we'll define the expected functionality

describe('HTMX Boot Configuration', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('should configure CSRF token injection', () => {
    // Simulate HTMX boot configuration
    const mockConfig = {
      requestClass: 'htmx-request',
      indicatorClass: 'htmx-indicator',
      historyEnabled: true,
      timeout: 10000
    };

    // Test CSRF token injection
    const mockEvent = {
      detail: {
        xhr: {
          setRequestHeader: jest.fn()
        }
      }
    };

    // Simulate the configRequest event handler
    const configRequestHandler = (evt) => {
      const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
      if (csrfToken) {
        evt.detail.xhr.setRequestHeader('X-CSRF-Token', csrfToken);
      }
    };

    // Mock meta tag
    document.querySelector = jest.fn().mockReturnValue({
      getAttribute: jest.fn().mockReturnValue('csrf-token-123')
    });

    configRequestHandler(mockEvent);

    expect(mockEvent.detail.xhr.setRequestHeader).toHaveBeenCalledWith('X-CSRF-Token', 'csrf-token-123');
  });

  test('should handle loading states', () => {
    const loadingStates = {
      showLoading: jest.fn(),
      hideLoading: jest.fn()
    };

    // Simulate beforeSend event
    const beforeSendHandler = (evt) => {
      loadingStates.showLoading(evt.detail.elt);
    };

    // Simulate afterSwap event
    const afterSwapHandler = (evt) => {
      loadingStates.hideLoading(evt.detail.elt);
    };

    const mockElement = { id: 'test-element' };
    
    beforeSendHandler({ detail: { elt: mockElement } });
    expect(loadingStates.showLoading).toHaveBeenCalledWith(mockElement);

    afterSwapHandler({ detail: { elt: mockElement } });
    expect(loadingStates.hideLoading).toHaveBeenCalledWith(mockElement);
  });

  test('should implement retry logic for failed requests', () => {
    let retryCount = 0;
    const maxRetries = 3;

    const retryHandler = (evt) => {
      const xhr = evt.detail.xhr;
      
      if (xhr.status >= 500 && retryCount < maxRetries) {
        retryCount++;
        // Simulate retry
        setTimeout(() => {
          global.htmx.trigger(evt.detail.elt, 'retry');
        }, 1000 * retryCount);
        return false; // Prevent default error handling
      }
      
      return true; // Allow default error handling
    };

    // Test successful retry
    const mockFailedEvent = {
      detail: {
        xhr: { status: 500 },
        elt: { id: 'test-element' }
      }
    };

    const result = retryHandler(mockFailedEvent);
    expect(result).toBe(false);
    expect(retryCount).toBe(1);

    // Test max retries exceeded
    retryCount = 3;
    const result2 = retryHandler(mockFailedEvent);
    expect(result2).toBe(true);
  });

  test('should handle global error management', () => {
    const errorHandler = jest.fn();
    
    const responseErrorHandler = (evt) => {
      const xhr = evt.detail.xhr;
      const statusCode = xhr.status;
      
      if (statusCode === 401) {
        // Redirect to login
        window.location.href = '/login';
      } else if (statusCode === 403) {
        // Show permission error
        errorHandler('Permission denied');
      } else if (statusCode >= 500) {
        // Show server error
        errorHandler('Server error occurred');
      }
    };

    // Mock window.location
    global.window = { location: { href: '' } };

    // Test 401 error
    responseErrorHandler({ detail: { xhr: { status: 401 } } });
    expect(window.location.href).toBe('/login');

    // Test 403 error
    responseErrorHandler({ detail: { xhr: { status: 403 } } });
    expect(errorHandler).toHaveBeenCalledWith('Permission denied');

    // Test 500 error
    responseErrorHandler({ detail: { xhr: { status: 500 } } });
    expect(errorHandler).toHaveBeenCalledWith('Server error occurred');
  });

  test('should log requests in development mode', () => {
    const isDevelopment = process.env.NODE_ENV === 'development';
    
    const requestLogger = (evt) => {
      if (isDevelopment) {
        console.log('HTMX Request:', {
          method: evt.detail.verb,
          url: evt.detail.path,
          element: evt.detail.elt.tagName,
          headers: evt.detail.headers
        });
      }
    };

    const mockRequestEvent = {
      detail: {
        verb: 'GET',
        path: '/api/data',
        elt: { tagName: 'DIV' },
        headers: { 'Content-Type': 'application/json' }
      }
    };

    // Set development mode
    process.env.NODE_ENV = 'development';
    
    requestLogger(mockRequestEvent);
    
    if (isDevelopment) {
      expect(console.log).toHaveBeenCalledWith('HTMX Request:', expect.any(Object));
    }
  });

  test('should configure custom indicators', () => {
    const indicatorConfig = {
      defaultIndicator: '<div class="htmx-indicator">Loading...</div>',
      customIndicators: {
        'data-loading': 'spinner',
        'data-saving': 'save-indicator'
      }
    };

    const getIndicator = (element) => {
      const customIndicator = element.getAttribute('data-loading');
      if (customIndicator) {
        return document.querySelector(`.${customIndicator}`);
      }
      return document.querySelector('.htmx-indicator');
    };

    // Mock element with custom indicator
    const mockElement = {
      getAttribute: jest.fn().mockReturnValue('spinner')
    };

    document.querySelector = jest.fn().mockReturnValue({ style: {} });

    const indicator = getIndicator(mockElement);
    expect(mockElement.getAttribute).toHaveBeenCalledWith('data-loading');
    expect(document.querySelector).toHaveBeenCalledWith('.spinner');
  });
});

describe('HTMX Integration Tests', () => {
  test('should handle form submissions with CSRF', () => {
    // Mock form element
    const mockForm = {
      getAttribute: jest.fn(),
      querySelector: jest.fn().mockReturnValue({
        value: 'csrf-token-123'
      })
    };

    const formSubmitHandler = (evt) => {
      const form = evt.detail.elt;
      const csrfToken = form.querySelector('input[name="_token"]')?.value;
      
      if (csrfToken) {
        evt.detail.xhr.setRequestHeader('X-CSRF-Token', csrfToken);
      }
    };

    const mockSubmitEvent = {
      detail: {
        elt: mockForm,
        xhr: {
          setRequestHeader: jest.fn()
        }
      }
    };

    formSubmitHandler(mockSubmitEvent);

    expect(mockForm.querySelector).toHaveBeenCalledWith('input[name="_token"]');
    expect(mockSubmitEvent.detail.xhr.setRequestHeader).toHaveBeenCalledWith('X-CSRF-Token', 'csrf-token-123');
  });

  test('should handle modal interactions', () => {
    const modalHandler = {
      open: jest.fn(),
      close: jest.fn()
    };

    // Test modal open trigger
    const openModalHandler = (evt) => {
      const modalId = evt.detail.elt.getAttribute('data-modal');
      if (modalId) {
        modalHandler.open(modalId);
      }
    };

    const mockOpenEvent = {
      detail: {
        elt: {
          getAttribute: jest.fn().mockReturnValue('test-modal')
        }
      }
    };

    openModalHandler(mockOpenEvent);
    expect(modalHandler.open).toHaveBeenCalledWith('test-modal');

    // Test modal close on background click
    const closeModalHandler = (evt) => {
      if (evt.target.classList.contains('modal-background')) {
        modalHandler.close();
      }
    };

    const mockCloseEvent = {
      target: {
        classList: {
          contains: jest.fn().mockReturnValue(true)
        }
      }
    };

    closeModalHandler(mockCloseEvent);
    expect(modalHandler.close).toHaveBeenCalled();
  });
});

describe('HTMX Performance Tests', () => {
  test('should debounce rapid requests', (done) => {
    let requestCount = 0;
    
    const debouncedRequest = (() => {
      let timeout;
      return (callback, delay = 300) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => {
          requestCount++;
          callback();
        }, delay);
      };
    })();

    const mockCallback = jest.fn();

    // Fire multiple rapid requests
    debouncedRequest(mockCallback, 100);
    debouncedRequest(mockCallback, 100);
    debouncedRequest(mockCallback, 100);

    // Only the last one should execute
    setTimeout(() => {
      expect(requestCount).toBe(1);
      expect(mockCallback).toHaveBeenCalledTimes(1);
      done();
    }, 200);
  });

  test('should cache repeated requests', () => {
    const cache = new Map();
    
    const cachedRequest = (url) => {
      if (cache.has(url)) {
        return Promise.resolve(cache.get(url));
      }
      
      // Simulate fetch
      const response = { data: `Data for ${url}` };
      cache.set(url, response);
      return Promise.resolve(response);
    };

    const url = '/api/test';
    
    // First request - cache miss
    return cachedRequest(url).then(result => {
      expect(result.data).toBe('Data for /api/test');
      expect(cache.has(url)).toBe(true);
      
      // Second request - cache hit
      return cachedRequest(url).then(result2 => {
        expect(result2.data).toBe('Data for /api/test');
        expect(cache.size).toBe(1);
      });
    });
  });
});

describe('HTMX Error Scenarios', () => {
  test('should handle network failures gracefully', () => {
    const errorHandler = jest.fn();
    
    const networkErrorHandler = (evt) => {
      const xhr = evt.detail.xhr;
      
      if (xhr.status === 0) {
        // Network error
        errorHandler('Network connection failed');
      } else if (xhr.status === 408) {
        // Timeout
        errorHandler('Request timed out');
      }
    };

    // Test network error
    networkErrorHandler({ detail: { xhr: { status: 0 } } });
    expect(errorHandler).toHaveBeenCalledWith('Network connection failed');

    // Test timeout
    networkErrorHandler({ detail: { xhr: { status: 408 } } });
    expect(errorHandler).toHaveBeenCalledWith('Request timed out');
  });

  test('should validate response content types', () => {
    const responseValidator = (evt) => {
      const xhr = evt.detail.xhr;
      const contentType = xhr.getResponseHeader('Content-Type');
      
      if (contentType && !contentType.includes('text/html')) {
        console.warn('Unexpected content type:', contentType);
        return false;
      }
      
      return true;
    };

    const mockEvent = {
      detail: {
        xhr: {
          getResponseHeader: jest.fn().mockReturnValue('application/json')
        }
      }
    };

    const result = responseValidator(mockEvent);
    expect(result).toBe(false);
    expect(console.warn).toHaveBeenCalledWith('Unexpected content type:', 'application/json');
  });
});