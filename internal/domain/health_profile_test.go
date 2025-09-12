package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthProfile_CalculateBMI(t *testing.T) {
	tests := []struct {
		name           string
		height         float64 // in cm
		weight         float64 // in kg
		expectedBMI    float64
		expectError    bool
	}{
		{
			name:        "normal_weight_adult",
			height:      170,
			weight:      65,
			expectedBMI: 22.49,
		},
		{
			name:        "underweight_adult",
			height:      180,
			weight:      55,
			expectedBMI: 16.98,
		},
		{
			name:        "overweight_adult",
			height:      165,
			weight:      80,
			expectedBMI: 29.38,
		},
		{
			name:        "obese_adult",
			height:      160,
			weight:      90,
			expectedBMI: 35.16,
		},
		{
			name:        "tall_person",
			height:      200,
			weight:      85,
			expectedBMI: 21.25,
		},
		{
			name:        "zero_height_error",
			height:      0,
			weight:      70,
			expectError: true,
		},
		{
			name:        "negative_height_error",
			height:      -170,
			weight:      70,
			expectError: true,
		},
		{
			name:        "zero_weight_error",
			height:      170,
			weight:      0,
			expectError: true,
		},
		{
			name:        "negative_weight_error",
			height:      170,
			weight:      -70,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := HealthProfile{
				Height: tt.height,
				Weight: tt.weight,
			}

			bmi, err := profile.CalculateBMI()

			if tt.expectError {
				assert.Error(t, err)
				assert.Zero(t, bmi)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedBMI, bmi, 0.01, "BMI calculation should be accurate to 2 decimal places")
			}
		})
	}
}

func TestHealthProfile_Validate(t *testing.T) {
	tests := []struct {
		name        string
		profile     HealthProfile
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_profile",
			profile: HealthProfile{
				ID:         "profile-1",
				UserID:     "user-1",
				Age:        30,
				Gender:     "male",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_female_profile",
			profile: HealthProfile{
				ID:         "profile-2",
				UserID:     "user-2",
				Age:        25,
				Gender:     "female",
				Height:     165.0,
				Weight:     60.0,
				FamilySize: 1,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_other_gender",
			profile: HealthProfile{
				ID:         "profile-3",
				UserID:     "user-3",
				Age:        40,
				Gender:     "other",
				Height:     180.0,
				Weight:     80.0,
				FamilySize: 3,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			expectError: false,
		},
		{
			name: "age_zero_invalid",
			profile: HealthProfile{
				ID:         "profile-4",
				UserID:     "user-4",
				Age:        0,
				Gender:     "male",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "age must be between 1 and 150",
		},
		{
			name: "age_negative_invalid",
			profile: HealthProfile{
				ID:         "profile-5",
				UserID:     "user-5",
				Age:        -5,
				Gender:     "male",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "age must be between 1 and 150",
		},
		{
			name: "age_too_high_invalid",
			profile: HealthProfile{
				ID:         "profile-6",
				UserID:     "user-6",
				Age:        151,
				Gender:     "male",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "age must be between 1 and 150",
		},
		{
			name: "age_boundary_150_valid",
			profile: HealthProfile{
				ID:         "profile-7",
				UserID:     "user-7",
				Age:        150,
				Gender:     "female",
				Height:     160.0,
				Weight:     50.0,
				FamilySize: 1,
			},
			expectError: false,
		},
		{
			name: "height_zero_invalid",
			profile: HealthProfile{
				ID:         "profile-8",
				UserID:     "user-8",
				Age:        30,
				Gender:     "male",
				Height:     0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "height must be positive",
		},
		{
			name: "height_negative_invalid",
			profile: HealthProfile{
				ID:         "profile-9",
				UserID:     "user-9",
				Age:        30,
				Gender:     "male",
				Height:     -175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "height must be positive",
		},
		{
			name: "weight_zero_invalid",
			profile: HealthProfile{
				ID:         "profile-10",
				UserID:     "user-10",
				Age:        30,
				Gender:     "male",
				Height:     175.0,
				Weight:     0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "weight must be positive",
		},
		{
			name: "weight_negative_invalid",
			profile: HealthProfile{
				ID:         "profile-11",
				UserID:     "user-11",
				Age:        30,
				Gender:     "male",
				Height:     175.0,
				Weight:     -70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "weight must be positive",
		},
		{
			name: "invalid_gender",
			profile: HealthProfile{
				ID:         "profile-12",
				UserID:     "user-12",
				Age:        30,
				Gender:     "unknown",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "gender must be one of: male, female, other",
		},
		{
			name: "empty_gender_invalid",
			profile: HealthProfile{
				ID:         "profile-13",
				UserID:     "user-13",
				Age:        30,
				Gender:     "",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "gender must be one of: male, female, other",
		},
		{
			name: "family_size_zero_invalid",
			profile: HealthProfile{
				ID:         "profile-14",
				UserID:     "user-14",
				Age:        30,
				Gender:     "male",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 0,
			},
			expectError: true,
			errorMsg:    "family size must be at least 1",
		},
		{
			name: "family_size_negative_invalid",
			profile: HealthProfile{
				ID:         "profile-15",
				UserID:     "user-15",
				Age:        30,
				Gender:     "male",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: -2,
			},
			expectError: true,
			errorMsg:    "family size must be at least 1",
		},
		{
			name: "empty_user_id_invalid",
			profile: HealthProfile{
				ID:         "profile-16",
				UserID:     "",
				Age:        30,
				Gender:     "male",
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHealthProfile_HasHighRisk(t *testing.T) {
	tests := []struct {
		name               string
		age                int
		height             float64
		weight             float64
		hasChronicConditions bool
		expectedHighRisk   bool
		reason             string
	}{
		{
			name:               "young_healthy_low_risk",
			age:                25,
			height:             175,
			weight:             70, // BMI: 22.86
			hasChronicConditions: false,
			expectedHighRisk:   false,
			reason:             "Young age, normal BMI, no chronic conditions",
		},
		{
			name:               "middle_aged_normal_weight_low_risk",
			age:                45,
			height:             170,
			weight:             65, // BMI: 22.49
			hasChronicConditions: false,
			expectedHighRisk:   false,
			reason:             "Middle age but normal BMI, no chronic conditions",
		},
		{
			name:               "elderly_high_risk",
			age:                70,
			height:             170,
			weight:             70, // BMI: 24.22
			hasChronicConditions: false,
			expectedHighRisk:   true,
			reason:             "Age over 65 is high risk",
		},
		{
			name:               "young_underweight_high_risk",
			age:                25,
			height:             180,
			weight:             50, // BMI: 15.43
			hasChronicConditions: false,
			expectedHighRisk:   true,
			reason:             "BMI under 18.5 is high risk",
		},
		{
			name:               "young_obese_high_risk",
			age:                30,
			height:             170,
			weight:             100, // BMI: 34.60
			hasChronicConditions: false,
			expectedHighRisk:   true,
			reason:             "BMI over 30 is high risk",
		},
		{
			name:               "middle_aged_overweight_borderline",
			age:                40,
			height:             175,
			weight:             85, // BMI: 27.76
			hasChronicConditions: false,
			expectedHighRisk:   false,
			reason:             "Overweight but not obese, under 65",
		},
		{
			name:               "young_with_chronic_conditions_high_risk",
			age:                30,
			height:             170,
			weight:             70, // BMI: 24.22
			hasChronicConditions: true,
			expectedHighRisk:   true,
			reason:             "Chronic conditions make it high risk regardless of other factors",
		},
		{
			name:               "elderly_underweight_multiple_risks",
			age:                75,
			height:             165,
			weight:             45, // BMI: 16.53
			hasChronicConditions: true,
			expectedHighRisk:   true,
			reason:             "Multiple risk factors: age, low BMI, chronic conditions",
		},
		{
			name:               "boundary_age_64_normal_bmi",
			age:                64,
			height:             175,
			weight:             75, // BMI: 24.49
			hasChronicConditions: false,
			expectedHighRisk:   false,
			reason:             "Just under 65 age threshold",
		},
		{
			name:               "boundary_age_65_normal_bmi",
			age:                65,
			height:             175,
			weight:             75, // BMI: 24.49
			hasChronicConditions: false,
			expectedHighRisk:   true,
			reason:             "Age 65 threshold reached",
		},
		{
			name:               "boundary_bmi_18_5",
			age:                30,
			height:             170,
			weight:             53.465, // BMI: exactly 18.5
			hasChronicConditions: false,
			expectedHighRisk:   false,
			reason:             "BMI exactly at 18.5 threshold",
		},
		{
			name:               "boundary_bmi_30",
			age:                30,
			height:             170,
			weight:             86.7, // BMI: exactly 30.0
			hasChronicConditions: false,
			expectedHighRisk:   true,
			reason:             "BMI exactly at 30.0 threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := HealthProfile{
				Age:                  tt.age,
				Height:               tt.height,
				Weight:               tt.weight,
				HasChronicConditions: tt.hasChronicConditions,
			}

			// Calculate BMI first
			bmi, err := profile.CalculateBMI()
			require.NoError(t, err)
			profile.BMI = bmi

			hasHighRisk := profile.HasHighRisk()
			assert.Equal(t, tt.expectedHighRisk, hasHighRisk, "Test case: %s - %s", tt.name, tt.reason)
		})
	}
}

func TestHealthProfile_UpdateBMI(t *testing.T) {
	tests := []struct {
		name          string
		initialHeight float64
		initialWeight float64
		newHeight     float64
		newWeight     float64
		expectedBMI   float64
	}{
		{
			name:          "height_weight_change",
			initialHeight: 170,
			initialWeight: 70,
			newHeight:     175,
			newWeight:     75,
			expectedBMI:   24.49,
		},
		{
			name:          "weight_only_change",
			initialHeight: 170,
			initialWeight: 70,
			newHeight:     170,
			newWeight:     80,
			expectedBMI:   27.68,
		},
		{
			name:          "height_only_change",
			initialHeight: 170,
			initialWeight: 70,
			newHeight:     180,
			newWeight:     70,
			expectedBMI:   21.60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := HealthProfile{
				Height: tt.initialHeight,
				Weight: tt.initialWeight,
			}

			// Update height and weight
			profile.Height = tt.newHeight
			profile.Weight = tt.newWeight

			err := profile.UpdateBMI()
			assert.NoError(t, err)
			assert.InDelta(t, tt.expectedBMI, profile.BMI, 0.01)
		})
	}
}

func TestHealthProfile_GetAgeGroup(t *testing.T) {
	tests := []struct {
		name             string
		age              int
		expectedAgeGroup string
	}{
		{
			name:             "young_adult",
			age:              25,
			expectedAgeGroup: "young_adult",
		},
		{
			name:             "boundary_young_adult_29",
			age:              29,
			expectedAgeGroup: "young_adult",
		},
		{
			name:             "middle_aged_30",
			age:              30,
			expectedAgeGroup: "middle_aged",
		},
		{
			name:             "middle_aged_45",
			age:              45,
			expectedAgeGroup: "middle_aged",
		},
		{
			name:             "middle_aged_64",
			age:              64,
			expectedAgeGroup: "middle_aged",
		},
		{
			name:             "senior_65",
			age:              65,
			expectedAgeGroup: "senior",
		},
		{
			name:             "senior_80",
			age:              80,
			expectedAgeGroup: "senior",
		},
		{
			name:             "child_5",
			age:              5,
			expectedAgeGroup: "child",
		},
		{
			name:             "child_17",
			age:              17,
			expectedAgeGroup: "child",
		},
		{
			name:             "young_adult_18",
			age:              18,
			expectedAgeGroup: "young_adult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := HealthProfile{
				Age: tt.age,
			}

			ageGroup := profile.GetAgeGroup()
			assert.Equal(t, tt.expectedAgeGroup, ageGroup)
		})
	}
}

func TestHealthProfile_GetBMICategory(t *testing.T) {
	tests := []struct {
		name            string
		bmi             float64
		expectedCategory string
	}{
		{
			name:            "underweight",
			bmi:             17.5,
			expectedCategory: "underweight",
		},
		{
			name:            "boundary_underweight_18_5",
			bmi:             18.5,
			expectedCategory: "normal",
		},
		{
			name:            "normal_low",
			bmi:             20.0,
			expectedCategory: "normal",
		},
		{
			name:            "normal_high",
			bmi:             24.9,
			expectedCategory: "normal",
		},
		{
			name:            "boundary_overweight_25",
			bmi:             25.0,
			expectedCategory: "overweight",
		},
		{
			name:            "overweight",
			bmi:             27.5,
			expectedCategory: "overweight",
		},
		{
			name:            "boundary_obese_30",
			bmi:             30.0,
			expectedCategory: "obese",
		},
		{
			name:            "obese",
			bmi:             35.0,
			expectedCategory: "obese",
		},
		{
			name:            "severely_obese",
			bmi:             40.0,
			expectedCategory: "obese",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := HealthProfile{
				BMI: tt.bmi,
			}

			category := profile.GetBMICategory()
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}