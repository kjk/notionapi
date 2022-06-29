package notionapi

type NotionUser struct {
	ID                         string `json:"id"`
	Version                    int    `json:"version"`
	Email                      string `json:"email"`
	GivenName                  string `json:"given_name"`
	FamilyName                 string `json:"family_name"`
	ProfilePhoto               string `json:"profile_photo"`
	OnboardingCompleted        bool   `json:"onboarding_completed"`
	MobileOnboardingCompleted  bool   `json:"mobile_onboarding_completed"`
	ClipperOnboardingCompleted bool   `json:"clipper_onboarding_completed"`
	Name                       string `json:"name"`

	RawJSON map[string]interface{} `json:"-"`
}

type UserRoot struct {
	Role  string `json:"role"`
	Value struct {
		ID                string   `json:"id"`
		Version           int      `json:"version"`
		SpaceViews        []string `json:"space_views"`
		LeftSpaces        []string `json:"left_spaces"`
		SpaceViewPointers []struct {
			ID      string `json:"id"`
			Table   string `json:"table"`
			SpaceID string `json:"spaceId"`
		} `json:"space_view_pointers"`
	} `json:"value"`

	RawJSON map[string]interface{} `json:"-"`
}

type UserSettings struct {
	ID       string `json:"id"`
	Version  int    `json:"version"`
	Settings struct {
		Type                          string   `json:"type"`
		Locale                        string   `json:"locale"`
		Source                        string   `json:"source"`
		Persona                       string   `json:"persona"`
		TimeZone                      string   `json:"time_zone"`
		UsedMacApp                    bool     `json:"used_mac_app"`
		PreferredLocale               string   `json:"preferred_locale"`
		UsedAndroidApp                bool     `json:"used_android_app"`
		UsedWindowsApp                bool     `json:"used_windows_app"`
		StartDayOfWeek                int      `json:"start_day_of_week"`
		UsedMobileWebApp              bool     `json:"used_mobile_web_app"`
		UsedDesktopWebApp             bool     `json:"used_desktop_web_app"`
		SeenViewsIntroModal           bool     `json:"seen_views_intro_modal"`
		PreferredLocaleOrigin         string   `json:"preferred_locale_origin"`
		SeenCommentSidebarV2          bool     `json:"seen_comment_sidebar_v2"`
		SeenPersonaCollection         bool     `json:"seen_persona_collection"`
		SeenFileAttachmentIntro       bool     `json:"seen_file_attachment_intro"`
		HiddenCollectionDescriptions  []string `json:"hidden_collection_descriptions"`
		CreatedEvernoteGettingStarted bool     `json:"created_evernote_getting_started"`
	} `json:"settings"`

	RawJSON map[string]interface{} `json:"-"`
}
