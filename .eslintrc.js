module.exports = {
  env: {
    browser: true,
    es2021: true
  },
  extends: [
    'eslint:recommended'
  ],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module'
  },
  globals: {
    Alpine: 'readonly',
    htmx: 'readonly'
  },
  rules: {
    'indent': ['error', 2],
    'linebreak-style': ['error', 'unix'],
    'quotes': ['error', 'single'],
    'semi': ['error', 'always'],
    'no-unused-vars': ['warn'],
    'no-console': ['warn'],
    'no-debugger': ['error'],
    'prefer-const': ['error'],
    'no-var': ['error'],
    'arrow-spacing': ['error'],
    'comma-dangle': ['error', 'never'],
    'object-curly-spacing': ['error', 'always'],
    'array-bracket-spacing': ['error', 'never'],
    'space-before-function-paren': ['error', 'never'],
    'keyword-spacing': ['error'],
    'space-infix-ops': ['error'],
    'eol-last': ['error'],
    'no-trailing-spaces': ['error']
  }
};