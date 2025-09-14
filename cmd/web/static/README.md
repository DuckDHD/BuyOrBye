# Static Assets

This directory contains all static assets for the BuyOrBye web application.

## Directory Structure

```
cmd/web/static/
├── css/
│   ├── input.css          # Tailwind CSS source file
│   └── output.css         # Generated CSS (not in version control)
├── js/
│   ├── htmx.boot.js       # HTMX configuration and error handling
│   └── alpine.boot.js     # Alpine.js global stores and helpers
├── icons/
│   ├── decision-*.svg     # Decision-specific icons (buy, wait, bye)
│   ├── category-*.svg     # Category icons (electronics, clothing, etc.)
│   └── ui-*.svg           # UI icons (close, menu, chevron, etc.)
├── images/
│   └── (app icons and images)
├── dist/                  # Built/optimized assets (not in version control)
│   ├── css/
│   ├── js/
│   ├── icons/
│   └── images/
├── favicon.svg            # App favicon
├── manifest.json          # PWA manifest
└── robots.txt             # Search engine directives
```

## Build Process

### Development
```bash
# Watch for changes and rebuild assets
make assets-watch

# Or with npm (if available)
npm run dev
```

### Production
```bash
# Build optimized assets
make assets-build

# Or with npm
npm run build
```

## CSS Architecture

### Tailwind CSS Structure
- **Base Layer**: HTML element defaults and resets
- **Components Layer**: Reusable component classes
- **Utilities Layer**: Low-level utility classes

### Custom Components
- `.decision-card` - Cards for purchase decisions
- `.btn--*` - Button variants (primary, secondary, success, etc.)
- `.input-primary` - Styled form inputs
- `.badge--*` - Status badges with color variants
- `.toast--*` - Notification toast styles

### Color System
- **Primary**: Blue color palette for main UI elements
- **Decision Colors**: Green (BUY), Yellow (WAIT), Red (BYE)
- **Health Colors**: Risk level indicators
- **Finance Colors**: Income, expense, loan, savings indicators

## JavaScript Architecture

### HTMX Boot System (`htmx.boot.js`)
- **CSRF Protection**: Automatic token injection
- **Error Handling**: Global error management and user feedback
- **Request Logging**: Development debugging and monitoring
- **Retry Logic**: Automatic retry for failed requests
- **Loading States**: UI loading indicators

### Alpine Boot System (`alpine.boot.js`)
- **Global Stores**: UI state management (sidebar, theme, modals, toasts)
- **Magic Helpers**: Currency formatting, date formatting, validation
- **Utility Functions**: Debounce, throttle, clipboard operations
- **Decision Helpers**: Color coding and formatting for decisions

### Store Usage Examples
```javascript
// Theme management
Alpine.store('ui').theme.toggle();

// Toast notifications
Alpine.store('ui').toasts.success('Operation completed');

// Modal management
Alpine.store('modals').open('my-modal-id');

// Magic helpers in templates
$currency(1234.56)     // Returns "$1,234.56"
$date(new Date())      // Returns formatted date
$validate(value, 'required|email')  // Validation
```

## Icon System

### Categories
- **Decision Icons**: Visual representations for BUY/WAIT/BYE decisions
- **Category Icons**: Product category representations
- **UI Icons**: Common interface elements

### Usage
Icons are optimized SVGs that use `currentColor` for easy theming:
```html
<svg class="w-5 h-5 text-green-600">
  <!-- SVG content -->
</svg>
```

## Optimization

### CSS Optimization
- **PurgeCSS**: Removes unused CSS classes
- **Minification**: Reduces file size
- **Gzip Compression**: Server-level compression

### JavaScript Optimization
- **ESBuild**: Fast bundling and minification
- **Tree Shaking**: Removes unused code
- **Target ES2020**: Modern browser compatibility

### Image Optimization
- **ImageMin**: Compresses PNG/JPEG images
- **SVGO**: Optimizes SVG files
- **WebP Conversion**: Modern image format support

## Development Workflow

1. **Start Development Server**:
   ```bash
   make dev  # Includes asset watching
   ```

2. **Add New Components**:
   - Add CSS to `css/input.css` in the components layer
   - Use existing color system and spacing
   - Follow BEM naming convention for complex components

3. **Add New Icons**:
   - Create SVG with `viewBox="0 0 24 24"`
   - Use `currentColor` for fill/stroke
   - Optimize with SVGO

4. **Build for Production**:
   ```bash
   make build-prod  # Full optimized build
   ```

## Browser Support

- **Modern Browsers**: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+
- **Features**: CSS Grid, Custom Properties, ES2020 JavaScript
- **Fallbacks**: Graceful degradation for older browsers

## Performance Targets

- **CSS Bundle**: < 50KB gzipped
- **JavaScript Bundle**: < 100KB gzipped
- **Images**: WebP with JPEG fallback
- **Load Time**: < 2s on 3G connection