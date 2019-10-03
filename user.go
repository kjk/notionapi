package notionapi

// User represents a Notion user
type User struct {
	Email                      string `json:"email"`
	FamilyName                 string `json:"family_name"`
	GivenName                  string `json:"given_name"`
	ID                         string `json:"id"`
	Locale                     string `json:"locale"`
	MobileOnboardingCompleted  bool   `json:"mobile_onboarding_completed"`
	OnboardingCompleted        bool   `json:"onboarding_completed"`
	ClipperOnboardingCompleted bool   `json:"clipper_onboarding_completed"`
	ProfilePhoto               string `json:"profile_photo"`
	TimeZone                   string `json:"time_zone"`
	Version                    int    `json:"version"`

	RawJSON map[string]interface{} `json:"-"`
}
