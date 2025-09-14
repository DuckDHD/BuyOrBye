package dtos

/*
UI UserDTO dto
Simple user data transfer object for UI templates
*/
type UserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

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
	Title string   `json:"title"`
	User  *UserDTO `json:"user"`
}

/*
Page FinanceOverviewPageDTO dto
Finance overview page data transfer object for full page rendering
*/
type FinanceOverviewPageDTO struct {
	Title   string      `json:"title"`
	User    *UserDTO    `json:"user"`
	Summary interface{} `json:"summary"`
}

/*
Page HealthProfilePageDTO dto
Health profile page data transfer object for full page rendering
*/
type HealthProfilePageDTO struct {
	Title    string        `json:"title"`
	User     *UserDTO      `json:"user"`
	Profiles []interface{} `json:"profiles"`
}

/*
Page DecisionNewPageDTO dto
New decision page data transfer object for full page rendering
*/
type DecisionNewPageDTO struct {
	Title string   `json:"title"`
	User  *UserDTO `json:"user"`
}

/*
Page DecisionHistoryPageDTO dto
Decision history page data transfer object for full page rendering
*/
type DecisionHistoryPageDTO struct {
	Title     string        `json:"title"`
	User      *UserDTO      `json:"user"`
	Decisions []interface{} `json:"decisions"`
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
	Expense interface{} `json:"expense,omitempty"`
	IsEdit  bool        `json:"is_edit"`
}

/*
Partial FinanceSummaryDTO dto
Finance summary partial data transfer object for HTMX rendering
*/
type FinanceSummaryDTO struct {
	Summary interface{} `json:"summary"`
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
Partial ErrorPartialDTO dto
Error partial data transfer object for HTMX error rendering
*/
type ErrorPartialDTO struct {
	Error   *ErrorResponseDTO `json:"error"`
	Message string            `json:"message"`
}