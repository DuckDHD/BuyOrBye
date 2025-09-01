package models

type FinancialValidation struct {
	Income       float64 `json:"income"`
	FixedCharges float64 `json:"fixed_charges"`
	MonthlyLoans float64 `json:"monthly_loans"`
	Confirmed    bool    `json:"confirmed"`
}

type DebtAssessment struct {
	Loans     []LoanInfo `json:"loans"`
	Confirmed bool       `json:"confirmed"`
}

type LoanInfo struct {
	Type            string  `json:"type"` // mortgage, car, personal, etc.
	MonthlyPayment  float64 `json:"monthly_payment"`
	RemainingAmount float64 `json:"remaining_amount"`
	InterestRate    float64 `json:"interest_rate,omitempty"`
}

type HealthContext struct {
	HouseholdSize      int      `json:"household_size"`
	SpecialNeeds       bool     `json:"special_needs"`
	MedicalConditions  []string `json:"medical_conditions,omitempty"`
	MonthlyHealthCosts float64  `json:"monthly_health_costs"`
	Confirmed          bool     `json:"confirmed"`
}

type TransportationInfo struct {
	MainTransport         string       `json:"main_transport"` // car, public, bike, walk, etc.
	VehicleInfo           *VehicleInfo `json:"vehicle_info,omitempty"`
	MonthlyTransportCosts float64      `json:"monthly_transport_costs"`
	Confirmed             bool         `json:"confirmed"`
}

type VehicleInfo struct {
	Brand                    string  `json:"brand"`
	Model                    string  `json:"model,omitempty"`
	Year                     int     `json:"year,omitempty"`
	Kilometers               int     `json:"kilometers"`
	Engine                   string  `json:"engine"`        // diesel, gasoline, electric, hybrid
	GeneralState             string  `json:"general_state"` // excellent, good, fair, poor
	MaintenanceNeeded        bool    `json:"maintenance_needed"`
	EstimatedMaintenanceCost float64 `json:"estimated_maintenance_cost,omitempty"`
}

type PurchaseReason struct {
	PracticalNeed string `json:"practical_need"`
	EmotionalWant string `json:"emotional_want"`
	Urgency       string `json:"urgency"`      // immediate, soon, can_wait
	Alternatives  string `json:"alternatives"` // what alternatives were considered
	Confirmed     bool   `json:"confirmed"`
}
