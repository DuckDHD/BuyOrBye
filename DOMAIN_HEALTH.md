# DOMAIN_HEALTH.md

## Health Domain for BuyOrBye Project
*Managing health context for informed purchase decisions - Following strict layered architecture*

---

## 🏗️ Architecture Alignment

### Layer Responsibilities (STRICT)
- **Handlers:** HTTP transport only, parse/validate DTOs, return JSON responses
- **Services:** Health risk calculations, medical expense analysis, insurance coverage logic, use domain structs exclusively
- **Repositories:** Health profile/conditions/expenses persistence with GORM only, use model structs exclusively
- **No cross-layer violations allowed**

### Health Data Flow
```
Request → Auth Middleware → Validation → Handler → Service → Repository → Database
         ↓                  ↓              ↓          ↓           ↓          ↓
    [JWT Check]        [DTO Validation] [HealthDTO] [Profile]  [ProfileModel] [SQL]
```

---

## 📁 Project Structure Integration

```
internal/
  database/
    migrations/
      006_create_health_profiles_table.sql
      007_create_medical_conditions_table.sql
      008_create_medical_expenses_table.sql
      009_create_insurance_policies_table.sql
  models/                             # GORM models (DB schema)
    health_profile_model.go          # GORM HealthProfile model
    medical_condition_model.go       # GORM MedicalCondition model
    medical_expense_model.go         # GORM MedicalExpense model
    insurance_policy_model.go        # GORM InsurancePolicy model
  domain/                             # Business entities
    health_profile.go                # Health profile domain entity
    medical_condition.go             # Medical condition domain entity
    medical_expense.go               # Medical expense domain entity
    insurance_policy.go              # Insurance policy domain entity
    health_summary.go                # Aggregated health summary
  repositories/                       # GORM implementations
    health_profile_repo.go           # Profile repository with GORM
    medical_condition_repo.go        # Condition repository with GORM
    medical_expense_repo.go          # Expense repository with GORM
    insurance_policy_repo.go         # Insurance repository with GORM
  services/                           # Business logic
    health_service.go                # Health management & calculations
    risk_calculator.go               # Health risk scoring
    medical_cost_analyzer.go         # Medical expense analysis
    insurance_evaluator.go           # Insurance coverage evaluation
  handlers/                           # HTTP transport
    health_handler.go                # Health CRUD endpoints
  types/
    health_dto.go                    # Health-specific DTOs
  middleware/
    auth.go                          # Existing - ensures user context
tests/
  integration/
    health_integration_test.go
  testutils/
    health_fixtures.go               # Test data builders
```

---

## 🏥 Domain Models (Service Layer)

```go
// internal/domain/health_profile.go
type HealthProfile struct {
    ID                 string
    UserID             string
    Age                int
    Gender             string    // "male", "female", "other"
    Height             float64   // in cm
    Weight             float64   // in kg
    BMI                float64   // calculated
    FamilySize         int       // household members
    HasChronicConditions bool    // quick flag
    EmergencyFundHealth float64  // health-specific emergency fund
    CreatedAt          time.Time
    UpdatedAt          time.Time
}

// internal/domain/medical_condition.go
type MedicalCondition struct {
    ID                string
    UserID            string
    ProfileID         string
    Name              string    // standardized condition name
    Category          string    // "chronic", "acute", "mental_health", "preventive"
    Severity          string    // "mild", "moderate", "severe", "critical"
    DiagnosedDate     time.Time
    IsActive          bool
    RequiresMedication bool
    MonthlyMedCost    float64   // estimated monthly medication cost
    RiskFactor        float64   // 0.0 to 1.0 risk multiplier
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

// internal/domain/medical_expense.go
type MedicalExpense struct {
    ID               string
    UserID           string
    ProfileID        string
    Amount           float64
    Category         string    // "doctor_visit", "medication", "hospital", "lab_test", "therapy", "equipment"
    Description      string
    IsRecurring      bool
    Frequency        string    // "monthly", "quarterly", "annually", "one_time"
    IsCovered        bool      // covered by insurance
    InsurancePayment float64   // amount paid by insurance
    OutOfPocket      float64   // actual user payment
    Date             time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

// internal/domain/insurance_policy.go
type InsurancePolicy struct {
    ID                  string
    UserID              string
    ProfileID           string
    Provider            string
    PolicyNumber        string
    Type                string    // "health", "dental", "vision", "comprehensive"
    MonthlyPremium      float64
    Deductible          float64
    DeductibleMet       float64   // amount already paid toward deductible
    OutOfPocketMax      float64
    OutOfPocketCurrent  float64   // current OOP expenses
    CoveragePercentage  float64   // after deductible (e.g., 80%)
    StartDate           time.Time
    EndDate             time.Time
    IsActive            bool
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

// internal/domain/health_summary.go
type HealthSummary struct {
    UserID                    string
    HealthRiskScore           int       // 0-100 (0=excellent, 100=critical)
    HealthRiskLevel           string    // "low", "moderate", "high", "critical"
    MonthlyMedicalExpenses    float64
    MonthlyInsurancePremiums  float64
    AnnualDeductibleRemaining float64
    OutOfPocketRemaining      float64
    TotalHealthCosts          float64   // premiums + out-of-pocket
    CoverageGapRisk           float64   // uncovered potential expenses
    RecommendedEmergencyFund  float64   // based on health risks
    FinancialVulnerability    string    // "secure", "moderate", "vulnerable", "critical"
    PriorityAdjustment        float64   // multiplier for purchase decisions
    UpdatedAt                 time.Time
}
```

---

## 📋 Data Transfer Objects (Handler Layer)

```go
// internal/types/health_dto.go

// Request DTOs
type CreateHealthProfileDTO struct {
    Age        int     `json:"age" binding:"required,min=0,max=150"`
    Gender     string  `json:"gender" binding:"required,oneof=male female other"`
    Height     float64 `json:"height" binding:"required,gt=0,max=300"`
    Weight     float64 `json:"weight" binding:"required,gt=0,max=500"`
    FamilySize int     `json:"family_size" binding:"required,min=1,max=20"`
}

type AddMedicalConditionDTO struct {
    Name               string    `json:"name" binding:"required,min=2,max=100"`
    Category           string    `json:"category" binding:"required,oneof=chronic acute mental_health preventive"`
    Severity           string    `json:"severity" binding:"required,oneof=mild moderate severe critical"`
    DiagnosedDate      time.Time `json:"diagnosed_date" binding:"required"`
    RequiresMedication bool      `json:"requires_medication"`
    MonthlyMedCost     float64   `json:"monthly_med_cost" binding:"min=0"`
}

type AddMedicalExpenseDTO struct {
    Amount           float64   `json:"amount" binding:"required,gt=0"`
    Category         string    `json:"category" binding:"required,oneof=doctor_visit medication hospital lab_test therapy equipment"`
    Description      string    `json:"description" binding:"max=500"`
    IsRecurring      bool      `json:"is_recurring"`
    Frequency        string    `json:"frequency" binding:"required_if=IsRecurring true,omitempty,oneof=monthly quarterly annually"`
    IsCovered        bool      `json:"is_covered"`
    InsurancePayment float64   `json:"insurance_payment" binding:"min=0"`
    Date             time.Time `json:"date" binding:"required"`
}

type AddInsurancePolicyDTO struct {
    Provider           string    `json:"provider" binding:"required,min=2,max=100"`
    PolicyNumber       string    `json:"policy_number" binding:"required,min=5,max=50"`
    Type               string    `json:"type" binding:"required,oneof=health dental vision comprehensive"`
    MonthlyPremium     float64   `json:"monthly_premium" binding:"required,gt=0"`
    Deductible         float64   `json:"deductible" binding:"required,gte=0"`
    OutOfPocketMax     float64   `json:"out_of_pocket_max" binding:"required,gt=0"`
    CoveragePercentage float64   `json:"coverage_percentage" binding:"required,min=0,max=100"`
    StartDate          time.Time `json:"start_date" binding:"required"`
    EndDate            time.Time `json:"end_date" binding:"required,gtfield=StartDate"`
}

// Response DTOs
type HealthSummaryDTO struct {
    HealthRiskScore           int     `json:"health_risk_score"`
    HealthRiskLevel           string  `json:"health_risk_level"`
    MonthlyMedicalExpenses    float64 `json:"monthly_medical_expenses"`
    MonthlyInsurancePremiums  float64 `json:"monthly_insurance_premiums"`
    TotalHealthCosts          float64 `json:"total_health_costs"`
    CoverageGapRisk           float64 `json:"coverage_gap_risk"`
    RecommendedEmergencyFund  float64 `json:"recommended_emergency_fund"`
    FinancialVulnerability    string  `json:"financial_vulnerability"`
    HealthImpactOnPurchases   string  `json:"health_impact_on_purchases"`
}

// Conversion methods
func (dto CreateHealthProfileDTO) ToDomain(userID string) domain.HealthProfile {
    bmi := dto.Weight / ((dto.Height / 100) * (dto.Height / 100))
    return domain.HealthProfile{
        UserID:     userID,
        Age:        dto.Age,
        Gender:     dto.Gender,
        Height:     dto.Height,
        Weight:     dto.Weight,
        BMI:        bmi,
        FamilySize: dto.FamilySize,
    }
}
```

---

## 💾 GORM Models (Repository Layer)

```go
// internal/models/health_profile_model.go
type HealthProfileModel struct {
    gorm.Model
    UserID              uint    `gorm:"uniqueIndex;not null"`
    Age                 int     `gorm:"not null"`
    Gender              string  `gorm:"not null"`
    Height              float64 `gorm:"not null"`
    Weight              float64 `gorm:"not null"`
    BMI                 float64 `gorm:"not null"`
    FamilySize          int     `gorm:"not null"`
    HasChronicConditions bool   `gorm:"default:false"`
    EmergencyFundHealth float64 `gorm:"default:0"`
}

// internal/models/medical_condition_model.go
type MedicalConditionModel struct {
    gorm.Model
    UserID             uint      `gorm:"index;not null"`
    ProfileID          uint      `gorm:"index;not null"`
    Name               string    `gorm:"not null;index"`
    Category           string    `gorm:"not null;index"`
    Severity           string    `gorm:"not null"`
    DiagnosedDate      time.Time `gorm:"not null"`
    IsActive           bool      `gorm:"default:true"`
    RequiresMedication bool      `gorm:"default:false"`
    MonthlyMedCost     float64   `gorm:"default:0"`
    RiskFactor         float64   `gorm:"default:0.1"`
}

// internal/models/medical_expense_model.go
type MedicalExpenseModel struct {
    gorm.Model
    UserID           uint      `gorm:"index;not null"`
    ProfileID        uint      `gorm:"index;not null"`
    Amount           float64   `gorm:"not null"`
    Category         string    `gorm:"not null;index"`
    Description      string
    IsRecurring      bool      `gorm:"default:false"`
    Frequency        string
    IsCovered        bool      `gorm:"default:false"`
    InsurancePayment float64   `gorm:"default:0"`
    OutOfPocket      float64   `gorm:"not null"`
    Date             time.Time `gorm:"not null;index"`
}

// internal/models/insurance_policy_model.go
type InsurancePolicyModel struct {
    gorm.Model
    UserID              uint      `gorm:"index;not null"`
    ProfileID           uint      `gorm:"index;not null"`
    Provider            string    `gorm:"not null"`
    PolicyNumber        string    `gorm:"uniqueIndex;not null"`
    Type                string    `gorm:"not null;index"`
    MonthlyPremium      float64   `gorm:"not null"`
    Deductible          float64   `gorm:"not null"`
    DeductibleMet       float64   `gorm:"default:0"`
    OutOfPocketMax      float64   `gorm:"not null"`
    OutOfPocketCurrent  float64   `gorm:"default:0"`
    CoveragePercentage  float64   `gorm:"not null"`
    StartDate           time.Time `gorm:"not null"`
    EndDate             time.Time `gorm:"not null"`
    IsActive            bool      `gorm:"default:true"`
}
```

---

## 🧮 Service Layer Implementation

```go
// internal/services/health_service.go
type HealthService interface {
    // Profile operations
    CreateProfile(ctx context.Context, profile *domain.HealthProfile) error
    GetProfile(ctx context.Context, userID string) (*domain.HealthProfile, error)
    UpdateProfile(ctx context.Context, profile *domain.HealthProfile) error
    
    // Medical conditions
    AddCondition(ctx context.Context, condition *domain.MedicalCondition) error
    GetConditions(ctx context.Context, userID string) ([]domain.MedicalCondition, error)
    UpdateCondition(ctx context.Context, condition *domain.MedicalCondition) error
    RemoveCondition(ctx context.Context, userID, conditionID string) error
    
    // Medical expenses
    AddExpense(ctx context.Context, expense *domain.MedicalExpense) error
    GetExpenses(ctx context.Context, userID string, timeRange TimeRange) ([]domain.MedicalExpense, error)
    GetRecurringExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error)
    
    // Insurance policies
    AddInsurancePolicy(ctx context.Context, policy *domain.InsurancePolicy) error
    GetActivePolicies(ctx context.Context, userID string) ([]domain.InsurancePolicy, error)
    UpdateDeductibleProgress(ctx context.Context, policyID string, amount float64) error
    
    // Calculations & Analysis
    CalculateHealthSummary(ctx context.Context, userID string) (*domain.HealthSummary, error)
    GetHealthContext(ctx context.Context, userID string) (*HealthContext, error)
}

// internal/services/risk_calculator.go
type RiskCalculator interface {
    CalculateHealthRiskScore(profile *domain.HealthProfile, conditions []domain.MedicalCondition) int
    AssessFinancialVulnerability(healthCosts, income float64) string
    RecommendEmergencyFund(riskScore int, monthlyExpenses float64) float64
}

// Risk calculation logic
func (r *riskCalculator) CalculateHealthRiskScore(profile *domain.HealthProfile, conditions []domain.MedicalCondition) int {
    baseScore := 0
    
    // Age factor (0-20 points)
    switch {
    case profile.Age < 30:
        baseScore += 0
    case profile.Age < 40:
        baseScore += 5
    case profile.Age < 50:
        baseScore += 10
    case profile.Age < 60:
        baseScore += 15
    default:
        baseScore += 20
    }
    
    // BMI factor (0-15 points)
    switch {
    case profile.BMI < 18.5 || profile.BMI > 30:
        baseScore += 15
    case profile.BMI > 25:
        baseScore += 8
    default:
        baseScore += 0
    }
    
    // Conditions factor (0-50 points)
    for _, condition := range conditions {
        if !condition.IsActive {
            continue
        }
        switch condition.Severity {
        case "critical":
            baseScore += 15
        case "severe":
            baseScore += 10
        case "moderate":
            baseScore += 5
        case "mild":
            baseScore += 2
        }
    }
    
    // Family size factor (0-15 points)
    if profile.FamilySize > 4 {
        baseScore += 10
    } else if profile.FamilySize > 2 {
        baseScore += 5
    }
    
    // Cap at 100
    if baseScore > 100 {
        baseScore = 100
    }
    
    return baseScore
}

// internal/services/medical_cost_analyzer.go
type MedicalCostAnalyzer interface {
    CalculateMonthlyAverage(expenses []domain.MedicalExpense) float64
    ProjectAnnualCosts(expenses []domain.MedicalExpense, conditions []domain.MedicalCondition) float64
    IdentifyCostReductionOpportunities(expenses []domain.MedicalExpense) []CostReduction
}

// internal/services/insurance_evaluator.go
type InsuranceEvaluator interface {
    CalculateCoverage(expense *domain.MedicalExpense, policies []domain.InsurancePolicy) CoverageResult
    EvaluateCoverageGaps(conditions []domain.MedicalCondition, policies []domain.InsurancePolicy) []CoverageGap
    RecommendPolicyAdjustments(usage []domain.MedicalExpense, policies []domain.InsurancePolicy) []PolicyRecommendation
}
```

---

## 🧪 TDD Test Structure

### Service Tests (Unit Tests with Mocks)
```go
// internal/services/health_service_test.go
func TestHealthService_CalculateHealthRiskScore_Scenarios(t *testing.T) {
    tests := []struct {
        name           string
        profile        domain.HealthProfile
        conditions     []domain.MedicalCondition
        expectedScore  int
        expectedLevel  string
    }{
        {
            name: "young_healthy_individual",
            profile: domain.HealthProfile{
                Age: 25, BMI: 22.5, FamilySize: 1,
            },
            conditions:    []domain.MedicalCondition{},
            expectedScore: 0,
            expectedLevel: "low",
        },
        {
            name: "middle_aged_with_chronic_conditions",
            profile: domain.HealthProfile{
                Age: 45, BMI: 28, FamilySize: 4,
            },
            conditions: []domain.MedicalCondition{
                {Name: "Diabetes", Severity: "moderate", IsActive: true},
                {Name: "Hypertension", Severity: "mild", IsActive: true},
            },
            expectedScore: 32,
            expectedLevel: "moderate",
        },
        {
            name: "elderly_multiple_severe_conditions",
            profile: domain.HealthProfile{
                Age: 65, BMI: 32, FamilySize: 2,
            },
            conditions: []domain.MedicalCondition{
                {Name: "Heart Disease", Severity: "severe", IsActive: true},
                {Name: "Diabetes", Severity: "severe", IsActive: true},
            },
            expectedScore: 65,
            expectedLevel: "high",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

func TestHealthService_CalculateMonthlyMedicalCosts_WithInsurance(t *testing.T) {
    // Test medical expense calculations with insurance coverage
    // Include deductible logic, coverage percentages, out-of-pocket maximums
}
```

### Repository Integration Tests
```go
// internal/repositories/health_profile_repo_test.go
func TestHealthProfileRepository_OneProfilePerUser(t *testing.T) {
    // Test that only one profile can exist per user
    // Test update replaces existing profile
}

func TestMedicalConditionRepository_CascadeDelete(t *testing.T) {
    // Test that deleting profile removes associated conditions
}
```

---

## 🚀 Claude Code Commands

### Slash Commands (.claude/commands/)

**`/health-init`**
```markdown
Initialize health domain following BuyOrBye architecture:
1. Create GORM models: HealthProfileModel, MedicalConditionModel, MedicalExpenseModel, InsurancePolicyModel
2. Create domain entities with validation
3. Create DTOs with comprehensive validation
4. Define repository interfaces in services
5. Implement repositories with GORM
6. Create HealthService with risk calculations
7. Create RiskCalculator, MedicalCostAnalyzer, InsuranceEvaluator
8. Create handlers with Gin/Echo
9. Generate database migrations with foreign keys
```

**`/health-test-tdd`**
```markdown
Generate comprehensive health domain tests:
1. Service tests for risk scoring (>90% coverage)
2. Test health risk calculation with various scenarios
3. Test medical cost aggregation with recurring expenses
4. Test insurance coverage calculations
5. Test financial vulnerability assessment
6. Repository tests with cascade operations (>85%)
7. Handler tests with validation (>80%)
```

**`/health-analyze`**
```markdown
Implement health analysis features:
1. Risk scoring algorithm with multiple factors
2. Medical expense trend analysis
3. Insurance coverage gap identification
4. Cost reduction recommendations
5. Emergency fund calculation based on health
```

---

## 🔄 Development Workflow

```bash
# Step 1: TDD - Write tests first
claude "/agent tdd-writer Create health service tests for risk scoring and medical cost calculations"

# Step 2: Create domain models
claude "Create health domain models: HealthProfile, MedicalCondition, MedicalExpense, InsurancePolicy"

# Step 3: Implement repositories
claude "Create health repositories with GORM, ensure one profile per user constraint"

# Step 4: Implement services
claude "Implement HealthService with risk calculator and cost analyzer"

# Step 5: Create handlers
claude "Create health handlers with comprehensive validation"

# Step 6: Wire into main
claude "Integrate health domain with auth middleware, add routes"
```

---

## 📊 Business Rules & Calculations

### Health Risk Scoring (0-100)
```
Components:
- Age Factor: 0-20 points
- BMI Factor: 0-15 points  
- Medical Conditions: 0-50 points
- Family Size: 0-15 points

Risk Levels:
- 0-25: Low Risk
- 26-50: Moderate Risk
- 51-75: High Risk
- 76-100: Critical Risk
```

### Financial Vulnerability Assessment
```
Vulnerability = (Monthly Health Costs / Monthly Income) × 100

Levels:
- <5%: Secure
- 5-10%: Moderate
- 10-20%: Vulnerable
- >20%: Critical
```

### Emergency Fund Recommendation
```
Base Emergency Fund = 6 months × monthly expenses

Health Adjusted = Base × (1 + Risk Score/100)
Example: Risk Score 50 = 9 months emergency fund
```

---

## 🔗 Integration with Decision Domain

```go
// Health domain provides to Decision domain:
type HealthContext struct {
    UserID                 string
    HealthRiskScore        int
    MonthlyHealthCosts     float64
    InsuranceCoverage      float64
    UncoveredRiskExposure  float64
    EmergencyFundNeeded    float64
    PurchasePriorityAdjustment float64 // Multiplier for health-related purchases
}

// Decision domain uses this for:
// - Prioritizing health-related purchases
// - Adjusting budget for medical expenses
// - Evaluating financial resilience
// - Risk-adjusted purchase recommendations
```

---

## ✅ Implementation Checklist

- [ ] Domain models (HealthProfile, MedicalCondition, MedicalExpense, InsurancePolicy)
- [ ] GORM models with proper constraints
- [ ] DTOs with comprehensive validation
- [ ] Repository interfaces and implementations
- [ ] HealthService with CRUD operations
- [ ] RiskCalculator service
- [ ] MedicalCostAnalyzer service
- [ ] InsuranceEvaluator service
- [ ] Handlers with auth protection
- [ ] Database migrations with indexes
- [ ] Unit tests (>90% service coverage)
- [ ] Integration tests (>85% repository coverage)
- [ ] Handler tests (>80% coverage)
- [ ] One profile per user constraint
- [ ] Cascade delete for related records
- [ ] Privacy protection for sensitive data
- [ ] No GORM in services
- [ ] No DTOs in services

---

## 🔍 Common Health Domain Pitfalls

| Issue | Solution | Verification |
|-------|----------|--------------|
| Multiple profiles per user | Unique constraint on UserID | Test duplicate creation fails |
| Sensitive data in logs | Exclude conditions from logs | Audit log output |
| Invalid BMI calculations | Validate height/weight ranges | Test edge cases |
| Insurance overlap | Check policy dates don't overlap | Test concurrent policies |
| Missing cascade deletes | Set up foreign keys properly | Test profile deletion |
| Incorrect risk scoring | Unit test all score scenarios | Test boundary values |
| Privacy violations | Strict UserID filtering | Test cross-user access |

---

## 🔒 Privacy & Security Considerations

```go
// Sensitive data handling
- Never log medical condition details
- Use encryption for condition names in DB (optional)
- Implement audit trail for health data access
- Ensure HIPAA-like compliance if needed
- Strict access control via UserID
- Consider separate encryption key for health data
```

---

## 📚 References

- [BuyOrBye Architecture](./CLAUDE.md) - Project structure and rules
- [OWASP Healthcare](https://owasp.org/www-project-healthcare) - Security best practices
- [Go Validator](https://github.com/go-playground/validator) - DTO validation
- [Health Risk Scoring](https://www.cdc.gov/chronicdisease/index.htm) - CDC guidelines

---

*This health domain provides critical health context for informed financial decisions in the BuyOrBye platform.*