# DOMAIN_DECISION.md

## Decision Domain for BuyOrBye Project
*AI-driven purchase recommendations aggregating all user context - Following strict layered architecture*

---

## 🏗️ Architecture Alignment

### Layer Responsibilities (STRICT)
- **Handlers:** HTTP transport only, accept purchase intent DTOs, stream AI responses
- **Services:** Aggregate multi-domain data, build AI prompts, interpret responses, use domain structs exclusively
- **Repositories:** Decision history persistence with GORM only, use model structs exclusively
- **No cross-layer violations allowed**

### Decision Data Flow
```
Request → Auth Middleware → Validation → Handler → Service → AI Provider → Response
         ↓                  ↓              ↓          ↓         ↓           ↓
    [JWT Check]        [Intent DTO]   [Aggregate] [Prompt]  [AI Call]  [Decision]
                                           ↓
                                    [Finance + Health + Transport]
```

---

## 📁 Project Structure Integration

```
internal/
  database/
    migration.go                      # GORM AutoMigrate setup
  models/                             # GORM models (DB schema)
    decision_record_model.go          # GORM DecisionRecord model
    ai_prompt_log_model.go            # GORM AI interaction logs
  domain/                             # Business entities
    purchase_intent.go                # Purchase intent domain entity
    decision_outcome.go               # Decision outcome domain entity
    decision_context.go               # Aggregated context from all domains
    ai_prompt.go                      # Structured AI prompt
    decision_factors.go               # Factors influencing decision
  repositories/                       # GORM implementations
    decision_repo.go                  # Decision history repository
    prompt_log_repo.go                # AI prompt logging repository
  services/                           # Business logic
    decision_service.go               # Main decision orchestration
    context_aggregator.go             # Multi-domain data aggregation
    prompt_builder.go                 # AI prompt construction
    ai_client.go                      # AI provider interface & implementations
    decision_interpreter.go           # AI response parsing & normalization
    recommendation_engine.go          # Decision logic & rules
  handlers/                           # HTTP transport
    decision_handler.go               # Decision endpoints
  types/
    decision_dto.go                   # Decision-specific DTOs
  clients/
    claude_client.go                  # Claude AI implementation
    openai_client.go                  # OpenAI implementation
    ollama_client.go                  # Local LLaMA implementation
tests/
  integration/
    decision_flow_test.go
  testutils/
    ai_mock.go                        # Mock AI responses
```

---

## 🧠 Domain Models (Service Layer)

```go
// internal/domain/purchase_intent.go
type PurchaseIntent struct {
    ID          string
    UserID      string
    ItemName    string
    ItemCost    float64
    Category    string    // "electronics", "clothing", "food", "transport", "health", "entertainment", "other"
    Urgency     string    // "low", "medium", "high", "critical"
    Frequency   string    // "one_time", "monthly", "yearly"
    Purpose     string    // User's reason for purchase
    Alternative string    // Cheaper alternative considered
    CreatedAt   time.Time
}

// internal/domain/decision_outcome.go
type DecisionOutcome struct {
    ID              string
    UserID          string
    IntentID        string
    Decision        string    // "BUY", "WAIT", "BYE"
    Confidence      float64   // 0.0 to 1.0
    PrimaryReason   string    // Main reason for decision
    Factors         []DecisionFactor
    Recommendations []string  // Actionable advice
    WaitPeriod      int       // Days to wait (if WAIT)
    MaxBudget       float64   // Recommended max spend
    CreatedAt       time.Time
    ProcessingTime  int64     // Milliseconds
}

// internal/domain/decision_factors.go
type DecisionFactor struct {
    Category    string    // "financial", "health", "practical", "timing"
    Impact      string    // "positive", "negative", "neutral"
    Weight      float64   // 0.0 to 1.0 importance
    Description string
}

// internal/domain/decision_context.go
type DecisionContext struct {
    UserID            string
    FinancialContext  FinancialSnapshot
    HealthContext     HealthSnapshot
    TransportContext  TransportSnapshot  // If transport domain exists
    PurchaseHistory   []PastDecision
    CurrentDate       time.Time
}

type FinancialSnapshot struct {
    MonthlyIncome        float64
    MonthlyExpenses      float64
    DisposableIncome     float64
    DebtToIncomeRatio    float64
    EmergencyFundMonths  float64
    SavingsRate          float64
    FinancialHealth      string
    BudgetRemaining      float64
}

type HealthSnapshot struct {
    HealthRiskScore      int
    MonthlyHealthCosts   float64
    InsuranceCoverage    float64
    FinancialVulnerability string
    HasCriticalConditions bool
    EmergencyFundNeeded  float64
}

type TransportSnapshot struct {
    HasVehicle           bool
    MonthlyTransportCost float64
    PublicTransitAccess  bool
    CommuteDistance      float64
}

type PastDecision struct {
    ItemName    string
    ItemCost    float64
    Decision    string
    DaysAgo     int
    Category    string
}

// internal/domain/ai_prompt.go
type AIPrompt struct {
    SystemContext   string
    UserContext     string
    PurchaseDetails string
    DecisionCriteria string
    ResponseFormat  string
    MaxTokens       int
    Temperature     float64
}

type AIResponse struct {
    RawResponse string
    Decision    string
    Confidence  float64
    Reasoning   string
    Factors     []string
    Suggestions []string
    TokensUsed  int
}
```

---

## 📋 Data Transfer Objects (Handler Layer)

```go
// internal/types/decision_dto.go

// Request DTOs
type PurchaseIntentDTO struct {
    ItemName    string  `json:"item_name" binding:"required,min=2,max=200"`
    ItemCost    float64 `json:"item_cost" binding:"required,gt=0,max=1000000"`
    Category    string  `json:"category" binding:"required,oneof=electronics clothing food transport health entertainment other"`
    Urgency     string  `json:"urgency" binding:"required,oneof=low medium high critical"`
    Frequency   string  `json:"frequency" binding:"required,oneof=one_time monthly yearly"`
    Purpose     string  `json:"purpose" binding:"max=500"`
    Alternative string  `json:"alternative" binding:"max=200"`
}

type GetDecisionDTO struct {
    IntentID string `json:"intent_id" binding:"required,uuid"`
    Stream   bool   `json:"stream"` // Stream AI response
}

// Response DTOs
type DecisionResponseDTO struct {
    Decision        string                 `json:"decision"`
    Confidence      float64                `json:"confidence"`
    PrimaryReason   string                 `json:"primary_reason"`
    Factors         []DecisionFactorDTO    `json:"factors"`
    Recommendations []string               `json:"recommendations"`
    WaitPeriod      int                    `json:"wait_period_days,omitempty"`
    MaxBudget       float64                `json:"max_budget,omitempty"`
    ProcessingTime  int64                  `json:"processing_time_ms"`
}

type DecisionFactorDTO struct {
    Category    string  `json:"category"`
    Impact      string  `json:"impact"`
    Weight      float64 `json:"weight"`
    Description string  `json:"description"`
}

type DecisionHistoryDTO struct {
    Decisions []PastDecisionDTO `json:"decisions"`
    Stats     DecisionStatsDTO  `json:"stats"`
}

type PastDecisionDTO struct {
    ID          string    `json:"id"`
    ItemName    string    `json:"item_name"`
    ItemCost    float64   `json:"item_cost"`
    Category    string    `json:"category"`
    Decision    string    `json:"decision"`
    Confidence  float64   `json:"confidence"`
    Reason      string    `json:"reason"`
    CreatedAt   time.Time `json:"created_at"`
}

type DecisionStatsDTO struct {
    TotalDecisions int     `json:"total_decisions"`
    BuyCount       int     `json:"buy_count"`
    WaitCount      int     `json:"wait_count"`
    ByeCount       int     `json:"bye_count"`
    TotalSaved     float64 `json:"total_saved"`
    AvgConfidence  float64 `json:"avg_confidence"`
}

// Conversion methods
func (dto PurchaseIntentDTO) ToDomain(userID string) domain.PurchaseIntent {
    return domain.PurchaseIntent{
        UserID:      userID,
        ItemName:    dto.ItemName,
        ItemCost:    dto.ItemCost,
        Category:    dto.Category,
        Urgency:     dto.Urgency,
        Frequency:   dto.Frequency,
        Purpose:     dto.Purpose,
        Alternative: dto.Alternative,
    }
}
```

---

## 💾 GORM Models (Repository Layer)

```go
// internal/models/decision_record_model.go
type DecisionRecordModel struct {
    gorm.Model
    UserID          uint    `gorm:"index;not null"`
    ItemName        string  `gorm:"not null"`
    ItemCost        float64 `gorm:"not null"`
    Category        string  `gorm:"index;not null"`
    Urgency         string  `gorm:"not null"`
    Frequency       string  `gorm:"not null"`
    Purpose         string  `gorm:"type:text"`
    Decision        string  `gorm:"index;not null"` // BUY, WAIT, BYE
    Confidence      float64 `gorm:"not null"`
    PrimaryReason   string  `gorm:"type:text"`
    Recommendations string  `gorm:"type:text"` // JSON array
    WaitPeriod      int     `gorm:"default:0"`
    MaxBudget       float64 `gorm:"default:0"`
    ProcessingTime  int64   `gorm:"not null"`
}

// internal/models/ai_prompt_log_model.go
type AIPromptLogModel struct {
    gorm.Model
    UserID         uint   `gorm:"index;not null"`
    DecisionID     uint   `gorm:"index"`
    PromptHash     string `gorm:"index"` // For deduplication
    SystemPrompt   string `gorm:"type:text"`
    UserPrompt     string `gorm:"type:text"`
    AIProvider     string `gorm:"not null"` // claude, openai, ollama
    Model          string `gorm:"not null"` // specific model version
    Response       string `gorm:"type:text"`
    TokensInput    int    `gorm:"default:0"`
    TokensOutput   int    `gorm:"default:0"`
    ResponseTimeMs int64  `gorm:"not null"`
    Success        bool   `gorm:"default:true"`
    ErrorMessage   string `gorm:"type:text"`
}

// internal/database/migration.go - Add to existing
func RunDecisionMigrations(db *gorm.DB) error {
    // AutoMigrate decision domain models
    err := db.AutoMigrate(
        &models.DecisionRecordModel{},
        &models.AIPromptLogModel{},
    )
    if err != nil {
        return fmt.Errorf("failed to migrate decision models: %w", err)
    }
    
    // Add composite indexes for performance
    db.Exec("CREATE INDEX IF NOT EXISTS idx_decisions_user_created ON decision_record_models(user_id, created_at DESC)")
    db.Exec("CREATE INDEX IF NOT EXISTS idx_decisions_user_category ON decision_record_models(user_id, category)")
    
    return nil
}
```

---

## 🧮 Service Layer Implementation

```go
// internal/services/decision_service.go
type DecisionService interface {
    // Main decision flow
    MakeDecision(ctx context.Context, intent *domain.PurchaseIntent) (*domain.DecisionOutcome, error)
    GetDecisionHistory(ctx context.Context, userID string, limit int) ([]domain.DecisionOutcome, error)
    GetDecisionStats(ctx context.Context, userID string) (*domain.DecisionStats, error)
    
    // Analysis
    AnalyzePurchasePattern(ctx context.Context, userID string) (*domain.PurchasePattern, error)
    GetSimilarDecisions(ctx context.Context, userID string, category string) ([]domain.DecisionOutcome, error)
}

// internal/services/context_aggregator.go
type ContextAggregator interface {
    AggregateUserContext(ctx context.Context, userID string) (*domain.DecisionContext, error)
    GetFinancialSnapshot(ctx context.Context, userID string) (*domain.FinancialSnapshot, error)
    GetHealthSnapshot(ctx context.Context, userID string) (*domain.HealthSnapshot, error)
    GetTransportSnapshot(ctx context.Context, userID string) (*domain.TransportSnapshot, error)
    GetRecentDecisions(ctx context.Context, userID string, days int) ([]domain.PastDecision, error)
}

// internal/services/prompt_builder.go
type PromptBuilder interface {
    BuildDecisionPrompt(intent *domain.PurchaseIntent, context *domain.DecisionContext) (*domain.AIPrompt, error)
    BuildSystemPrompt() string
    BuildUserContext(context *domain.DecisionContext) string
    BuildDecisionCriteria() string
    BuildResponseFormat() string
}

// Prompt template example
func (pb *promptBuilder) BuildSystemPrompt() string {
    return `You are a financial advisor AI for the BuyOrBye app. Your role is to help users make informed purchase decisions based on their complete financial, health, and life context.

Analyze the user's situation and provide one of three recommendations:
- BUY: The purchase is financially sound and won't impact stability
- WAIT: The purchase should be delayed for better timing or planning
- BYE: The purchase is not recommended given current circumstances

Consider these factors:
1. Financial stability (income, expenses, savings, debt)
2. Health risks and medical expenses
3. Emergency fund adequacy
4. Purchase necessity vs want
5. Long-term financial goals
6. Current economic conditions

Provide clear reasoning and actionable recommendations.`
}

// internal/services/ai_client.go
type AIClient interface {
    SendPrompt(ctx context.Context, prompt *domain.AIPrompt) (*domain.AIResponse, error)
    GetModel() string
    GetProvider() string
    EstimateTokens(prompt string) int
}

// internal/services/decision_interpreter.go
type DecisionInterpreter interface {
    ParseAIResponse(response *domain.AIResponse) (*domain.DecisionOutcome, error)
    ExtractDecision(text string) string
    ExtractConfidence(text string) float64
    ExtractFactors(text string) []domain.DecisionFactor
    ExtractRecommendations(text string) []string
    ValidateDecision(decision string) error
}

// internal/services/recommendation_engine.go
type RecommendationEngine interface {
    // Fallback rules if AI fails
    ApplyBusinessRules(intent *domain.PurchaseIntent, context *domain.DecisionContext) *domain.DecisionOutcome
    CalculateAffordability(cost float64, disposableIncome float64) bool
    AssessUrgency(urgency string, category string) int
    DetermineWaitPeriod(context *domain.DecisionContext) int
    SuggestAlternatives(intent *domain.PurchaseIntent) []string
}

// Business Rules Implementation
func (re *recommendationEngine) ApplyBusinessRules(intent *domain.PurchaseIntent, context *domain.DecisionContext) *domain.DecisionOutcome {
    outcome := &domain.DecisionOutcome{
        IntentID: intent.ID,
        UserID:   intent.UserID,
    }
    
    financial := context.FinancialContext
    health := context.HealthContext
    
    // Rule 1: Emergency fund check
    if financial.EmergencyFundMonths < 3 && intent.Urgency != "critical" {
        outcome.Decision = "BYE"
        outcome.PrimaryReason = "Insufficient emergency fund (less than 3 months)"
        outcome.Confidence = 0.9
        return outcome
    }
    
    // Rule 2: Debt-to-income ratio
    if financial.DebtToIncomeRatio > 0.5 && intent.Category != "health" {
        outcome.Decision = "WAIT"
        outcome.PrimaryReason = "High debt-to-income ratio"
        outcome.WaitPeriod = 90
        outcome.Confidence = 0.85
        return outcome
    }
    
    // Rule 3: Health priority
    if health.HasCriticalConditions && intent.Category == "health" {
        outcome.Decision = "BUY"
        outcome.PrimaryReason = "Critical health need"
        outcome.Confidence = 0.95
        return outcome
    }
    
    // Rule 4: Affordability check
    monthlyImpact := intent.ItemCost
    if intent.Frequency == "monthly" {
        monthlyImpact = intent.ItemCost
    } else if intent.Frequency == "yearly" {
        monthlyImpact = intent.ItemCost / 12
    }
    
    if monthlyImpact > financial.DisposableIncome * 0.3 {
        outcome.Decision = "BYE"
        outcome.PrimaryReason = "Exceeds 30% of disposable income"
        outcome.MaxBudget = financial.DisposableIncome * 0.3
        outcome.Confidence = 0.8
        return outcome
    }
    
    // Rule 5: Default to BUY if all checks pass
    outcome.Decision = "BUY"
    outcome.PrimaryReason = "Purchase fits within budget"
    outcome.Confidence = 0.75
    return outcome
}
```

---

## 🧪 TDD Test Structure

### Service Tests (Unit Tests with Mocks)
```go
// internal/services/decision_service_test.go
func TestDecisionService_MakeDecision_Scenarios(t *testing.T) {
    tests := []struct {
        name             string
        intent           domain.PurchaseIntent
        financialContext domain.FinancialSnapshot
        healthContext    domain.HealthSnapshot
        mockAIResponse   string
        expectedDecision string
        expectedReason   string
    }{
        {
            name: "high_debt_ratio_non_essential",
            intent: domain.PurchaseIntent{
                ItemName: "Gaming Console",
                ItemCost: 500,
                Category: "entertainment",
                Urgency:  "low",
            },
            financialContext: domain.FinancialSnapshot{
                DebtToIncomeRatio: 0.6,
                DisposableIncome:  500,
            },
            expectedDecision: "WAIT",
            expectedReason:   "High debt-to-income ratio",
        },
        {
            name: "health_emergency_override",
            intent: domain.PurchaseIntent{
                ItemName: "Medical Equipment",
                ItemCost: 1000,
                Category: "health",
                Urgency:  "critical",
            },
            healthContext: domain.HealthSnapshot{
                HasCriticalConditions: true,
            },
            expectedDecision: "BUY",
            expectedReason:   "Critical health need",
        },
        {
            name: "within_budget_good_health",
            intent: domain.PurchaseIntent{
                ItemName: "Work Laptop",
                ItemCost: 1200,
                Category: "electronics",
                Urgency:  "medium",
            },
            financialContext: domain.FinancialSnapshot{
                DisposableIncome:    2000,
                EmergencyFundMonths: 6,
                DebtToIncomeRatio:   0.2,
            },
            expectedDecision: "BUY",
            expectedReason:   "Purchase fits within budget",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation with mocked AI client
        })
    }
}

func TestPromptBuilder_BuildDecisionPrompt(t *testing.T) {
    // Test prompt construction with various contexts
    // Verify all data is included correctly
    // Test prompt size limits
}

func TestDecisionInterpreter_ParseAIResponse(t *testing.T) {
    // Test parsing various AI response formats
    // Test handling of malformed responses
    // Test confidence extraction
}
```

---

## 🚀 Claude Code Commands

### Slash Commands (.claude/commands/)

**`/decision-init`**
```markdown
Initialize decision domain following BuyOrBye architecture:
1. Create GORM models: DecisionRecordModel, AIPromptLogModel
2. Create domain entities: PurchaseIntent, DecisionOutcome, DecisionContext
3. Create DTOs with validation
4. Define service interfaces
5. Implement DecisionService with business rules
6. Create ContextAggregator to gather multi-domain data
7. Create PromptBuilder for AI prompts
8. Implement AI clients (Claude, OpenAI, Ollama)
9. Create DecisionInterpreter for response parsing
10. Create handlers with streaming support
11. Run GORM AutoMigrate
```

**`/decision-test-tdd`**
```markdown
Generate comprehensive decision domain tests:
1. Service tests for decision logic (>90% coverage)
2. Test all business rules (emergency fund, DTI, health priority)
3. Test context aggregation from multiple domains
4. Test prompt building with various scenarios
5. Test AI response parsing and validation
6. Test fallback rules when AI fails
7. Repository tests for decision history (>85%)
8. Handler tests with streaming (>80%)
```

**`/decision-ai-setup`**
```markdown
Configure AI providers:
1. Create Claude client with API key from env
2. Create OpenAI client with fallback
3. Create Ollama client for local testing
4. Implement retry logic with exponential backoff
5. Add response caching for similar queries
6. Implement token counting and limits
7. Add prompt injection protection
```

---

## 🔄 Development Workflow

```bash
# Step 1: TDD - Write tests first
claude "/agent tdd-writer Create decision service tests for all business rules and AI integration"

# Step 2: Create domain models
claude "Create decision domain models: PurchaseIntent, DecisionOutcome, DecisionContext, AIPrompt"

# Step 3: Implement context aggregation
claude "Create ContextAggregator to gather data from Finance, Health, and Transport services"

# Step 4: Implement AI integration
claude "Create AI clients for Claude, OpenAI, and Ollama with proper error handling"

# Step 5: Implement decision service
claude "Implement DecisionService with business rules and AI fallback logic"

# Step 6: Create handlers
claude "Create decision handlers with streaming support for real-time AI responses"

# Step 7: Wire into main
claude "Integrate decision domain with all other domains, add routes with auth"
```

---

## 📊 Business Rules & Decision Logic

### Decision Criteria Hierarchy
1. **Critical Health Needs** → Always BUY if health category and critical
2. **Emergency Fund** → BYE if < 3 months (except critical items)
3. **Debt-to-Income** → WAIT if > 50% (except health)
4. **Affordability** → BYE if > 30% of disposable income
5. **Savings Impact** → WAIT if depletes savings below threshold
6. **Category Priority**:
   - Health: Highest priority
   - Transport: High (if needed for work)
   - Food: Medium-High
   - Education: Medium
   - Entertainment: Low

### Wait Period Calculation
```
If DTI > 40%: Wait 90 days
If Emergency Fund < 6 months: Wait 60 days
If Savings Rate < 10%: Wait 30 days
Default: Wait 14 days
```

### Confidence Scoring
```
High Confidence (0.8-1.0): Clear decision based on hard rules
Medium Confidence (0.6-0.8): Mixed signals, AI recommendation
Low Confidence (0.4-0.6): Edge case, suggest user reconsider
```

---

## 🔗 Integration with Other Domains

```go
// The Decision domain orchestrates all others:

// From Finance Domain
type FinanceClient interface {
    GetFinanceSummary(ctx context.Context, userID string) (*FinanceSummary, error)
    GetDisposableIncome(ctx context.Context, userID string) (float64, error)
    GetDebtToIncomeRatio(ctx context.Context, userID string) (float64, error)
}

// From Health Domain  
type HealthClient interface {
    GetHealthSummary(ctx context.Context, userID string) (*HealthSummary, error)
    GetHealthRiskScore(ctx context.Context, userID string) (int, error)
    HasCriticalConditions(ctx context.Context, userID string) (bool, error)
}

// From Transport Domain (if exists)
type TransportClient interface {
    GetTransportSummary(ctx context.Context, userID string) (*TransportSummary, error)
    GetMonthlyTransportCost(ctx context.Context, userID string) (float64, error)
}

// Decision Service uses all clients to build complete context
```

---

## ✅ Implementation Checklist

- [ ] Domain models (PurchaseIntent, DecisionOutcome, DecisionContext)
- [ ] GORM models with AutoMigrate
- [ ] DTOs with comprehensive validation
- [ ] Service interfaces defined
- [ ] DecisionService with business rules
- [ ] ContextAggregator implementation
- [ ] PromptBuilder for AI prompts
- [ ] AI client implementations (Claude, OpenAI, Ollama)
- [ ] DecisionInterpreter for response parsing
- [ ] RecommendationEngine for fallback rules
- [ ] Handlers with streaming support
- [ ] Database migrations via AutoMigrate
- [ ] Unit tests (>90% service coverage)
- [ ] Integration tests (>85% repository coverage)
- [ ] Handler tests (>80% coverage)
- [ ] AI response mocking for tests
- [ ] Rate limiting for AI calls
- [ ] Response caching for similar queries
- [ ] No GORM in services
- [ ] No DTOs in services

---

## 🔍 Common Decision Domain Pitfalls

| Issue | Solution | Verification |
|-------|----------|--------------|
| AI provider timeout | Implement timeout and fallback | Test with delayed responses |
| Prompt injection | Sanitize user input | Test with malicious prompts |
| Token limit exceeded | Truncate context intelligently | Test with large contexts |
| AI hallucination | Validate against business rules | Test response validation |
| Rate limiting | Implement caching and queuing | Test concurrent requests |
| Inconsistent decisions | Log and compare for same input | Test decision consistency |
| Missing context | Graceful degradation | Test with partial data |

---

## 🔒 Security & Privacy Considerations

```go
// Prompt sanitization
- Remove PII from prompts before logging
- Hash prompts for deduplication
- Encrypt stored AI responses
- Rate limit per user (10 decisions/hour)
- Implement prompt injection detection
- Never expose raw financial/health data in logs

// AI Provider Security
- Rotate API keys regularly
- Use environment variables for keys
- Implement request signing
- Monitor for unusual patterns
- Track token usage per user
```

---

## 📚 References

- [BuyOrBye Architecture](./CLAUDE.md) - Project structure and rules
- [OpenAI API Pricing](https://openai.com/api/pricing/) - GPT-4o-mini documentation
- [Prompt Engineering](https://www.promptingguide.ai) - Best practices

---

*This decision domain is the intelligent core of BuyOrBye, orchestrating all user context to provide personalized, responsible purchase recommendations.*