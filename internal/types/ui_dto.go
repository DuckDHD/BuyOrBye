package types


// FromUserProfile converts UserProfileDTO to simplified UserDTO
func (dto *UserDTO) FromUserProfile(profile UserProfileDTO) {
	dto.ID = profile.ID
	dto.Email = profile.Email
	dto.Name = profile.Name
}

/*
Page DashboardPageDTO dto
Dashboard page data transfer object for full page rendering
*/
type DashboardPageDTO struct {
	Layout LayoutDTO `json:"layout"`
}

/*
Page FinanceOverviewPageDTO dto
Finance overview page data transfer object for full page rendering
*/
type FinanceOverviewPageDTO struct {
	Layout  LayoutDTO   `json:"layout"`
	Summary interface{} `json:"summary"`
}

/*
Page HealthProfilePageDTO dto
Health profile page data transfer object for full page rendering
*/
type HealthProfilePageDTO struct {
	Layout   LayoutDTO     `json:"layout"`
	Profiles []interface{} `json:"profiles"`
}

/*
Page DecisionNewPageDTO dto
New decision page data transfer object for full page rendering
*/
type DecisionNewPageDTO struct {
	Layout     LayoutDTO `json:"layout"`
	Categories []string  `json:"categories"`
}

/*
Page DecisionHistoryPageDTO dto
Decision history page data transfer object for full page rendering
*/
type DecisionHistoryPageDTO struct {
	Layout    LayoutDTO           `json:"layout"`
	Decisions []interface{}       `json:"decisions"`
	Filters   DecisionFiltersDTO  `json:"filters"`
}

/*
Page ErrorPageDTO dto
Error page data transfer object for full page rendering
*/
type ErrorPageDTO struct {
	Title   string            `json:"title"`
	Error   *ErrorResponseDTO `json:"error"`
	Message string            `json:"message"`
}

/*
Partial DashboardContentDTO dto
Dashboard content partial data transfer object for HTMX rendering
*/
type DashboardContentDTO struct {
	User           *UserDTO      `json:"user"`
	FinanceSummary interface{}   `json:"finance_summary"`
	RecentActivity []interface{} `json:"recent_activity"`
}

/*
Partial FinanceOverviewDTO dto
Finance overview partial data transfer object for HTMX rendering
*/
type FinanceOverviewDTO struct {
	Summary interface{} `json:"summary"`
}

/*
Partial IncomeListDTO dto
Income list partial data transfer object for HTMX rendering
*/
type IncomeListDTO struct {
	Incomes []interface{} `json:"incomes"`
}

/*
Partial ExpenseFormDTO dto
Expense form partial data transfer object for HTMX rendering
*/
type ExpenseFormDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Frequency   string  `json:"frequency"`
	Description string  `json:"description"`
}

/*
Partial FinanceSummaryDTO dto
Finance summary partial data transfer object for HTMX rendering
*/
type FinanceSummaryDTO struct {
	TotalIncome              float64  `json:"total_income"`
	TotalExpenses            float64  `json:"total_expenses"`
	DisposableIncome         float64  `json:"disposable_income"`
	DisposableIncomeTrend    string   `json:"disposable_income_trend"`
	DebtToIncomeRatio        float64  `json:"debt_to_income_ratio"`
	SavingsRate              float64  `json:"savings_rate"`
	MonthlySavings           float64  `json:"monthly_savings"`
	HealthScore              int      `json:"health_score"`
	IncomeStabilityScore     int      `json:"income_stability_score"`
	DebtManagementScore      int      `json:"debt_management_score"`
	SavingsScore             int      `json:"savings_score"`
	Recommendations          []string `json:"recommendations"`
	Error                    *ErrorResponseDTO `json:"error,omitempty"`
}

/*
Partial HealthRiskGaugeDTO dto
Health risk gauge partial data transfer object for HTMX rendering
*/
type HealthRiskGaugeDTO struct {
	Risk interface{} `json:"risk"`
}

/*
Partial ConditionAddDTO dto
Condition add form partial data transfer object for HTMX rendering
*/
type ConditionAddDTO struct {
	ProfileID string `json:"profile_id"`
}

/*
Partial InsuranceCardDTO dto
Insurance card partial data transfer object for HTMX rendering
*/
type InsuranceCardDTO struct {
	Policy interface{} `json:"policy"`
}

/*
Partial DecisionResultDTO dto
Decision result partial data transfer object for HTMX rendering
*/
type DecisionResultDTO struct {
	Decision     string        `json:"decision"`
	Confidence   int           `json:"confidence"`
	Reasoning    string        `json:"reasoning"`
	Factors      []interface{} `json:"factors"`
	Alternatives []interface{} `json:"alternatives"`
}

/*
Partial DecisionFilterDTO dto
Decision filter partial data transfer object for HTMX rendering
*/
type DecisionFilterDTO struct {
	Categories []string `json:"categories"`
	DateRange  string   `json:"date_range"`
}

/*
Filter DecisionFiltersDTO dto
Decision filters for history page filtering
*/
type DecisionFiltersDTO struct {
	Search      string   `json:"search"`
	Categories  []string `json:"categories"`
	DateRange   string   `json:"date_range"`
	Decision    string   `json:"decision"`
	Category    string   `json:"category"`
	DateFrom    string   `json:"date_from"`
	DateTo      string   `json:"date_to"`
	MinCost     string   `json:"min_cost"`
	MaxCost     string   `json:"max_cost"`
	SortBy      string   `json:"sort_by"`
	SortOrder   string   `json:"sort_order"`
}

/*
Partial ErrorPartialDTO dto
Error partial data transfer object for HTMX error rendering
*/
type ErrorPartialDTO struct {
	Error   *ErrorResponseDTO `json:"error"`
	Message string            `json:"message"`
}