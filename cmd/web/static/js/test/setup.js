/**
 * Jest setup file for frontend tests
 * This file runs before each test suite
 */

// Mock global objects that don't exist in Jest/Node environment
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;

// Mock fetch if not available
if (!global.fetch) {
  global.fetch = jest.fn(() =>
    Promise.resolve({
      json: () => Promise.resolve({}),
      text: () => Promise.resolve(''),
      ok: true,
      status: 200,
      headers: new Map()
    })
  );
}

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn()
};
global.localStorage = localStorageMock;

// Mock sessionStorage
global.sessionStorage = localStorageMock;

// Mock URL and URLSearchParams
global.URL = URL;
global.URLSearchParams = URLSearchParams;

// Mock IntersectionObserver
global.IntersectionObserver = jest.fn().mockImplementation((callback) => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
  root: null,
  rootMargin: '',
  thresholds: []
}));

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation((callback) => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn()
}));

// Mock MutationObserver
global.MutationObserver = jest.fn().mockImplementation((callback) => ({
  observe: jest.fn(),
  disconnect: jest.fn(),
  takeRecords: jest.fn()
}));

// Mock CSS.supports
global.CSS = {
  supports: jest.fn().mockReturnValue(true)
};

// Mock matchMedia
global.matchMedia = jest.fn().mockImplementation(query => ({
  matches: false,
  media: query,
  onchange: null,
  addListener: jest.fn(),
  removeListener: jest.fn(),
  addEventListener: jest.fn(),
  removeEventListener: jest.fn(),
  dispatchEvent: jest.fn()
}));

// Mock requestAnimationFrame and cancelAnimationFrame
global.requestAnimationFrame = jest.fn().mockImplementation(cb => setTimeout(cb, 0));
global.cancelAnimationFrame = jest.fn().mockImplementation(id => clearTimeout(id));

// Mock getComputedStyle
global.getComputedStyle = jest.fn().mockImplementation(() => ({
  getPropertyValue: jest.fn().mockReturnValue(''),
  setProperty: jest.fn(),
  removeProperty: jest.fn()
}));

// Console spy setup - capture console methods for testing
global.console = {
  ...console,
  log: jest.fn(),
  warn: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  debug: jest.fn()
};

// Mock performance API
global.performance = {
  ...performance,
  mark: jest.fn(),
  measure: jest.fn(),
  getEntriesByName: jest.fn().mockReturnValue([]),
  getEntriesByType: jest.fn().mockReturnValue([]),
  clearMarks: jest.fn(),
  clearMeasures: jest.fn(),
  now: jest.fn().mockReturnValue(Date.now())
};

// Setup and teardown for each test
beforeEach(() => {
  // Clear all mocks before each test
  jest.clearAllMocks();
  
  // Reset DOM
  document.head.innerHTML = '';
  document.body.innerHTML = '';
  
  // Reset localStorage and sessionStorage
  localStorageMock.getItem.mockReturnValue(null);
  localStorageMock.setItem.mockImplementation(() => {});
  localStorageMock.removeItem.mockImplementation(() => {});
  localStorageMock.clear.mockImplementation(() => {});
});

afterEach(() => {
  // Cleanup after each test
  jest.restoreAllMocks();
});