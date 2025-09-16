/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    // Template files - all .templ files in the templates directory
    "./templates/**/*.templ",
    "./templates/**/*.html",

    // Generated template files - the _templ.go files contain classes
    "./templates/**/*_templ.go",

    // Static JavaScript files
    "./static/**/*.js",
    "./src/**/*.js",

    // Go source files that might contain CSS classes in strings
    "../../internal/handlers/**/*.go"
  ],

  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#172554',
        },
        secondary: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
          950: '#020617',
        },
        success: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#14532d',
          950: '#052e16',
        },
        warning: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
          950: '#451a03',
        },
        danger: {
          50: '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          300: '#fca5a5',
          400: '#f87171',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
          800: '#991b1b',
          900: '#7f1d1d',
          950: '#450a0a',
        },
      },

      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'Monaco', 'Consolas', 'monospace'],
      },

      fontSize: {
        'xs': ['0.75rem', { lineHeight: '1rem' }],
        'sm': ['0.875rem', { lineHeight: '1.25rem' }],
        'base': ['1rem', { lineHeight: '1.5rem' }],
        'lg': ['1.125rem', { lineHeight: '1.75rem' }],
        'xl': ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],
        '4xl': ['2.25rem', { lineHeight: '2.5rem' }],
        '5xl': ['3rem', { lineHeight: '1' }],
        '6xl': ['3.75rem', { lineHeight: '1' }],
        '7xl': ['4.5rem', { lineHeight: '1' }],
        '8xl': ['6rem', { lineHeight: '1' }],
        '9xl': ['8rem', { lineHeight: '1' }],
      },

      spacing: {
        '18': '4.5rem',
        '88': '22rem',
        '128': '32rem',
      },

      borderRadius: {
        '4xl': '2rem',
      },

      boxShadow: {
        'soft': '0 2px 15px -3px rgba(0, 0, 0, 0.07), 0 10px 20px -2px rgba(0, 0, 0, 0.04)',
        'strong': '0 10px 25px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
      },

      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'pulse-slow': 'pulse 3s infinite',
        'bounce-slow': 'bounce 2s infinite',
      },

      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(100%)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-100%)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
      },

      backdropBlur: {
        xs: '2px',
      },

      screens: {
        'xs': '475px',
        '3xl': '1680px',
      },
    },
  },

  plugins: [
    // Forms plugin for better form styling
    require('@tailwindcss/forms'),

    // Typography plugin for rich text content
    require('@tailwindcss/typography'),

    // Aspect ratio plugin for maintaining aspect ratios
    require('@tailwindcss/aspect-ratio'),

    // Container queries plugin
    require('@tailwindcss/container-queries'),
  ],

  // Dark mode configuration
  darkMode: 'class', // Use 'class' strategy for manual dark mode toggle

  // Safelist important classes that might be dynamically generated
  safelist: [
    'bg-red-50',
    'bg-red-100',
    'bg-yellow-50',
    'bg-yellow-100',
    'bg-green-50',
    'bg-green-100',
    'bg-blue-50',
    'bg-blue-100',
    'text-red-800',
    'text-yellow-800',
    'text-green-800',
    'text-blue-800',
    'border-red-200',
    'border-yellow-200',
    'border-green-200',
    'border-blue-200',
    // Dark mode variants
    'dark:bg-red-900',
    'dark:bg-yellow-900',
    'dark:bg-green-900',
    'dark:bg-blue-900',
    'dark:text-red-200',
    'dark:text-yellow-200',
    'dark:text-green-200',
    'dark:text-blue-200',
    'dark:border-red-800',
    'dark:border-yellow-800',
    'dark:border-green-800',
    'dark:border-blue-800',
  ],
}