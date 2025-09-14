# BuyOrBye Project — Complete Architecture & Rules (CLAUDE.md)

> **FE location:** all front-end code, templates, and assets live under `./cmd/web`.

---

## Project Description
Go-based web application using layered architecture with strict separation of concerns. Purchase decision platform with a GORM persistence layer and Templ templating.

- See `@go.mod` for exact dependency versions.
- See `@README.md` for project overview.

---

## Tech Stack
- **Language:** Go `1.24.4`
- **Web Framework:** `gin-gonic/gin v1.10.1`
- **ORM:** `gorm.io/gorm v1.30.4`
- **Database:** `mysql v1.9.3` (prod) / `sqlite v1.6.0` (dev/test)
- **Templating:** `a-h/templ v0.3.943`
- **Front-End:** Tailwind CSS, HTMX, AlpineJS (advanced visual effects & animations), Templ-generated HTML
- **Validation:** `go-playground/validator v10.27.0`
- **Auth:** `golang-jwt/jwt/v5 v5.3.0`
- **Testing:** `testing` + `stretchr/testify v1.10.0` + `testcontainers-go v0.38.0`
- **Build:** Go modules (+ Tailwind CLI)

---

## Environment Setup
- Go `1.24.4+`
- MySQL for production, SQLite for development/testing
- `go mod download` for dependencies
- Env vars via `.env` with `godotenv v1.5.1`
- Docker required for `testcontainers` integration tests
- `go install github.com/a-h/templ/cmd/templ@latest` for Templ generation
- Tailwind CLI: standalone binary **or** `npx tailwindcss` (Node optional; standalone preferred)

---

## Commands

### App
- `go run cmd/app/main.go` — Start dev server
- `go build -o bin/app cmd/app/main.go` — Build binary
- `go vet ./...` — Static analysis

### Templ
- `templ generate` — Generate Go from `.templ` files (must run before build)
- `templ generate --watch` — Watch mode for development

### Tests
- `go test ./...` — All tests
- `go test -race ./...` — With race detector
- `go test -cover ./...` — Coverage
- `go test -v ./internal/services/...` — Services only
- `go test -v ./internal/repositories/...` — Repos only
- `go test -v ./internal/handlers/...` — Handlers only

### Front-End
- **Dev (watch):**  
  `tailwindcss -c cmd/web/tailwind.config.js -i cmd/web/css/input.css -o cmd/web/static/css/app.css --watch`
- **Prod (minify):**  
  `tailwindcss -c cmd/web/tailwind.config.js -i cmd/web/css/input.css -o cmd/web/static/css/app.min.css --minify`

### Build Sequence (MANDATORY ORDER)
```bash
# Development
templ generate
tailwindcss -c cmd/web/tailwind.config.js -i cmd/web/css/input.css -o cmd/web/static/css/app.css
go run cmd/app/main.go

# Production
templ generate
tailwindcss -c cmd/web/tailwind.config.js -i cmd/web/css/input.css -o cmd/web/static/css/app.min.css --minify
go build -o bin/app cmd/app/main.go
```

---

## Architecture Rules — **STRICT ENFORCEMENT**

### Layer Responsibilities
- **Handlers:** HTTP transport only, use **DTOs** exclusively, render Templ templates or JSON
- **Services:** Business logic only, use **domain** structs exclusively
- **Repositories:** Persistence with **GORM** only, use **model** structs exclusively
- **Templates (Templ):** Presentation only; render **DTOs**; never see domain/models
- **Front-End (HTMX/Alpine/Tailwind):** UI behavior, interactivity, progressive enhancement only
- **No cross-layer violations allowed**

### Data Flow (**MANDATORY**)
```
Middleware → Validation → Handler → Service → Repository → Database
              DTO      →  Domain  →  Model
              ↓
         Templ Templates ← Handler (DTOs only)
              ↓
         HTMX requests → Handlers (partial DTO render) → Templ partials
```

### GORM Usage Rules
- GORM code **ONLY** in repository layer
- Never use GORM in services or handlers
- All DB operations via repository interfaces
- Use GORM v1.30.4 features appropriately (transactions, preloading, etc.)
- AutoMigrate for schema management (not SQL files)

### Templ Usage Rules
- Templates render **DTOs only**
- Generate with `templ generate` before building
- **Partial contract:** HTMX endpoints return **fragment components** (no full layout)
- Import DTOs explicitly from `internal/types`
- Use proper Templ syntax with @ for composition

---

## Claude Code CLI Specific Rules

### File Creation Order (CRITICAL)
Claude must ALWAYS create files in this sequence:
1. Domain models first (`internal/domain/`)
2. DTOs second (`internal/types/`)
3. Repository interfaces in services
4. Repository implementations
5. Services
6. Handlers
7. Templ templates LAST
8. Run `templ generate` after creating .templ files
9. Run Tailwind build after CSS changes

NEVER create templates before DTOs exist
NEVER create handlers before services exist

### Template Development Rules
- Write `.templ` files, not `.html`
- Run `templ generate` after EVERY .templ change
- Create golden test files FIRST (`testdata/*.golden.html`)
- Never mix partial and page responses
- Always include proper package declaration

Example Claude must follow:
```templ
package pages

import "internal/types"

templ PurchasePage(dto types.PurchasePageDTO) {
    @components.Layout(dto.Title) {
        // content
    }
}
```

### Build Dependencies Check
Claude must verify before running:
```bash
which templ        # Must be installed (v0.3.943)
which tailwindcss  # Must be v4.1.13+
tailwindcss --version  # Verify version
ls .env           # Must exist with required vars
```

---

## Project Structure

```
cmd/
  app/
    main.go               # Gin router, middleware, static mount
  web/                    # FRONT-END ROOT
    templates/            # .templ files (pages, partials, components)
      layouts/            # Base layouts, shells
      pages/              # Full pages: *_page.templ
      partials/           # HTMX fragments: *_partial.templ  
      components/         # Reusable UI bits (Card, Toast, Table...)
      shared/             # Meta tags, scripts includes
    static/
      css/
        app.css
        app.min.css
      js/
        alpine.boot.js    # Alpine init (stores, transitions)
        htmx.boot.js      # HTMX config (headers, swapping)
      img/
    css/
      input.css           # Tailwind layers (@tailwind base/components/utilities)
    tailwind.config.js
    postcss.config.js     # optional
    
internal/
  database/               # GORM setup/migrations
    migration.go          # AutoMigrate all models
  models/                 # DB models (repository layer only)
  domain/                 # Business entities & rules (pure)
  repositories/           # GORM implementations
  services/               # Business logic
  handlers/               # HTTP endpoints (DTOs ↔ services)
  types/                  # DTOs organized by feature
    auth_dto.go          # All auth-related DTOs
    purchase_dto.go      # All purchase-related DTOs
    finance_dto.go       # All finance-related DTOs
    health_dto.go        # All health-related DTOs
    decision_dto.go      # All decision-related DTOs
  middleware/             # CORS, JWT, validation, CSRF
  
tests/
  integration/
  testutils/
  testdata/
    *.golden.html        # Golden test files for templates
```

> **Note:** New FE work must live under `./cmd/web`. Migrate legacy templates gradually.

---

## Front-End Architecture

### Principles
- **Progressive enhancement:** App works without JS; HTMX/Alpine enhance progressively.
- **Separation of concerns:**  
  - **HTMX** — server-driven interactions (requests, swaps, partial updates)  
  - **Alpine** — view-layer state only, micro-interactions, animations (no business logic)  
  - **Tailwind** — styling, tokens, responsive behavior
- **Fragments first:** Every interactive region has a matching Templ **partial**.
- **Business logic remains server-side** (services). Browser holds **presentation** only.

### HTMX Standards

#### Route Patterns (STRICT)
Pages (full render):
- `GET /purchases` → `PurchaseIndexPage` 
- `GET /purchases/:id` → `PurchaseDetailPage`
- `GET /auth/login` → `LoginPage`

Partials (fragments only):
- `GET /ui/partials/purchases/list` → `PurchaseListPartial`
- `POST /ui/partials/purchases/:id/update` → `PurchaseUpdatePartial`
- `DELETE /ui/partials/purchases/:id` → empty or redirect header
- `GET /ui/partials/notifications` → `NotificationListPartial`

Claude must NEVER mix these patterns.

#### HTMX Configuration
```javascript
// cmd/web/static/js/htmx.boot.js
document.body.addEventListener('htmx:configRequest', (e) => {
    // CSRF Token injection
    const token = document.querySelector('meta[name="csrf-token"]')?.content
    if (token) e.detail.headers['X-CSRF-Token'] = token
    
    // Add request ID for tracing
    e.detail.headers['X-Request-ID'] = crypto.randomUUID()
})

// Global error handling
document.body.addEventListener('htmx:responseError', (e) => {
    const target = e.detail.target
    target.innerHTML = '<div class="alert alert-error">Something went wrong. Please try again.</div>'
})
```

Use `hx-boost="true"` for PJAX-style nav where safe.
Use `hx-trigger` deliberately (`change`, `keyup changed delay:300ms`, `revealed`, etc.).
Optional SSE: `hx-sse="connect:/events"`.

### Alpine.js Standards

#### State Management Rules
Claude must NEVER:
- Store business data in Alpine (only UI state)
- Make API calls from Alpine (use HTMX)
- Calculate business logic in Alpine

Claude must ALWAYS:
- Use Alpine only for: open/closed, active tab, animation state, form validation hints
- Keep Alpine data ephemeral and UI-focused
- Sync with server via HTMX events

#### Alpine Configuration
```javascript
// cmd/web/static/js/alpine.boot.js
document.addEventListener('alpine:init', () => {
    // Global stores for UI state only
    Alpine.store('ui', {
        sidebarOpen: false,
        theme: localStorage.getItem('theme') || 'light',
        toggleSidebar() {
            this.sidebarOpen = !this.sidebarOpen
        }
    })
    
    // Magic helpers
    Alpine.magic('formatCurrency', () => (amount) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD'
        }).format(amount)
    })
})
```

Use declarative bindings: `:class`, `:aria-expanded`, `@click`, `@keydown.escape.window`.

### Tailwind Standards
- Centralize tokens in `tailwind.config.js`
- Utility-first; extract patterns with `@apply` in `cmd/web/css/input.css` under `@layer components`
- Name UI regions with stable IDs for HTMX targets
- Performance constraints:
  - Bundle size: `app.min.css < 50KB`
  - Use PurgeCSS in production
  - Critical CSS inline in production layout

---

## Lazy Loading Implementation

### Component-Level Lazy Loading
Goal: Load each component/region independently when visible.

#### Lazy Slot Component (Templ)
```templ
// cmd/web/templates/components/lazy_slot.templ
package components

templ LazySlot(id, url string) {
    <div id={ id }
         hx-get={ url }
         hx-trigger="revealed"
         hx-swap="outerHTML"
         hx-indicator="#global-indicator">
        @SkeletonCard()
    </div>
}

templ SkeletonCard() {
    <div class="animate-pulse">
        <div class="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
        <div class="h-4 bg-gray-200 rounded w-1/2"></div>
    </div>
}
```

#### Handler for Partial
```go
// internal/handlers/purchase_ui.go
func (h *PurchaseUI) ListPartial(c *gin.Context) {
    items, err := h.service.ListRecent(c.Request.Context())
    if err != nil {
        c.Status(http.StatusInternalServerError)
        // Return error partial, not full page
        templates.RenderErrorBannerPartial(c, types.ErrorDTO{
            Message: "Failed to load purchases",
        })
        return
    }
    
    dto := types.PurchaseListDTO{
        Items: types.FromDomainList(items),
    }
    templates.RenderPurchaseListPartial(c, dto) // fragment only
}
```

Claude must NOT lazy load critical above-fold content.

---

## Template Error Handling Pattern

Claude must implement error states in every template:

```templ
templ PurchaseListPartial(dto types.PurchaseListDTO) {
    if dto.Error != nil {
        @components.ErrorBanner(dto.Error.Message)
    } else if len(dto.Items) == 0 {
        @components.EmptyState("No purchases found", "Start by adding your first purchase")
    } else {
        <ul class="divide-y">
            for _, item := range dto.Items {
                @PurchaseListItem(item)
            }
        </ul>
    }
}
```

---

## CSRF Implementation (Complete)

1. **Middleware sets token:**
```go
// internal/middleware/csrf.go
func CSRFMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := generateCSRFToken()
        c.Set("csrf_token", token)
        c.SetCookie("csrf", token, 3600, "/", "", true, true)
        c.Next()
    }
}
```

2. **Handler passes to layout:**
```go
dto := types.LayoutDTO{
    Title:     "Dashboard",
    CSRFToken: c.GetString("csrf_token"),
}
```

3. **Layout renders meta tag:**
```templ
templ Layout(dto types.LayoutDTO) {
    <meta name="csrf-token" content={ dto.CSRFToken }/>
}
```

4. **HTMX reads from meta** (see htmx.boot.js above)

Claude must verify all 4 steps are implemented.

---

## DTO Organization Rules

### File Structure (STRICT)
Claude must organize DTOs by feature, not layer:
```go
// internal/types/auth_dto.go
type LoginRequestDTO struct { }
type LoginResponseDTO struct { }
type RegisterRequestDTO struct { }
type TokenResponseDTO struct { }

// internal/types/purchase_dto.go  
type PurchaseCreateDTO struct { }
type PurchaseListDTO struct { }
type PurchaseDetailDTO struct { }

// NEVER create generic dto.go with mixed concerns
```

### DTO Validation
```go
// Always use validation tags
type PurchaseCreateDTO struct {
    Name     string  `json:"name" binding:"required,min=2,max=100"`
    Amount   float64 `json:"amount" binding:"required,gt=0"`
    Category string  `json:"category" binding:"required,oneof=food clothing electronics"`
}

// Initialize validator once
var validatorInstance = validator.New()
```

---

## Testing Patterns

### Front-End Testing Sequence
1. Create golden test files first (`testdata/*.golden.html`)
2. Write handler tests with expected HTML fragments
3. Test HTMX interactions with httptest
4. Verify partial vs full page responses

### Golden Tests for Templates
```go
func TestPurchaseListPartial_Renders(t *testing.T) {
    dto := types.PurchaseListDTO{
        Items: []types.PurchaseItemDTO{
            {ID: "1", Name: "Coffee", Amount: 4.50},
        },
    }
    
    var buf bytes.Buffer
    err := templates.PurchaseListPartial(dto).Render(context.Background(), &buf)
    require.NoError(t, err)
    
    golden := filepath.Join("testdata", "purchase_list.golden.html")
    if *update {
        os.WriteFile(golden, buf.Bytes(), 0644)
    }
    
    expected, _ := os.ReadFile(golden)
    assert.Equal(t, string(expected), buf.String())
}
```

### Handler Test Patterns
```go
func TestPurchaseHandler_ListPartial_ReturnsFragment(t *testing.T) {
    // Setup
    mockService := &mocks.PurchaseService{}
    handler := NewPurchaseHandler(mockService)
    
    // Mock expectations
    mockService.On("ListRecent", mock.Anything).Return(purchases, nil)
    
    // Request
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/ui/partials/purchases/list", nil)
    
    router := gin.New()
    router.GET("/ui/partials/purchases/list", handler.ListPartial)
    router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), `<ul class="divide-y">`)
    assert.NotContains(t, w.Body.String(), `<!DOCTYPE`)  // Ensure partial
}
```

---

## Environment-Specific Configuration

### Development
```env
ENV=development
DB_DRIVER=sqlite
DB_DSN=file:dev.db?cache=shared&mode=rwc
LOG_LEVEL=debug
TEMPL_DEV=true
TAILWIND_DEV=true
```

### Production  
```env
ENV=production
DB_DRIVER=mysql
DB_DSN=user:pass@tcp(localhost:3306)/buyorbye?parseTime=true
LOG_LEVEL=info
TEMPL_DEV=false
TAILWIND_DEV=false
CACHE_STATIC=31536000  # 1 year for fingerprinted assets
```

---

## Performance Guidelines

### Front-End Performance
- Bundle size: `app.min.css < 50KB`
- No JS framework > 10KB
- Lazy load below-fold components
- Use `loading="lazy"` for images
- Inline critical CSS in production
- Cache static assets with fingerprinting

### Database Performance
- Use GORM preloading for related data: `db.Preload("Purchases").Find(&users)`
- Implement pagination: `db.Limit(20).Offset(page * 20)`
- Add indexes on frequently queried fields
- Monitor for N+1 queries
- Use batch operations where possible

---

## Security Rules

### Input Validation
- Validate all input at handler layer using `go-playground/validator`
- Use struct tags for automatic validation
- Sanitize user input before storage

### Template Security
- Templ auto-escapes by default (XSS safe)
- Never use `templ.Raw()` with user input
- Use CSP headers to prevent inline script injection

### Authentication & Authorization
- JWT tokens with 15-minute expiry
- Refresh tokens with 7-day expiry
- Bcrypt cost 14 for passwords
- CSRF protection on all state-changing operations
- Rate limiting on auth endpoints

---

## Development Workflow

1. Write failing tests (service/handler/template golden)
2. Implement minimal code to pass
3. Refactor with tests green
4. Run `templ generate` if templates changed
5. Build Tailwind (watch in dev; minify for prod)
6. Run full test suite with race detector
7. Verify strict layer separation
8. Confirm coverage thresholds
9. Validate config across environments
10. Ensure structured logging (zap); remove `fmt.Print`
11. Manually verify HTMX lazy loading & partial contracts

### Parallel Development Setup
```bash
# Terminal 1: Tailwind watch
tailwindcss -c cmd/web/tailwind.config.js -i cmd/web/css/input.css -o cmd/web/static/css/app.css --watch

# Terminal 2: Templ watch
templ generate --watch

# Terminal 3: Go with auto-reload (using air or similar)
air
```

---

## Do Not

- Import GORM in services/handlers
- Use model structs outside repository layer
- Use DTOs in services (domain only)
- Pass domain/models to Templ templates
- Call repositories from handlers (go through services)
- Put business logic in handlers, repositories, or front-end code
- Commit without full tests (incl. race)
- Skip error handling at boundaries
- Use direct SQL where GORM suffices
- Expose internal model structures in API responses
- Forget `templ generate` before building
- Import concrete repo impls in service constructors (use interfaces)
- Mix MySQL/SQLite specifics without compatibility handling
- Create DTOs outside `internal/types/`
- Mix different domain DTOs in the same file
- Omit validation struct tags on DTOs
- Return full pages from partial endpoints
- Add inline scripts/styles that break CSP without hashes
- Store business data in Alpine.js
- Make API calls from Alpine (use HTMX)
- Lazy load above-fold content
- Create templates before DTOs exist
- Create handlers before services exist

---

## Claude Instructions

- Use "think hard" for new layer interactions or complex UX/server contracts
- Use "ultrathink" for architectural refactors across FE/BE boundaries
- Enforce strict separation (DTOs ↔ Templates; Services ↔ Repos)
- Start with tests (incl. template golden tests) for new UI flows
- Consider MySQL/SQLite compatibility in all GORM code
- Always follow the file creation order (Domain → DTO → Repo → Service → Handler → Template)
- Run `templ generate` after every .templ file change
- Verify all 4 CSRF implementation steps
- Create golden test files before implementing templates
- Never mix page and partial endpoints
- Use Alpine only for UI state, never business logic
- Organize DTOs by feature, not by request/response type
- Check for `templ` and `tailwindcss` before running builds

---

## Common Pitfalls & Solutions

| Issue | Solution |
|-------|----------|
| Template not rendering | Run `templ generate` |
| CSS not updating | Check Tailwind watch is running |
| HTMX not sending CSRF | Verify all 4 CSRF steps |
| Partial returns full page | Check handler returns fragment only |
| Alpine storing business data | Move to server, use HTMX |
| N+1 queries | Use GORM Preload |
| Tests failing on CI | Check MySQL/SQLite compatibility |
| Golden tests failing | Update with `--update` flag |
| DTOs mixed up | Organize by feature, not layer |
| Build fails | Check file creation order |