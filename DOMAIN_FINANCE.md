# DOMAIN_FINANCE.md

## Finance Domain for BuyOrBye Project
*Managing user financial data for intelligent purchase decisions - Following strict layered architecture*

---

## 🏗️ Architecture Alignment

### Layer Responsibilities (STRICT)
- **Handlers:** HTTP transport only, parse/validate DTOs, return JSON responses
- **Services:** Financial calculations, aggregations, business rules, use domain structs exclusively
- **Repositories:** Income/Expense/Loan persistence with GORM only, use model structs exclusively
- **No cross-layer violations allowed**

### Finance Data Flow
```
Request → Auth Middleware → Validation → Handler → Service → Repository → Database
         ↓                  ↓              ↓          ↓           ↓          ↓
    [JWT Check]        [DTO Validation] [IncomeDTO] [Income]  [IncomeModel] [SQL]
```

---

## 📁 Project Structure Integration

```
internal/
  database/
    migrations/
      003_create_incomes_table.sql
      004_create_expenses_table.sql
      005_create_loans_table.sql
  models/                        # GORM models (DB schema)
    income_model.go             # GORM Income model
    expense_model.go            # GORM Expense model
    loan_model.go               # GORM Loan model
  domain/                        # Business entities
    income.go                   # Income domain entity
    expense.go                  # Expense domain entity
    loan.go                     # Loan domain entity
    finance_summary.go          # Aggregated financial summary
  repositories/                  # GORM implementations
    income_repo.go              # Income repository with GORM
    expense_repo.go             # Expense repository with GORM
    loan_repo.go                # Loan repository with GORM
  services/                      # Business logic
    finance_service.go          # Financial calculations & aggregations
    budget_analyzer.go          # Budget analysis logic
    debt_calculator.go          # Debt-to-income calculations
  handlers/                      # HTTP transport
    finance_handler.go          # Finance CRUD endpoints
  types/
    finance_dto.go              # Finance-specific DTOs
  middleware/
    auth.go                     # Existing - ensures user context
tests/
  integration/
    finance_integration_test.go
  testutils/
    finance_fixtures.go         # Test data builders
```

---

## 💰 Domain Models (Service Layer)

```go
// internal/domain/income.go
type Income struct {
    ID        string
    UserID    string
    Source    string    // "Salary", "Freelance", "Investment", etc.
    Amount    float64
    Frequency string    // "Monthly", "Weekly", "One-time"
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
}

// internal/domain/expense.go
type Expense struct {
    ID        string
    UserID    string
    Category  string    // "Housing", "Food", "Transport", "Entertainment"
    Name      string    // Specific expense name
    Amount    float64
    Frequency string    // "Monthly", "Weekly", "Daily"
    IsFixed   bool      // Fixed vs Variable expense
    Priority  int       // 1=Essential, 2=Important, 3=Nice-to-have
    CreatedAt time.Time
    UpdatedAt time.Time
}

// internal/domain/loan.go
type Loan struct {
    ID               string
    UserID           string
    Lender           string
    Type             string    // "Mortgage", "Auto", "Personal", "Student"
    PrincipalAmount  float64
    RemainingBalance float64
    MonthlyPayment   float64
    InterestRate     float64
    EndDate          time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

// internal/domain/finance_summary.go
type FinanceSummary struct {
    UserID              string
    MonthlyIncome       float64
    MonthlyExpenses     float64
    MonthlyLoanPayments float64
    DisposableIncome    float64
    DebtToIncomeRatio   float64
    SavingsRate         float64
    FinancialHealth     string    // "Excellent", "Good", "Fair", "Poor"
    BudgetRemaining     float64
    UpdatedAt           time.Time
}
```

---

## 📋 Data Transfer Objects (Handler Layer)

```go
// internal/types/finance_dto.go

// Request DTOs
type AddIncomeDTO struct {
    Source    string  `json:"source" binding:"required,min=2,max=100"`
    Amount    float64 `json:"amount" binding:"required,gt=0"`
    Frequency string  `json:"frequency" binding:"required,oneof=monthly weekly daily one-time"`
}

type AddExpenseDTO struct {
    Category  string  `json:"category" binding:"required,oneof=housing food transport entertainment utilities other"`
    Name      string  `json:"name" binding:"required,min=2,max=100"`
    Amount    float64 `json:"amount" binding:"required,gt=0"`
    Frequency string  `json:"frequency" binding:"required,oneof=monthly weekly daily"`
    IsFixed   bool    `json:"is_fixed"`
    Priority  int     `json:"priority" binding:"required,min=1,max=3"`
}

type AddLoanDTO struct {
    Lender           string    `json:"lender" binding:"required,min=2,max=100"`
    Type             string    `json:"type" binding:"required,oneof=mortgage auto personal student"`
    PrincipalAmount  float64   `json:"principal_amount" binding:"required,gt=0"`
    RemainingBalance float64   `json:"remaining_balance" binding:"required,gt=0"`
    MonthlyPayment   float64   `json:"monthly_payment" binding:"required,gt=0"`
    InterestRate     float64   `json:"interest_rate" binding:"required,gte=0,lte=100"`
    EndDate          time.Time `json:"end_date" binding:"required"`
}

// Response DTOs
type FinanceSummaryDTO struct {
    MonthlyIncome       float64 `json:"monthly_income"`
    MonthlyExpenses     float64 `json:"monthly_expenses"`
    MonthlyLoanPayments float64 `json:"monthly_loan_payments"`
    DisposableIncome    float64 `json:"disposable_income"`
    DebtToIncomeRatio   float64 `json:"debt_to_income_ratio"`
    SavingsRate         float64 `json:"savings_rate"`
    FinancialHealth     string  `json:"financial_health"`
    BudgetRemaining     float64 `json:"budget_remaining"`
    CanAfford           float64 `json:"can_afford_up_to"` // Max purchase amount
}

// Conversion methods
func (dto AddIncomeDTO) ToDomain(userID string) domain.Income {
    return domain.Income{
        UserID:    userID,
        Source:    dto.Source,
        Amount:    dto.Amount,
        Frequency: dto.Frequency,
        IsActive:  true,
    }
}
```

---

## 💾 GORM Models (Repository Layer)

```go
// internal/models/income_model.go
type IncomeModel struct {
    gorm.Model
    UserID    uint    `gorm:"index;not null"`
    Source    string  `gorm:"not null"`
    Amount    float64 `gorm:"not null"`
    Frequency string  `gorm:"not null"`
    IsActive  bool    `gorm:"default:true"`
}

// internal/models/expense_model.go
type ExpenseModel struct {
    gorm.Model
    UserID    uint    `gorm:"index;not null"`
    Category  string  `gorm:"not null;index"`
    Name      string  `gorm:"not null"`
    Amount    float64 `gorm:"not null"`
    Frequency string  `gorm:"not null"`
    IsFixed   bool    `gorm:"default:false"`
    Priority  int     `gorm:"default:2"`
}

// internal/models/loan_model.go
type LoanModel struct {
    gorm.Model
    UserID           uint      `gorm:"index;not null"`
    Lender           string    `gorm:"not null"`
    Type             string    `gorm:"not null;index"`
    PrincipalAmount  float64   `gorm:"not null"`
    RemainingBalance float64   `gorm:"not null"`
    MonthlyPayment   float64   `gorm:"not null"`
    InterestRate     float64   `gorm:"not null"`
    EndDate          time.Time `gorm:"not null"`
}

// Conversion methods
func (m IncomeModel) ToDomain() domain.Income {
    return domain.Income{
        ID:        strconv.FormatUint(uint64(m.ID), 10),
        UserID:    strconv.FormatUint(uint64(m.UserID), 10),
        Source:    m.Source,
        Amount:    m.Amount,
        Frequency: m.Frequency,
        IsActive:  m.IsActive,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }
}
```

---

## 🧮 Service Layer Implementation

```go
// internal/services/finance_service.go
type FinanceService interface {
    // Income operations
    AddIncome(ctx context.Context, income *domain.Income) error
    GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error)
    UpdateIncome(ctx context.Context, income *domain.Income) error
    DeleteIncome(ctx context.Context, userID, incomeID string) error
    
    // Expense operations
    AddExpense(ctx context.Context, expense *domain.Expense) error
    GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error)
    CategorizeExpenses(ctx context.Context, userID string) (map[string]float64, error)
    
    // Loan operations
    AddLoan(ctx context.Context, loan *domain.Loan) error
    GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error)
    
    // Calculations & Analysis
    CalculateFinanceSummary(ctx context.Context, userID string) (*domain.FinanceSummary, error)
    CalculateDisposableIncome(ctx context.Context, userID string) (float64, error)
    CalculateDebtToIncomeRatio(ctx context.Context, userID string) (float64, error)
    EvaluateFinancialHealth(ctx context.Context, userID string) (string, error)
    GetMaxAffordableAmount(ctx context.Context, userID string) (float64, error)
}

// internal/services/budget_analyzer.go
type BudgetAnalyzer interface {
    AnalyzeBudget(summary *domain.FinanceSummary) *BudgetAnalysis
    GetSpendingInsights(expenses []domain.Expense) []SpendingInsight
    RecommendSavings(summary *domain.FinanceSummary) []SavingRecommendation
}

// Business Rules
const (
    HealthyDebtToIncomeRatio = 0.36  // 36% is considered healthy
    MinimumSavingsRate       = 0.20  // 20% savings rate target
    EmergencyFundMonths      = 6     // 6 months of expenses
)

// Financial Health Evaluation
func (s *financeService) EvaluateFinancialHealth(ctx context.Context, userID string) (string, error) {
    summary, err := s.CalculateFinanceSummary(ctx, userID)
    if err != nil {
        return "", err
    }
    
    switch {
    case summary.DebtToIncomeRatio > 0.50:
        return "Poor", nil
    case summary.DebtToIncomeRatio > 0.36:
        return "Fair", nil
    case summary.SavingsRate < 0.10:
        return "Fair", nil
    case summary.SavingsRate >= 0.20 && summary.DebtToIncomeRatio <= 0.28:
        return "Excellent", nil
    default:
        return "Good", nil
    }
}
```

---

## 🧪 TDD Test Structure

### Service Tests (Unit Tests with Mocks)
```go
// internal/services/finance_service_test.go
func TestFinanceService_CalculateDisposableIncome_Success(t *testing.T) {
    tests := []struct {
        name             string
        incomes          []domain.Income
        expenses         []domain.Expense
        loans            []domain.Loan
        expectedDisposable float64
        expectedError    error
    }{
        {
            name: "positive_disposable_income",
            incomes: []domain.Income{
                {Amount: 5000, Frequency: "monthly"},
            },
            expenses: []domain.Expense{
                {Amount: 2000, Frequency: "monthly"},
            },
            loans: []domain.Loan{
                {MonthlyPayment: 500},
            },
            expectedDisposable: 2500.00,
        },
        {
            name: "negative_disposable_income_warning",
            incomes: []domain.Income{
                {Amount: 3000, Frequency: "monthly"},
            },
            expenses: []domain.Expense{
                {Amount: 3500, Frequency: "monthly"},
            },
            expectedDisposable: -500.00,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

func TestFinanceService_EvaluateFinancialHealth_Scenarios(t *testing.T) {
    tests := []struct {
        name           string
        debtToIncome   float64
        savingsRate    float64
        expectedHealth string
    }{
        {"excellent_finances", 0.20, 0.25, "Excellent"},
        {"high_debt_ratio", 0.55, 0.10, "Poor"},
        {"good_with_moderate_debt", 0.30, 0.15, "Good"},
        {"low_savings", 0.25, 0.05, "Fair"},
    }
    // Implementation
}
```

### Repository Integration Tests
```go
// internal/repositories/finance_repo_test.go
func TestIncomeRepository_SaveAndRetrieve_Success(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    repo := repositories.NewIncomeRepository(db)
    
    income := domain.Income{
        UserID:    "1",
        Source:    "Salary",
        Amount:    5000.00,
        Frequency: "monthly",
    }
    
    // Act
    err := repo.Save(context.Background(), &income)
    assert.NoError(t, err)
    
    retrieved, err := repo.GetByUserID(context.Background(), "1")
    
    // Assert
    assert.NoError(t, err)
    assert.Len(t, retrieved, 1)
    assert.Equal(t, income.Amount, retrieved[0].Amount)
}
```

---

## 🚀 Claude Code Commands

### Slash Commands (.claude/commands/)

**`/finance-init`**
```markdown
Initialize finance domain following BuyOrBye architecture:
1. Create GORM models: IncomeModel, ExpenseModel, LoanModel
2. Create domain entities with validation methods
3. Create DTOs with binding validation
4. Define repository interfaces in services
5. Implement repositories with GORM
6. Create FinanceService with calculations
7. Create handlers with Gin/Echo
8. Generate database migrations
```

**`/finance-test-tdd`**
```markdown
Generate comprehensive finance tests:
1. Service tests for calculations (>90% coverage)
2. Test DisposableIncome with various scenarios
3. Test DebtToIncomeRatio edge cases
4. Test FinancialHealth evaluation rules
5. Repository integration tests (>85% coverage)
6. Handler tests with validation (>80% coverage)
Use table-driven tests for all scenarios
```

**`/finance-analyze`**
```markdown
Implement budget analysis features:
1. Spending pattern analysis
2. Category-wise expense breakdown
3. Savings recommendations
4. Purchase affordability calculator
5. Financial health scoring
```

---

## 🔄 Development Workflow

```bash
# Step 1: TDD - Write tests first
claude "/agent tdd-writer Create finance service tests for DisposableIncome and DebtToIncome calculations"

# Step 2: Create domain models
claude "Create finance domain models: Income, Expense, Loan, FinanceSummary with validation"

# Step 3: Implement repositories
claude "Create finance repositories with GORM for Income, Expense, Loan models"

# Step 4: Implement service
claude "Implement FinanceService with calculation methods following TDD tests"

# Step 5: Create handlers
claude "Create finance handlers with DTO validation for adding income, expenses, loans"

# Step 6: Integration
claude "Wire finance domain into main.go with auth middleware protecting all finance routes"
```

---

## 📊 Business Rules & Calculations

### Disposable Income Formula
```
Monthly Disposable = Total Monthly Income - Total Monthly Expenses - Total Monthly Loan Payments
```

### Debt-to-Income Ratio
```
DTI = (Total Monthly Debt Payments / Gross Monthly Income) × 100
```

### Financial Health Scoring
| Metric | Excellent | Good | Fair | Poor |
|--------|-----------|------|------|------|
| DTI Ratio | <28% | 28-36% | 36-50% | >50% |
| Savings Rate | >20% | 15-20% | 10-15% | <10% |
| Emergency Fund | 6+ months | 3-6 months | 1-3 months | <1 month |

### Purchase Affordability Rules
```go
// Maximum affordable purchase = 
// min(DisposableIncome * 3, EmergencyFundBalance * 0.5)

// For monthly payments (subscriptions, loans):
// MaxMonthlyPayment = DisposableIncome * 0.3
```

---

## 🔗 Integration with Decision Domain

The Finance domain provides critical data to the Decision domain:

```go
// Finance domain exposes to Decision domain:
type FinancialContext struct {
    UserID              string
    DisposableIncome    float64
    DebtToIncomeRatio   float64
    MonthlyBudget       float64
    CategorySpending    map[string]float64
    FinancialHealth     string
    MaxAffordableAmount float64
}

// Decision domain uses this for:
// - Purchase recommendations
// - Budget impact analysis
// - Financial risk assessment
```

---

## ✅ Implementation Checklist

- [ ] Domain models created (Income, Expense, Loan, FinanceSummary)
- [ ] GORM models with proper indexes
- [ ] DTOs with validation tags
- [ ] Repository interfaces defined
- [ ] GORM repositories implemented
- [ ] FinanceService with calculations
- [ ] BudgetAnalyzer service
- [ ] Handlers with DTO conversions
- [ ] Database migrations
- [ ] Unit tests (>90% service coverage)
- [ ] Integration tests (>85% repository coverage)
- [ ] Handler tests (>80% coverage)
- [ ] Auth middleware protecting routes
- [ ] Monthly/weekly/daily frequency handling
- [ ] Negative disposable income warnings
- [ ] Financial health evaluation
- [ ] No GORM imports in services
- [ ] No DTOs in services

---

## 🔍 Common Finance Domain Pitfalls

| Issue | Solution | Verification |
|-------|----------|--------------|
| Float precision errors | Use decimal library for money | Test with edge cases like 0.01 |
| Missing frequency normalization | Convert all to monthly for calculations | Test weekly/daily conversions |
| Negative income allowed | Validate amount > 0 in DTO | Test validation rules |
| Missing user scoping | Always filter by UserID | Test data isolation |
| Hardcoded categories | Use enum validation in DTOs | Test invalid categories rejected |
| Missing loan end date | Calculate from remaining balance | Test loan projections |

---

## 📚 References

- [BuyOrBye Architecture](./CLAUDE.md) - Project structure and rules
- [Go Validator](https://github.com/go-playground/validator) - DTO validation
- [Decimal Library](https://github.com/shopspring/decimal) - Precise financial calculations
- [Finance Domain Best Practices](https://martinfowler.com/articles/patterns-of-distributed-systems/)

---

*This finance domain provides the financial foundation for intelligent purchase decisions in the BuyOrBye platform.*