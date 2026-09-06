package featurespg

type Feature int

const (
	FeatureUnknown Feature = iota
	FeatureChat
	FeatureAnnouncements
	FeatureAbsences
	FeatureEvents
	FeatureGallery
	FeatureMealCancellations
)

func (f Feature) String() string {
	switch f {
	case FeatureChat:
		return "Chat"
	case FeatureAnnouncements:
		return "Announcements"
	case FeatureAbsences:
		return "Absences"
	case FeatureEvents:
		return "Events"
	case FeatureGallery:
		return "Gallery"
	case FeatureMealCancellations:
		return "MealCancellations"
	default:
		return "Unknown"
	}
}

func (f Feature) Ordinal() int {
	return int(f)
}

// AllFeatures lists every feature constant, used to seed one
// OrganizationFeature row per feature when an organization is created.
func AllFeatures() []Feature {
	return []Feature{
		FeatureChat,
		FeatureAnnouncements,
		FeatureAbsences,
		FeatureEvents,
		FeatureGallery,
		FeatureMealCancellations,
	}
}
