package model

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// OpenRTB 2.5 Data Models

// OpenRTBRequest is the standard RTB request object
type OpenRTBRequest struct {
	ID     string  `json:"id"`
	Imp    []Imp   `json:"imp"`
	Site   *Site   `json:"site,omitempty"`
	App    *App    `json:"app,omitempty"`
	Device *Device `json:"device,omitempty"`
	User   *User   `json:"user,omitempty"`
	AT     int     `json:"at,omitempty"`   // Auction Type: 1=First Price, 2=Second Price
	TMax   int     `json:"tmax,omitempty"` // Max time in ms
}

// Imp describes the ad position or impression being auctioned
type Imp struct {
	ID          string  `json:"id"`
	Banner      *Banner `json:"banner,omitempty"`
	Video       *Video  `json:"video,omitempty"`
	Native      *Native `json:"native,omitempty"`
	Audio       *Audio  `json:"audio,omitempty"`
	DisplayMng  string  `json:"displaymanager,omitempty"`
	DisplayVer  string  `json:"displaymanagerver,omitempty"`
	Instl       int     `json:"instl,omitempty"` // 1 = Interstitial
	TagID       string  `json:"tagid,omitempty"` // Ad Unit ID
	BidFloor    float64 `json:"bidfloor,omitempty"`
	BidFloorCur string  `json:"bidfloorcur,omitempty"`
	Secure      int     `json:"secure,omitempty"` // 1 = HTTPS
	Pmp         *Pmp    `json:"pmp,omitempty"`    // Private Marketplace
}

// Pmp represents Private Marketplace container
type Pmp struct {
	PrivateAuction int    `json:"private_auction,omitempty"` // 1=Private Only
	Deals          []Deal `json:"deals,omitempty"`
}

// Deal represents a direct deal between buyer and seller
type Deal struct {
	ID          string   `json:"id"`
	BidFloor    float64  `json:"bidfloor,omitempty"`
	BidFloorCur string   `json:"bidfloorcur,omitempty"`
	At          int      `json:"at,omitempty"`       // Auction Type
	WSeat       []string `json:"wseat,omitempty"`    // Whitelisted Seats
	WAdomain    []string `json:"wadomain,omitempty"` // Whitelisted Domains
}

// Banner object
type Banner struct {
	W        int      `json:"w,omitempty"`
	H        int      `json:"h,omitempty"`
	Format   []Format `json:"format,omitempty"` // Array of allowed sizes
	Pos      int      `json:"pos,omitempty"`    // Ad Position
	TopFrame int      `json:"topframe,omitempty"`
}

// Format object
type Format struct {
	W int `json:"w,omitempty"`
	H int `json:"h,omitempty"`
}

// Video object (Simplified VAST 4.0)
type Video struct {
	Mimes       []string `json:"mimes,omitempty"`
	MinDuration int      `json:"minduration,omitempty"`
	MaxDuration int      `json:"maxduration,omitempty"`
	Protocols   []int    `json:"protocols,omitempty"` // 2=VAST 2.0, 3=VAST 3.0, 7=VAST 4.0
	W           int      `json:"w,omitempty"`
	H           int      `json:"h,omitempty"`
	StartDelay  int      `json:"startdelay,omitempty"` // 0=Pre-roll
	Linearity   int      `json:"linearity,omitempty"`  // 1=Linear, 2=Non-Linear
}

// Native object (OpenRTB Native 1.2)
type Native struct {
	Request string `json:"request"` // Payload string (JSON)
	Ver     string `json:"ver,omitempty"`
}

// Audio object (DAAST)
type Audio struct {
	Mimes       []string `json:"mimes,omitempty"`
	MinDuration int      `json:"minduration,omitempty"`
	MaxDuration int      `json:"maxduration,omitempty"`
}

// Site object (Web)
type Site struct {
	ID     string   `json:"id,omitempty"`
	Name   string   `json:"name,omitempty"`
	Domain string   `json:"domain,omitempty"`
	Page   string   `json:"page,omitempty"`
	Ref    string   `json:"ref,omitempty"`
	Cat    []string `json:"cat,omitempty"`
}

// App object (Mobile)
type App struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name,omitempty"`
	Bundle   string   `json:"bundle,omitempty"`
	Domain   string   `json:"domain,omitempty"`
	StoreURL string   `json:"storeurl,omitempty"`
	Cat      []string `json:"cat,omitempty"`
	Ver      string   `json:"ver,omitempty"`
}

// BidRequest represents an incoming RTB bid request (Legacy / Proprietary)
type BidRequest struct {
	ID          string                 `json:"id" binding:"required"`
	Timestamp   time.Time              `json:"timestamp"`
	PublisherID string                 `json:"publisher_id" binding:"required"`
	AdSlot      AdSlot                 `json:"ad_slot" binding:"required"`
	User        InternalUser           `json:"user"`
	Device      InternalDevice         `json:"device" binding:"required"`
	Context     map[string]interface{} `json:"context"`
	Pmp         *Pmp                   `json:"pmp,omitempty"` // Internal PMP mapping
	AuctionType int                    `json:"at,omitempty"`  // 1=First Price, 2=Second Price (default: 1)
}

// AdSlot represents the ad placement
type AdSlot struct {
	ID          string   `json:"id" binding:"required"`
	Dimensions  []int    `json:"dimensions"`  // [width, height]
	Position    string   `json:"position"`    // "above-fold", "below-fold"
	Formats     []string `json:"formats"`     // ["banner", "native", "video"]
	BidFloor    float64  `json:"bidfloor"`    // Minimum acceptable bid price in CPM
	BidFloorCur string   `json:"bidfloorcur"` // Currency for bid floor (default: "USD")
}

// InternalUser represents the user/visitor (Legacy)
type InternalUser struct {
	ID         string   `json:"id"`
	Country    string   `json:"country"`
	Language   string   `json:"language"`
	Categories []string `json:"categories"` // Interest categories
	Age        int      `json:"age"`
	Gender     string   `json:"gender"`
}

// User (OpenRTB) - Shared due to similarity but needs careful mapping if strictly separate
// For now, let's rename the Legacy User to InternalUser and use a shared User struct or keep them separate.
// Given strict OpenRTB, let's redefine User to be compliant, and map legacy to it.

// User object (OpenRTB 2.5)
type User struct {
	ID         string `json:"id,omitempty"`
	BuyerUID   string `json:"buyeruid,omitempty"`
	Yob        int    `json:"yob,omitempty"`
	Gender     string `json:"gender,omitempty"`
	Keywords   string `json:"keywords,omitempty"`
	CustomData string `json:"customdata,omitempty"`
	Geo        *Geo   `json:"geo,omitempty"`
	Data       []Data `json:"data,omitempty"`
}

// Data object for User segments
type Data struct {
	ID      string    `json:"id,omitempty"`
	Name    string    `json:"name,omitempty"`
	Segment []Segment `json:"segment,omitempty"`
}

// Segment object
type Segment struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// InternalDevice represents the user's device (Legacy)
type InternalDevice struct {
	Type      string      `json:"type" binding:"required"` // "mobile", "desktop", "tablet"
	OS        string      `json:"os"`                      // "ios", "android", "windows", "macos"
	Browser   string      `json:"browser"`
	IP        string      `json:"ip"`
	UserAgent string      `json:"user_agent"`
	DeviceID  string      `json:"device_id"`
	Make      string      `json:"make,omitempty"`
	Model     string      `json:"model,omitempty"`
	Geo       InternalGeo `json:"geo"`
}

// Device object (OpenRTB 2.5)
type Device struct {
	UA             string  `json:"ua,omitempty"`
	Geo            *Geo    `json:"geo,omitempty"`
	DNT            int     `json:"dnt,omitempty"`
	Lmt            int     `json:"lmt,omitempty"`
	IP             string  `json:"ip,omitempty"`
	IPv6           string  `json:"ipv6,omitempty"`
	DeviceType     int     `json:"devicetype,omitempty"` // 1=Mobile/Tablet, 2=PC, 3=TV...
	Make           string  `json:"make,omitempty"`
	Model          string  `json:"model,omitempty"`
	OS             string  `json:"os,omitempty"`
	OSV            string  `json:"osv,omitempty"`
	Hmv            string  `json:"hwv,omitempty"`
	W              int     `json:"w,omitempty"`
	H              int     `json:"h,omitempty"`
	PPI            int     `json:"ppi,omitempty"`
	PxRatio        float64 `json:"pxratio,omitempty"`
	JS             int     `json:"js,omitempty"` // 1 if support JS
	Language       string  `json:"language,omitempty"`
	Carrier        string  `json:"carrier,omitempty"`
	ConnectionType int     `json:"connectiontype,omitempty"`
	IFA            string  `json:"ifa,omitempty"`
}

// InternalGeo represents geographic location (Legacy)
type InternalGeo struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"country"`
	City    string  `json:"city"`
	Zip     string  `json:"zip"`
}

// Geo object (OpenRTB 2.5)
type Geo struct {
	Lat       float64 `json:"lat,omitempty"`
	Lon       float64 `json:"lon,omitempty"`
	Type      int     `json:"type,omitempty"` // 1=GPS, 2=IP...
	Country   string  `json:"country,omitempty"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	Zip       string  `json:"zip,omitempty"`
	UTCOffset int     `json:"utcoffset,omitempty"`
}

// BidResponse represents the bid response
type BidResponse struct {
	RequestID     string    `json:"request_id"`
	CampaignID    string    `json:"campaign_id"`
	BidPrice      float64   `json:"bid_price"`
	CreativeURL   string    `json:"creative_url"`
	ImpressionURL string    `json:"impression_url"`
	ClickURL      string    `json:"click_url"`
	TTL           int       `json:"ttl"`                 // Time to live in seconds
	AdMarkup      string    `json:"ad_markup,omitempty"` // HTML or VAST XML
	DealID        string    `json:"deal_id,omitempty"`   // PMP Deal ID
	Timestamp     time.Time `json:"timestamp"`
}

// NoBidResponse indicates no suitable ad found
type NoBidResponse struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason"`
}

// Campaign represents an active campaign in cache
type Campaign struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenantId"`
	Name           string  `json:"name"`
	Type           string  `json:"type"` // "cpm", "cpc", "cpa"
	BidPrice       float64 `json:"bidPrice"`
	Budget         float64 `json:"budget"`
	DailyBudget    float64 `json:"dailyBudget,omitempty"`    // Daily spend cap (defaults to Budget/30 if 0)
	PacingStrategy string  `json:"pacingStrategy,omitempty"` // "asap", "even", "front", "back" (default: even)
	Priority       int     `json:"priority,omitempty"`       // Priority level 1-10 (default: 5, higher = more important)
	Spent          float64 `json:"spent"`
	// Goal-Based Pacing
	GoalType      string `json:"goalType,omitempty"`      // "impressions", "clicks", "conversions"
	GoalTarget    int64  `json:"goalTarget,omitempty"`    // Target number of events to deliver
	GoalDelivered int64  `json:"goalDelivered,omitempty"` // Current progress toward goal
	GoalEndDate   string `json:"goalEndDate,omitempty"`   // Date goal must be achieved (YYYY-MM-DD)
	// Brand Safety
	BrandSafetyLevel  string    `json:"brandSafetyLevel,omitempty"`  // "strict", "standard", "relaxed" (default: standard)
	BlockedCategories []string  `json:"blockedCategories,omitempty"` // IAB categories to avoid (e.g., "IAB25" adult)
	BlockedPublishers []string  `json:"blockedPublishers,omitempty"` // Publisher IDs to block
	BlockedKeywords   []string  `json:"blockedKeywords,omitempty"`   // Keywords to avoid in content
	Targeting         Targeting `json:"targeting"`
	Creative          Creative  `json:"creative"`
	Status            string    `json:"status"`
	// PMP / Deal Targeting
	DealID       string  `json:"dealId,omitempty"`       // For PMP Deals
	DealType     string  `json:"dealType,omitempty"`     // "preferred", "guaranteed", "private_auction", "open_auction"
	DealPriority int     `json:"dealPriority,omitempty"` // Priority within deal (1-10, higher = more important)
	DealPrice    float64 `json:"dealPrice,omitempty"`    // Override bid price for this deal (0 = use BidPrice)
}

type Creative struct {
	Type        string `json:"type"` // "banner", "video", "native", "audio", "rich_media", "playable", "pop", "push"
	URL         string `json:"url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Duration    int    `json:"duration"` // seconds
	MimeType    string `json:"mimeType"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	CTAText     string `json:"ctaText"`
	// Rich Media / Playable / Pop specific
	HTMLSnippet string `json:"htmlSnippet,omitempty"`
	Expandable  bool   `json:"expandable,omitempty"`

	// Audio specific
	Bitrate int `json:"bitrate,omitempty"`

	// Rewarded Video specific
	Rewarded   bool   `json:"rewarded,omitempty"`   // Is this a rewarded video?
	RewardAmt  int    `json:"rewardAmt,omitempty"`  // Amount of currency
	RewardType string `json:"rewardType,omitempty"` // Type (e.g., "coins", "lives")
}

// Targeting represents campaign targeting criteria
type Targeting struct {
	Countries  []string   `json:"countries"`
	Devices    []string   `json:"devices"`
	OS         []string   `json:"os"`
	Categories []string   `json:"categories"`
	GeoFences  []GeoFence `json:"geoFences"`
	MinAge     int        `json:"minAge"`
	MaxAge     int        `json:"maxAge"`
	// Frequency Capping
	FreqCapImpressions int `json:"freqCapImpressions,omitempty"` // Max impressions per user per window
	FreqCapWindowSecs  int `json:"freqCapWindowSecs,omitempty"`  // Window size in seconds (e.g., 86400 = 24h)
	// Dayparting (Hour-of-Day Targeting)
	HourSchedule []int `json:"hourSchedule,omitempty"` // Hours 0-23 when campaign is active (empty = always active)
	// Day-of-Week Targeting
	DayOfWeekTargeting *DayOfWeekTargeting `json:"dayOfWeekTargeting,omitempty"` // Day-specific scheduling with bid modifiers
	// Retargeting
	RetargetingMode       string   `json:"retargetingMode,omitempty"`       // "include" (only target), "exclude" (suppress), "" (no retargeting)
	RetargetingEvents     []string `json:"retargetingEvents,omitempty"`     // Event types: "impression", "click", "conversion", "add_to_cart", "page_view"
	RetargetingCampaigns  []string `json:"retargetingCampaigns,omitempty"`  // Campaign IDs to target users from (empty = this campaign)
	RetargetingWindowDays int      `json:"retargetingWindowDays,omitempty"` // Lookback window in days (default: 30)
	// Contextual Targeting
	ContextualKeywords     []ContextualKeyword `json:"contextualKeywords,omitempty"`     // Keywords to match in page content
	ContextualCategories   []string            `json:"contextualCategories,omitempty"`   // IAB categories to target (e.g., "IAB1-1" = Books & Literature)
	ContextualExcludeWords []string            `json:"contextualExcludeWords,omitempty"` // Keywords that disqualify a page
	// Audience Segment Scoring
	AudienceSegments []AudienceSegment `json:"audienceSegments,omitempty"` // Weighted audience segments to target
	// Weather-Based Targeting
	WeatherTargeting *WeatherTargeting `json:"weatherTargeting,omitempty"` // Weather condition targeting
	// Cross-Device Targeting
	CrossDeviceEnabled bool `json:"crossDeviceEnabled,omitempty"` // Enable cross-device frequency capping and targeting
	// POI (Point-of-Interest) Targeting
	POITargeting *POITargeting `json:"poiTargeting,omitempty"` // Location-based POI targeting
	// Carrier/ISP Targeting
	CarrierTargeting *CarrierTargeting `json:"carrierTargeting,omitempty"` // Mobile carrier and ISP targeting
	// Language Targeting
	LanguageTargeting *LanguageTargeting `json:"languageTargeting,omitempty"` // Language-based targeting with boosts
	// Ad Position Targeting
	AdPositionTargeting *AdPositionTargeting `json:"adPositionTargeting,omitempty"` // Ad position-based targeting with boosts
	// App Category Targeting
	AppTargeting *AppTargeting `json:"appTargeting,omitempty"` // In-app targeting by category and bundle ID
	// Seasonal/Event Targeting
	SeasonalTargeting *SeasonalTargeting `json:"seasonalTargeting,omitempty"` // Seasonal and event-based bid modifiers
	// Demographic Targeting
	DemographicTargeting *DemographicTargeting `json:"demographicTargeting,omitempty"` // Age, gender, income targeting
	VideoTargeting       *VideoTargeting       `json:"videoTargeting,omitempty"`       // Video ad targeting options
	PerformanceGoals     *PerformanceGoals     `json:"performanceGoals,omitempty"`     // Goal-based optimization settings
	InventoryQuality     *InventoryQuality     `json:"inventoryQuality,omitempty"`     // Inventory quality targeting
	DealTargeting        *DealTargeting        `json:"dealTargeting,omitempty"`        // PMP/Deal targeting options
	// Advanced Features
	BidLandscape            *BidLandscape            `json:"bidLandscape,omitempty"`            // Bid landscape analysis
	CreativeOptimization    *CreativeOptimization    `json:"creativeOptimization,omitempty"`    // Auto creative selection
	IncrementalityConfig    *IncrementalityConfig    `json:"incrementalityConfig,omitempty"`    // Lift measurement
	PrivacySandbox          *PrivacySandbox          `json:"privacySandbox,omitempty"`          // Privacy Sandbox APIs
	ContextualAI            *ContextualAI            `json:"contextualAi,omitempty"`            // ML contextual targeting
	AlertConfig             *AlertConfig             `json:"alertConfig,omitempty"`             // Real-time alerts
	CompetitiveIntelligence *CompetitiveIntelligence `json:"competitiveIntelligence,omitempty"` // Competitor analysis
	UnifiedIDConfig         *UnifiedIDConfig         `json:"unifiedIdConfig,omitempty"`         // Cross-platform identity
}

// ContextualKeyword represents a keyword with optional boost multiplier
type ContextualKeyword struct {
	Keyword string  `json:"keyword"`         // The keyword to match (case-insensitive)
	Boost   float64 `json:"boost,omitempty"` // Bid multiplier when matched (default: 1.2)
	Exact   bool    `json:"exact,omitempty"` // Require exact match vs contains
}

// AudienceSegment represents a targetable user segment with scoring weight
type AudienceSegment struct {
	SegmentID string  `json:"segmentId"`          // Segment identifier (e.g., "seg-123", "dmp-automotive")
	Name      string  `json:"name,omitempty"`     // Human-readable name
	Source    string  `json:"source,omitempty"`   // "first_party", "third_party", "lookalike", "contextual"
	Weight    float64 `json:"weight,omitempty"`   // Bid multiplier (default: 1.2, range: 0.5-3.0)
	Required  bool    `json:"required,omitempty"` // If true, user MUST be in this segment
	Exclude   bool    `json:"exclude,omitempty"`  // If true, exclude users in this segment
}

// WeatherTargeting defines weather-based targeting conditions
type WeatherTargeting struct {
	Conditions     []WeatherCondition `json:"conditions,omitempty"`     // Weather conditions to target
	TemperatureMin *float64           `json:"temperatureMin,omitempty"` // Min temp in Celsius (nil = no min)
	TemperatureMax *float64           `json:"temperatureMax,omitempty"` // Max temp in Celsius (nil = no max)
	HumidityMin    *int               `json:"humidityMin,omitempty"`    // Min humidity % (nil = no min)
	HumidityMax    *int               `json:"humidityMax,omitempty"`    // Max humidity % (nil = no max)
	DefaultBoost   float64            `json:"defaultBoost,omitempty"`   // Default boost when conditions match (default: 1.3)
}

// WeatherCondition represents a specific weather condition to target
type WeatherCondition struct {
	Condition string  `json:"condition"`          // "sunny", "cloudy", "rainy", "snowy", "stormy", "windy", "foggy", "hot", "cold"
	Boost     float64 `json:"boost,omitempty"`    // Bid multiplier when matched (default: 1.3)
	Required  bool    `json:"required,omitempty"` // If true, this condition must be present
}

// GeoFence defines a circular region
type GeoFence struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Radius float64 `json:"radius"` // in km
	Name   string  `json:"name"`
}

// POI represents a Point of Interest with targeting configuration
type POI struct {
	ID          string   `json:"id"`                    // Unique POI ID
	Name        string   `json:"name"`                  // POI name (e.g., "Times Square", "LAX Airport")
	Lat         float64  `json:"lat"`                   // Latitude
	Lon         float64  `json:"lon"`                   // Longitude
	Category    string   `json:"category"`              // POI category: retail, restaurant, airport, stadium, hotel, etc.
	Subcategory string   `json:"subcategory,omitempty"` // More specific: fast_food, luxury_retail, etc.
	Radius      float64  `json:"radius"`                // Targeting radius in km
	Boost       float64  `json:"boost,omitempty"`       // Bid multiplier when user is within radius (default: 1.3)
	Required    bool     `json:"required,omitempty"`    // If true, user must be near this POI
	Tags        []string `json:"tags,omitempty"`        // Additional tags for filtering
}

// POITargeting configures point-of-interest based targeting
type POITargeting struct {
	POIs              []POI           `json:"pois,omitempty"`              // Specific POIs to target
	Categories        []string        `json:"categories,omitempty"`        // Target all POIs in these categories
	ExcludeCategories []string        `json:"excludeCategories,omitempty"` // Exclude POIs in these categories
	MinDistance       float64         `json:"minDistance,omitempty"`       // Minimum distance in km (for exclusion zones)
	MaxDistance       float64         `json:"maxDistance,omitempty"`       // Maximum distance in km
	DistanceBoosts    []DistanceBoost `json:"distanceBoosts,omitempty"`    // Distance-based bid multipliers
}

// DistanceBoost defines bid multiplier based on distance from POI
type DistanceBoost struct {
	MaxDistance float64 `json:"maxDistance"` // Up to this distance (km)
	Boost       float64 `json:"boost"`       // Bid multiplier (e.g., 1.5 = 50% boost)
}

// POIResult represents the result of POI targeting evaluation
type POIResult struct {
	Matched     bool     `json:"matched"`
	Blocked     bool     `json:"blocked"`
	Multiplier  float64  `json:"multiplier"`
	NearestPOI  *POI     `json:"nearestPoi,omitempty"`
	Distance    float64  `json:"distance"` // Distance to nearest POI in km
	MatchedPOIs []string `json:"matchedPois,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

// CarrierTargeting configures mobile carrier and ISP targeting
type CarrierTargeting struct {
	Carriers           []CarrierRule `json:"carriers,omitempty"`           // Target specific carriers with boosts
	ISPs               []ISPRule     `json:"isps,omitempty"`               // Target specific ISPs (for WiFi/desktop)
	ConnectionTypes    []string      `json:"connectionTypes,omitempty"`    // "wifi", "cellular", "ethernet", "unknown"
	ExcludeCarriers    []string      `json:"excludeCarriers,omitempty"`    // Carrier names to exclude
	ExcludeISPs        []string      `json:"excludeIsps,omitempty"`        // ISP names to exclude
	CellularOnly       bool          `json:"cellularOnly,omitempty"`       // Only target users on cellular data
	WiFiOnly           bool          `json:"wifiOnly,omitempty"`           // Only target users on WiFi
	MinConnectionSpeed int           `json:"minConnectionSpeed,omitempty"` // Minimum connection speed in Mbps (estimated)
}

// CarrierRule defines a mobile carrier target with optional boost
type CarrierRule struct {
	Name     string  `json:"name"`               // Carrier name: "verizon", "att", "t-mobile", "sprint", etc.
	MCC      string  `json:"mcc,omitempty"`      // Mobile Country Code (optional for precision)
	MNC      string  `json:"mnc,omitempty"`      // Mobile Network Code (optional for precision)
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier (default: 1.2)
	Required bool    `json:"required,omitempty"` // If true, user MUST be on this carrier
}

// ISPRule defines an ISP target with optional boost
type ISPRule struct {
	Name     string  `json:"name"`               // ISP name: "comcast", "verizon_fios", "spectrum", etc.
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier (default: 1.2)
	Required bool    `json:"required,omitempty"` // If true, user MUST be on this ISP
}

// CarrierResult represents the result of carrier/ISP targeting evaluation
type CarrierResult struct {
	Matched        bool    `json:"matched"`
	Blocked        bool    `json:"blocked"`
	Multiplier     float64 `json:"multiplier"`
	Carrier        string  `json:"carrier,omitempty"`
	ISP            string  `json:"isp,omitempty"`
	ConnectionType string  `json:"connectionType,omitempty"`
	Reason         string  `json:"reason,omitempty"`
}

// LanguageTargeting configures language-based targeting with bid modifiers
type LanguageTargeting struct {
	Languages         []LanguageRule `json:"languages,omitempty"`         // Target specific languages with boosts
	ExcludeLanguages  []string       `json:"excludeLanguages,omitempty"`  // Language codes to exclude (e.g., "zh", "ar")
	PrimaryOnly       bool           `json:"primaryOnly,omitempty"`       // Only match primary/device language, not content language
	ContentLanguage   bool           `json:"contentLanguage,omitempty"`   // Also match page/content language
	LocaleMatching    bool           `json:"localeMatching,omitempty"`    // Match full locale (e.g., "en-US" vs just "en")
	DefaultMultiplier float64        `json:"defaultMultiplier,omitempty"` // Multiplier when no specific language match (default: 1.0)
}

// LanguageRule defines a language target with optional boost
type LanguageRule struct {
	Code     string  `json:"code"`               // ISO 639-1 code: "en", "es", "fr", "de", "zh", "ja", etc.
	Locale   string  `json:"locale,omitempty"`   // Optional locale: "en-US", "en-GB", "es-MX", "zh-CN"
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier (default: 1.2)
	Required bool    `json:"required,omitempty"` // If true, user MUST have this language
}

// LanguageResult represents the result of language targeting evaluation
type LanguageResult struct {
	Matched         bool    `json:"matched"`
	Blocked         bool    `json:"blocked"`
	Multiplier      float64 `json:"multiplier"`
	UserLanguage    string  `json:"userLanguage,omitempty"`
	ContentLanguage string  `json:"contentLanguage,omitempty"`
	MatchedCode     string  `json:"matchedCode,omitempty"`
	Reason          string  `json:"reason,omitempty"`
}

// DayOfWeekTargeting configures day-specific scheduling with bid modifiers
type DayOfWeekTargeting struct {
	Days         []DaySchedule `json:"days,omitempty"`         // Per-day configuration
	WeekdaysOnly bool          `json:"weekdaysOnly,omitempty"` // Only run Mon-Fri
	WeekendsOnly bool          `json:"weekendsOnly,omitempty"` // Only run Sat-Sun
	Timezone     string        `json:"timezone,omitempty"`     // Timezone for scheduling (default: UTC)
	DefaultBoost float64       `json:"defaultBoost,omitempty"` // Default multiplier when no specific day config
}

// DaySchedule defines configuration for a specific day of the week
type DaySchedule struct {
	Day      int     `json:"day"`                // Day of week: 0=Sunday, 1=Monday, ..., 6=Saturday
	Active   bool    `json:"active"`             // Whether campaign runs on this day
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier for this day (default: 1.0)
	Hours    []int   `json:"hours,omitempty"`    // Override hours for this day (0-23)
	MaxSpend float64 `json:"maxSpend,omitempty"` // Maximum budget for this day
	Priority int     `json:"priority,omitempty"` // Priority level for this day (higher = more aggressive)
}

// DayOfWeekResult represents the result of day-of-week targeting evaluation
type DayOfWeekResult struct {
	Allowed    bool    `json:"allowed"`
	Multiplier float64 `json:"multiplier"`
	DayName    string  `json:"dayName"`
	DayNumber  int     `json:"dayNumber"`
	IsWeekend  bool    `json:"isWeekend"`
	Reason     string  `json:"reason,omitempty"`
}

// AdPositionTargeting configures ad position-based targeting with bid modifiers
type AdPositionTargeting struct {
	Positions         []PositionRule `json:"positions,omitempty"`         // Target specific positions with boosts
	ExcludePositions  []string       `json:"excludePositions,omitempty"`  // Positions to exclude
	AboveFoldOnly     bool           `json:"aboveFoldOnly,omitempty"`     // Only bid on above-the-fold inventory
	AboveFoldBoost    float64        `json:"aboveFoldBoost,omitempty"`    // Boost for above-fold placements (default: 1.3)
	BelowFoldDiscount float64        `json:"belowFoldDiscount,omitempty"` // Discount for below-fold (default: 0.7)
	InterstitialBoost float64        `json:"interstitialBoost,omitempty"` // Boost for interstitial/fullscreen ads
	StickyBoost       float64        `json:"stickyBoost,omitempty"`       // Boost for sticky/fixed position ads
	MinViewability    float64        `json:"minViewability,omitempty"`    // Minimum predicted viewability (0-1)
}

// PositionRule defines a position target with optional boost
type PositionRule struct {
	Position string  `json:"position"`           // Position: "above_fold", "below_fold", "sidebar", "header", "footer", "interstitial", "sticky", "in_feed", "in_article"
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier (default: 1.2)
	Required bool    `json:"required,omitempty"` // If true, ad MUST be in this position
}

// AdPositionResult represents the result of ad position targeting evaluation
type AdPositionResult struct {
	Matched              bool    `json:"matched"`
	Blocked              bool    `json:"blocked"`
	Multiplier           float64 `json:"multiplier"`
	DetectedPosition     string  `json:"detectedPosition,omitempty"`
	IsAboveFold          bool    `json:"isAboveFold"`
	PredictedViewability float64 `json:"predictedViewability,omitempty"`
	Reason               string  `json:"reason,omitempty"`
}

// AppTargeting configures in-app targeting by category and bundle ID
type AppTargeting struct {
	BundleIDs         []AppRule `json:"bundleIds,omitempty"`         // Target specific app bundle IDs with boosts
	Categories        []AppRule `json:"categories,omitempty"`        // Target app store categories
	ExcludeBundleIDs  []string  `json:"excludeBundleIds,omitempty"`  // Bundle IDs to exclude
	ExcludeCategories []string  `json:"excludeCategories,omitempty"` // Categories to exclude
	InAppOnly         bool      `json:"inAppOnly,omitempty"`         // Only bid on in-app inventory
	MobileWebOnly     bool      `json:"mobileWebOnly,omitempty"`     // Only bid on mobile web inventory
	MinAppRating      float64   `json:"minAppRating,omitempty"`      // Minimum app store rating (1-5)
	MinAppAge         int       `json:"minAppAge,omitempty"`         // Minimum app age in days (avoid new/fake apps)
	PremiumAppsBoost  float64   `json:"premiumAppsBoost,omitempty"`  // Boost for premium/top apps
}

// AppRule defines an app target with optional boost
type AppRule struct {
	Value    string  `json:"value"`              // Bundle ID (e.g., "com.spotify.music") or category (e.g., "Games", "News")
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier (default: 1.2)
	Required bool    `json:"required,omitempty"` // If true, app MUST match this rule
}

// AppTargetingResult represents the result of app targeting evaluation
type AppTargetingResult struct {
	Matched      bool    `json:"matched"`
	Blocked      bool    `json:"blocked"`
	Multiplier   float64 `json:"multiplier"`
	BundleID     string  `json:"bundleId,omitempty"`
	AppName      string  `json:"appName,omitempty"`
	Category     string  `json:"category,omitempty"`
	IsInApp      bool    `json:"isInApp"`
	AppRating    float64 `json:"appRating,omitempty"`
	IsPremiumApp bool    `json:"isPremiumApp"`
	Reason       string  `json:"reason,omitempty"`
}

// SeasonalTargeting configures seasonal and event-based bid modifiers
type SeasonalTargeting struct {
	Events            []SeasonalEvent `json:"events,omitempty"`            // Custom events with date ranges
	EnableHolidays    bool            `json:"enableHolidays,omitempty"`    // Auto-detect major holidays
	HolidayBoost      float64         `json:"holidayBoost,omitempty"`      // Boost for holidays (default: 1.3)
	WeekendBoost      float64         `json:"weekendBoost,omitempty"`      // Boost for weekends
	MonthEndBoost     float64         `json:"monthEndBoost,omitempty"`     // Boost for end of month (payday)
	Q4Boost           float64         `json:"q4Boost,omitempty"`           // Boost for Q4 holiday shopping season
	SummerBoost       float64         `json:"summerBoost,omitempty"`       // Boost for summer months
	BackToSchoolBoost float64         `json:"backToSchoolBoost,omitempty"` // Boost for back-to-school (Aug-Sep)
	Timezone          string          `json:"timezone,omitempty"`          // Timezone for date calculations
	Country           string          `json:"country,omitempty"`           // Country for holiday calendar (US, UK, etc.)
}

// SeasonalEvent defines a custom event or promotion period
type SeasonalEvent struct {
	Name      string  `json:"name"`                // Event name: "Black Friday", "Cyber Monday", "Prime Day"
	StartDate string  `json:"startDate"`           // Start date: "2026-11-27" or "MM-DD" for recurring
	EndDate   string  `json:"endDate"`             // End date: "2026-11-27" or "MM-DD" for recurring
	Boost     float64 `json:"boost,omitempty"`     // Bid multiplier (default: 1.5)
	Recurring bool    `json:"recurring,omitempty"` // If true, repeats every year
	Active    bool    `json:"active"`              // Whether event is enabled
}

// SeasonalResult represents the result of seasonal targeting evaluation
type SeasonalResult struct {
	Matched      bool     `json:"matched"`
	Multiplier   float64  `json:"multiplier"`
	ActiveEvents []string `json:"activeEvents,omitempty"`
	IsHoliday    bool     `json:"isHoliday"`
	HolidayName  string   `json:"holidayName,omitempty"`
	Season       string   `json:"season,omitempty"`
	IsQ4         bool     `json:"isQ4"`
	IsWeekend    bool     `json:"isWeekend"`
	IsMonthEnd   bool     `json:"isMonthEnd"`
}

// DemographicTargeting configures age, gender, and income-based targeting
type DemographicTargeting struct {
	AgeRanges          []AgeRange   `json:"ageRanges,omitempty"`          // Target specific age ranges with boosts
	Genders            []GenderRule `json:"genders,omitempty"`            // Target specific genders with boosts
	IncomeLevels       []IncomeRule `json:"incomeLevels,omitempty"`       // Target income brackets
	ParentalStatus     []string     `json:"parentalStatus,omitempty"`     // "parent", "not_parent", "unknown"
	EducationLevels    []string     `json:"educationLevels,omitempty"`    // "high_school", "college", "graduate", etc.
	HomeOwnership      []string     `json:"homeOwnership,omitempty"`      // "owner", "renter", "unknown"
	ExcludeAgeRanges   []AgeRange   `json:"excludeAgeRanges,omitempty"`   // Age ranges to exclude
	ExcludeGenders     []string     `json:"excludeGenders,omitempty"`     // Genders to exclude
	UnknownAgeBoost    float64      `json:"unknownAgeBoost,omitempty"`    // Multiplier when age unknown (default: 0.8)
	UnknownGenderBoost float64      `json:"unknownGenderBoost,omitempty"` // Multiplier when gender unknown (default: 0.8)
}

// AgeRange defines an age bracket with optional boost
type AgeRange struct {
	MinAge   int     `json:"minAge"`             // Minimum age (inclusive)
	MaxAge   int     `json:"maxAge"`             // Maximum age (inclusive)
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier (default: 1.2)
	Required bool    `json:"required,omitempty"` // If true, user MUST be in this age range
}

// GenderRule defines a gender target with optional boost
type GenderRule struct {
	Gender   string  `json:"gender"`             // "male", "female", "other", "unknown"
	Boost    float64 `json:"boost,omitempty"`    // Bid multiplier (default: 1.2)
	Required bool    `json:"required,omitempty"` // If true, user MUST be this gender
}

// IncomeRule defines an income bracket target
type IncomeRule struct {
	Level     string  `json:"level"`               // "low", "medium", "high", "affluent"
	MinIncome int     `json:"minIncome,omitempty"` // Minimum annual income (optional)
	MaxIncome int     `json:"maxIncome,omitempty"` // Maximum annual income (optional)
	Boost     float64 `json:"boost,omitempty"`     // Bid multiplier (default: 1.2)
	Required  bool    `json:"required,omitempty"`  // If true, user MUST be in this bracket
}

// DemographicResult represents the result of demographic targeting evaluation
type DemographicResult struct {
	Matched       bool    `json:"matched"`
	Blocked       bool    `json:"blocked"`
	Multiplier    float64 `json:"multiplier"`
	UserAge       int     `json:"userAge,omitempty"`
	UserGender    string  `json:"userGender,omitempty"`
	IncomeLevel   string  `json:"incomeLevel,omitempty"`
	AgeRangeMatch string  `json:"ageRangeMatch,omitempty"`
	Reason        string  `json:"reason,omitempty"`
}

// VideoTargeting defines video ad targeting configuration
type VideoTargeting struct {
	MinDuration        int                 `json:"minDuration,omitempty"`        // Minimum video ad duration in seconds
	MaxDuration        int                 `json:"maxDuration,omitempty"`        // Maximum video ad duration in seconds
	PlayerSizes        []VideoPlayerSize   `json:"playerSizes,omitempty"`        // Targeted player sizes (small, medium, large)
	Placements         []string            `json:"placements,omitempty"`         // "instream", "outstream", "interstitial", "in-feed"
	Protocols          []int               `json:"protocols,omitempty"`          // VAST protocols (1=VAST 1.0, 2=VAST 2.0, etc.)
	Linearity          *int                `json:"linearity,omitempty"`          // 1=linear, 2=non-linear
	StartDelays        []int               `json:"startDelays,omitempty"`        // Pre-roll(0), mid-roll(>0), post-roll(-1), generic(-2)
	SkipSettings       *VideoSkipSettings  `json:"skipSettings,omitempty"`       // Skip behavior settings
	CompletionRates    *CompletionRateRule `json:"completionRates,omitempty"`    // Target by historical completion rates
	Mimes              []string            `json:"mimes,omitempty"`              // Supported MIME types
	RequireSound       *bool               `json:"requireSound,omitempty"`       // Require sound-on capable
	BoxingAllowed      *bool               `json:"boxingAllowed,omitempty"`      // Allow letterboxing/pillarboxing
	CompanionRequired  bool                `json:"companionRequired,omitempty"`  // Require companion ad support
	InteractiveAllowed bool                `json:"interactiveAllowed,omitempty"` // Allow interactive video ads
}

// VideoPlayerSize defines player size targeting
type VideoPlayerSize struct {
	Size     string  `json:"size"`            // "small", "medium", "large", "xlarge", "unknown"
	MinWidth int     `json:"minWidth"`        // Minimum player width in pixels
	MaxWidth int     `json:"maxWidth"`        // Maximum player width (0 = no max)
	Boost    float64 `json:"boost,omitempty"` // Bid multiplier for this size (default: 1.0)
	Required bool    `json:"required,omitempty"`
}

// VideoSkipSettings defines skippability requirements
type VideoSkipSettings struct {
	SkippableOnly    bool    `json:"skippableOnly,omitempty"`    // Only target skippable inventory
	NonSkippableOnly bool    `json:"nonSkippableOnly,omitempty"` // Only target non-skippable
	MinSkipOffset    int     `json:"minSkipOffset,omitempty"`    // Min seconds before skip allowed
	MaxSkipOffset    int     `json:"maxSkipOffset,omitempty"`    // Max seconds before skip
	SkippableBoost   float64 `json:"skippableBoost,omitempty"`   // Boost for skippable (default: 1.0)
	NonSkipBoost     float64 `json:"nonSkipBoost,omitempty"`     // Boost for non-skippable (default: 1.3)
}

// CompletionRateRule defines completion rate targeting
type CompletionRateRule struct {
	MinCompletionRate    float64 `json:"minCompletionRate,omitempty"`    // Minimum historical completion rate (0-1)
	TargetQuartile       int     `json:"targetQuartile,omitempty"`       // 1=25%, 2=50%, 3=75%, 4=100%
	HighCompletionBoost  float64 `json:"highCompletionBoost,omitempty"`  // Boost for >75% completion rate
	LowCompletionPenalty float64 `json:"lowCompletionPenalty,omitempty"` // Penalty for <25% completion
}

// VideoTargetingResult represents video targeting evaluation
type VideoTargetingResult struct {
	Matched        bool    `json:"matched"`
	Blocked        bool    `json:"blocked"`
	Multiplier     float64 `json:"multiplier"`
	PlayerSize     string  `json:"playerSize,omitempty"`
	Placement      string  `json:"placement,omitempty"`
	Duration       int     `json:"duration,omitempty"`
	Skippable      bool    `json:"skippable"`
	CompletionRate float64 `json:"completionRate,omitempty"`
	Reason         string  `json:"reason,omitempty"`
}

// PerformanceGoals defines campaign optimization goals
type PerformanceGoals struct {
	PrimaryGoal      string                 `json:"primaryGoal"`                // "cpm", "cpc", "cpa", "cpi", "cps", "cpr", "ctv", "viewability", "completion", "engagement", "cpl", "cpv", "cpcv", "cpe", "vcpm", "dcpm", "cpa_d", "cpiaap"
	TargetCPA        float64                `json:"targetCpa,omitempty"`        // Target cost per acquisition
	TargetCPC        float64                `json:"targetCpc,omitempty"`        // Target cost per click
	TargetCPM        float64                `json:"targetCpm,omitempty"`        // Target cost per mille
	TargetCPI        float64                `json:"targetCpi,omitempty"`        // Target cost per install (app campaigns)
	TargetCPS        float64                `json:"targetCps,omitempty"`        // Target cost per sale (e-commerce)
	TargetCPR        float64                `json:"targetCpr,omitempty"`        // Target cost per registration/result
	TargetCPL        float64                `json:"targetCpl,omitempty"`        // Target cost per lead
	TargetCPV        float64                `json:"targetCpv,omitempty"`        // Target cost per view
	TargetCPCV       float64                `json:"targetCpcv,omitempty"`       // Target cost per completed view (video)
	TargetCPE        float64                `json:"targetCpe,omitempty"`        // Target cost per engagement
	TargetVCPM       float64                `json:"targetVcpm,omitempty"`       // Target viewable CPM
	TargetDCPM       float64                `json:"targetDcpm,omitempty"`       // Target dynamic CPM (ML-optimized)
	TargetCPAD       float64                `json:"targetCpad,omitempty"`       // Target cost per app download
	TargetCPIAAP     float64                `json:"targetCpiaap,omitempty"`     // Target cost per in-app purchase
	TargetROAS       float64                `json:"targetRoas,omitempty"`       // Target return on ad spend (e-commerce)
	ViewabilityGoal  float64                `json:"viewabilityGoal,omitempty"`  // Target viewability rate (0-1)
	CompletionGoal   float64                `json:"completionGoal,omitempty"`   // Target video completion rate (0-1)
	EngagementGoal   float64                `json:"engagementGoal,omitempty"`   // Target engagement rate
	CTVGoals         *CTVOptimization       `json:"ctvGoals,omitempty"`         // CTV-specific optimization settings
	AppGoals         *AppOptimization       `json:"appGoals,omitempty"`         // App campaign optimization settings
	EcommerceGoals   *EcommerceOptimization `json:"ecommerceGoals,omitempty"`   // E-commerce optimization settings
	OptimizeFor      []string               `json:"optimizeFor,omitempty"`      // Additional optimization signals
	BidStrategy      string                 `json:"bidStrategy,omitempty"`      // "maximize_conversions", "target_cpa", "target_roas", "maximize_clicks", "manual"
	MaxBidAdjust     float64                `json:"maxBidAdjust,omitempty"`     // Maximum bid adjustment multiplier
	MinBidAdjust     float64                `json:"minBidAdjust,omitempty"`     // Minimum bid adjustment multiplier
	LearningMode     bool                   `json:"learningMode,omitempty"`     // In learning phase (less aggressive optimization)
	ConversionWindow int                    `json:"conversionWindow,omitempty"` // Days to attribute conversions (default: 30)
	ConversionTypes  []string               `json:"conversionTypes,omitempty"`  // Types of conversions to track
	Thresholds       *PerformanceThresholds `json:"thresholds,omitempty"`       // Performance thresholds
	// Attribution Configuration
	AttributionModel  string  `json:"attributionModel,omitempty"`  // "last_click", "first_click", "linear", "time_decay", "position_based"
	TimeDecayHalfLife float64 `json:"timeDecayHalfLife,omitempty"` // Half-life in hours for time_decay model (default: 168 = 7 days)
	// Dayparting Optimization
	DaypartingOptimization *DaypartingOptimization `json:"daypartingOptimization,omitempty"` // Automatic hourly bid adjustments
	// Audience Modeling
	AudienceModeling *AudienceModeling `json:"audienceModeling,omitempty"` // Lookalike expansion and suppression
}

// CTVOptimization defines CTV-specific optimization settings
type CTVOptimization struct {
	TargetCompletionRate  float64  `json:"targetCompletionRate,omitempty"`  // Target VCR for CTV (usually higher)
	TargetReach           int64    `json:"targetReach,omitempty"`           // Target unique households
	TargetFrequency       float64  `json:"targetFrequency,omitempty"`       // Target avg frequency per household
	PreferredDevices      []string `json:"preferredDevices,omitempty"`      // "smart_tv", "roku", "fire_tv", "apple_tv", "gaming_console"
	PreferredApps         []string `json:"preferredApps,omitempty"`         // Preferred streaming apps
	CoViewingBoost        float64  `json:"coViewingBoost,omitempty"`        // Boost for co-viewing inventory
	PrimtimeBoost         float64  `json:"primetimeBoost,omitempty"`        // Boost for primetime hours
	LiveContentBoost      float64  `json:"liveContentBoost,omitempty"`      // Boost for live content
	RequireACR            bool     `json:"requireAcr,omitempty"`            // Require Automatic Content Recognition
	DeduplicationEnabled  bool     `json:"deduplicationEnabled,omitempty"`  // Cross-device deduplication
	HouseholdFrequencyCap int      `json:"householdFrequencyCap,omitempty"` // Max impressions per household
}

// AppOptimization defines app campaign optimization settings
type AppOptimization struct {
	TargetInstallRate    float64    `json:"targetInstallRate,omitempty"`    // Target install rate
	TargetCostPerInstall float64    `json:"targetCostPerInstall,omitempty"` // Target CPI
	TargetRetentionD1    float64    `json:"targetRetentionD1,omitempty"`    // Target Day-1 retention
	TargetRetentionD7    float64    `json:"targetRetentionD7,omitempty"`    // Target Day-7 retention
	TargetLTV            float64    `json:"targetLtv,omitempty"`            // Target lifetime value
	OptimizeForEvents    []string   `json:"optimizeForEvents,omitempty"`    // In-app events to optimize for
	ValuedEvents         []AppEvent `json:"valuedEvents,omitempty"`         // Events with assigned values
	PreferredPlacements  []string   `json:"preferredPlacements,omitempty"`  // "rewarded", "interstitial", "banner"
	ExcludeLowLTVSources bool       `json:"excludeLowLtvSources,omitempty"` // Exclude low-LTV traffic sources
	SKAdNetworkOptimized bool       `json:"skadNetworkOptimized,omitempty"` // iOS SKAdNetwork optimization
	GooglePlayOptimized  bool       `json:"googlePlayOptimized,omitempty"`  // Google Play attribution
}

// AppEvent defines a valued in-app event
type AppEvent struct {
	EventName  string  `json:"eventName"`            // "purchase", "signup", "level_complete", "add_to_cart"
	EventValue float64 `json:"eventValue,omitempty"` // Attributed value
	Priority   int     `json:"priority,omitempty"`   // Optimization priority (1-10)
}

// EcommerceOptimization defines e-commerce campaign optimization
type EcommerceOptimization struct {
	TargetROAS            float64            `json:"targetRoas,omitempty"`            // Target return on ad spend
	TargetCostPerSale     float64            `json:"targetCostPerSale,omitempty"`     // Target CPS
	TargetAOV             float64            `json:"targetAov,omitempty"`             // Target average order value
	MinOrderValue         float64            `json:"minOrderValue,omitempty"`         // Minimum order value to count
	TrackMicroConversions bool               `json:"trackMicroConversions,omitempty"` // Track add-to-cart, wishlist
	ProductCategories     []string           `json:"productCategories,omitempty"`     // Prioritized product categories
	SeasonalAdjustments   map[string]float64 `json:"seasonalAdjustments,omitempty"`   // Seasonal bid adjustments
	CartAbandonBoost      float64            `json:"cartAbandonBoost,omitempty"`      // Boost for cart abandoners
	RepeatCustomerBoost   float64            `json:"repeatCustomerBoost,omitempty"`   // Boost for repeat customers
	NewCustomerPriority   bool               `json:"newCustomerPriority,omitempty"`   // Prioritize new customer acquisition
	DynamicProductAds     bool               `json:"dynamicProductAds,omitempty"`     // Enable DPA optimization
}

// PerformanceThresholds defines minimum performance requirements
type PerformanceThresholds struct {
	MinCTR         float64 `json:"minCtr,omitempty"`         // Minimum click-through rate
	MinViewability float64 `json:"minViewability,omitempty"` // Minimum viewability
	MinCompletion  float64 `json:"minCompletion,omitempty"`  // Minimum video completion
	MaxCPA         float64 `json:"maxCpa,omitempty"`         // Maximum acceptable CPA
	MaxCPI         float64 `json:"maxCpi,omitempty"`         // Maximum acceptable CPI
	MaxCPS         float64 `json:"maxCps,omitempty"`         // Maximum acceptable CPS
	MinROAS        float64 `json:"minRoas,omitempty"`        // Minimum acceptable ROAS
	MinEngagement  float64 `json:"minEngagement,omitempty"`  // Minimum engagement rate
	MinConvRate    float64 `json:"minConvRate,omitempty"`    // Minimum conversion rate
	MinInstallRate float64 `json:"minInstallRate,omitempty"` // Minimum install rate (apps)
}

// PerformanceGoalResult represents performance optimization evaluation
type PerformanceGoalResult struct {
	Matched              bool    `json:"matched"`
	Blocked              bool    `json:"blocked"`
	Multiplier           float64 `json:"multiplier"`
	RecommendedBid       float64 `json:"recommendedBid,omitempty"`
	PredictedCTR         float64 `json:"predictedCtr,omitempty"`
	PredictedCVR         float64 `json:"predictedCvr,omitempty"`
	PredictedViewRate    float64 `json:"predictedViewRate,omitempty"`
	PredictedInstallRate float64 `json:"predictedInstallRate,omitempty"` // For app campaigns
	PredictedROAS        float64 `json:"predictedRoas,omitempty"`        // For e-commerce
	PredictedLTV         float64 `json:"predictedLtv,omitempty"`         // Predicted lifetime value
	IsCTV                bool    `json:"isCtv,omitempty"`                // CTV inventory flag
	HouseholdID          string  `json:"householdId,omitempty"`          // CTV household identifier
	OptimizationLevel    string  `json:"optimizationLevel,omitempty"`    // "aggressive", "moderate", "conservative"
	GoalType             string  `json:"goalType,omitempty"`             // Primary goal being optimized
	Reason               string  `json:"reason,omitempty"`
}

// InventoryQuality defines inventory quality targeting requirements
type InventoryQuality struct {
	MinQualityScore    float64             `json:"minQualityScore,omitempty"`    // Minimum inventory quality (0-1)
	MaxQualityScore    float64             `json:"maxQualityScore,omitempty"`    // Maximum (for cost control)
	TrustLevels        []string            `json:"trustLevels,omitempty"`        // "direct", "reseller", "authorized", "unknown"
	ExcludeTrustLevels []string            `json:"excludeTrustLevels,omitempty"` // Levels to exclude
	RequireAdsTxt      bool                `json:"requireAdsTxt,omitempty"`      // Require ads.txt verification
	RequireSellerJson  bool                `json:"requireSellerJson,omitempty"`  // Require sellers.json verification
	BrandSuitability   *BrandSuitability   `json:"brandSuitability,omitempty"`   // Brand suitability controls
	FraudProtection    *FraudProtection    `json:"fraudProtection,omitempty"`    // Fraud protection settings
	ViewabilityHistory *ViewabilityHistory `json:"viewabilityHistory,omitempty"` // Historical viewability requirements
	QualityTiers       []QualityTier       `json:"qualityTiers,omitempty"`       // Tiered quality boosts
}

// BrandSuitability defines brand safety and suitability controls
type BrandSuitability struct {
	FloorRating         string   `json:"floorRating,omitempty"`         // "G", "PG", "PG13", "R"
	BlockedCategories   []string `json:"blockedCategories,omitempty"`   // IAB content categories to block
	AllowedCategories   []string `json:"allowedCategories,omitempty"`   // Only allow these categories
	RequireVerification bool     `json:"requireVerification,omitempty"` // Require brand safety verification
	SentimentFilters    []string `json:"sentimentFilters,omitempty"`    // "negative", "controversial", "political"
	CustomKeywordBlock  []string `json:"customKeywordBlock,omitempty"`  // Custom blocked keywords
}

// FraudProtection defines fraud protection requirements
type FraudProtection struct {
	MinTrustScore          float64  `json:"minTrustScore,omitempty"`          // Minimum IVT trust score
	BlockBotTraffic        bool     `json:"blockBotTraffic,omitempty"`        // Block suspected bot traffic
	BlockProxyTraffic      bool     `json:"blockProxyTraffic,omitempty"`      // Block proxy/VPN traffic
	RequireAdsVerification bool     `json:"requireAdsVerification,omitempty"` // Require third-party verification
	BlockedSources         []string `json:"blockedSources,omitempty"`         // Known fraudulent sources
	RequireGoogleTag       bool     `json:"requireGoogleTag,omitempty"`       // Require Google Analytics tag
}

// ViewabilityHistory defines historical viewability requirements
type ViewabilityHistory struct {
	MinHistoricalRate float64 `json:"minHistoricalRate,omitempty"` // Min historical viewability (0-1)
	LookbackDays      int     `json:"lookbackDays,omitempty"`      // Days to consider
	MinSampleSize     int     `json:"minSampleSize,omitempty"`     // Minimum impressions for validity
	HighViewBoost     float64 `json:"highViewBoost,omitempty"`     // Boost for >70% viewability
	LowViewPenalty    float64 `json:"lowViewPenalty,omitempty"`    // Penalty for <40% viewability
}

// QualityTier defines quality-based bid adjustments
type QualityTier struct {
	Tier           string  `json:"tier"`                     // "premium", "standard", "remnant"
	MinScore       float64 `json:"minScore,omitempty"`       // Minimum quality score for tier
	MaxScore       float64 `json:"maxScore,omitempty"`       // Maximum quality score
	BidMultiplier  float64 `json:"bidMultiplier,omitempty"`  // Bid adjustment for tier
	MaxBidIncrease float64 `json:"maxBidIncrease,omitempty"` // Cap on bid increase
}

// InventoryQualityResult represents inventory quality evaluation
type InventoryQualityResult struct {
	Matched         bool    `json:"matched"`
	Blocked         bool    `json:"blocked"`
	Multiplier      float64 `json:"multiplier"`
	QualityScore    float64 `json:"qualityScore,omitempty"`
	QualityTier     string  `json:"qualityTier,omitempty"`
	TrustLevel      string  `json:"trustLevel,omitempty"`
	AdsTxtVerified  bool    `json:"adsTxtVerified,omitempty"`
	BrandSafe       bool    `json:"brandSafe,omitempty"`
	FraudRisk       float64 `json:"fraudRisk,omitempty"`
	ViewabilityRate float64 `json:"viewabilityRate,omitempty"`
	Reason          string  `json:"reason,omitempty"`
}

// DealTargeting defines PMP/Deal targeting configuration
type DealTargeting struct {
	PreferredDealIDs   []string        `json:"preferredDealIds,omitempty"`   // Priority deal IDs
	DealTypes          []string        `json:"dealTypes,omitempty"`          // "programmatic_guaranteed", "preferred", "private_auction", "open"
	MinDealPriority    int             `json:"minDealPriority,omitempty"`    // Minimum deal priority (1-10)
	RequireDeal        bool            `json:"requireDeal,omitempty"`        // Only bid on deals
	DealBidAdjustments []DealBidAdjust `json:"dealBidAdjustments,omitempty"` // Deal-specific bid adjustments
	PublisherDeals     []PublisherDeal `json:"publisherDeals,omitempty"`     // Publisher-specific deal preferences
	ExcludedDealIDs    []string        `json:"excludedDealIds,omitempty"`    // Deals to exclude
	PreferPG           bool            `json:"preferPG,omitempty"`           // Prefer Programmatic Guaranteed
	FallbackToOpen     bool            `json:"fallbackToOpen,omitempty"`     // Fall back to open auction if no deal
}

// DealBidAdjust defines bid adjustments for specific deals
type DealBidAdjust struct {
	DealID        string  `json:"dealId"`                  // Deal ID
	BidMultiplier float64 `json:"bidMultiplier,omitempty"` // Bid adjustment (1.0 = no change)
	MaxBid        float64 `json:"maxBid,omitempty"`        // Maximum bid for this deal
	MinBid        float64 `json:"minBid,omitempty"`        // Minimum bid for this deal
	Priority      int     `json:"priority,omitempty"`      // Deal priority override
}

// PublisherDeal defines publisher-specific deal preferences
type PublisherDeal struct {
	PublisherID  string   `json:"publisherId"`            // Publisher ID
	DealIDs      []string `json:"dealIds,omitempty"`      // Associated deal IDs
	BidBoost     float64  `json:"bidBoost,omitempty"`     // Bid boost for this publisher
	PreferredFmt []string `json:"preferredFmt,omitempty"` // Preferred ad formats
	Exclusive    bool     `json:"exclusive,omitempty"`    // Only bid through these deals
}

// DealTargetingResult represents deal targeting evaluation
type DealTargetingResult struct {
	Matched        bool    `json:"matched"`
	Blocked        bool    `json:"blocked"`
	Multiplier     float64 `json:"multiplier"`
	MatchedDealID  string  `json:"matchedDealId,omitempty"`
	DealType       string  `json:"dealType,omitempty"`
	DealPriority   int     `json:"dealPriority,omitempty"`
	EffectiveFloor float64 `json:"effectiveFloor,omitempty"`
	IsPG           bool    `json:"isPG,omitempty"`
	Reason         string  `json:"reason,omitempty"`
}

// BidResult represents the result of a bid evaluation
type BidResult struct {
	Campaign   *Campaign
	Score      float64
	BidPrice   float64
	MatchScore float64
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// IsMatch checks if campaign targeting matches bid request
func (c *Campaign) IsMatch(req *BidRequest) bool {
	// --- PMP Logic ---
	// 1. If Campaign has a DealID, it MUST match a valid Deal in the request (and meet floor price).
	if c.DealID != "" {
		if req.Pmp == nil || len(req.Pmp.Deals) == 0 {
			// Campaign requires a deal, but none requested.
			return false
		}

		matchedDeal := false
		for _, deal := range req.Pmp.Deals {
			if deal.ID == c.DealID {
				// Check Floor Price
				if c.BidPrice >= deal.BidFloor {
					matchedDeal = true
				}
				break
			}
		}
		if !matchedDeal {
			return false // Campaign's deal ID not found in request or price too low
		}
	}

	// 2. If Request is Private Auction (Strict), ONLY campaigns with DealID are allowed.
	if req.Pmp != nil && req.Pmp.PrivateAuction == 1 {
		if c.DealID == "" {
			return false // Open market campaigns not allowed in private auction
		}
		// If c.DealID != "", it was already validated in step 1.
	}

	// Check if campaign creative type is supported by the ad slot
	supported := false
	for _, format := range req.AdSlot.Formats {
		if format == c.Creative.Type {
			supported = true
			break
		}
		// Allow "rich_media" or "video" if "interstitial" is requested and they are compatible
		if format == "interstitial" && (c.Creative.Type == "rich_media" || c.Creative.Type == "video") {
			supported = true
			break
		}
	}
	if !supported {
		return false
	}

	// Check country
	if len(c.Targeting.Countries) > 0 {
		if !contains(c.Targeting.Countries, req.User.Country) {
			return false
		}
	}

	// Check GeoFences
	if len(c.Targeting.GeoFences) > 0 {
		// If request has no valid geo, we cannot match
		if req.Device.Geo.Lat == 0 && req.Device.Geo.Lon == 0 {
			return false
		}

		matchGeo := false
		for _, fence := range c.Targeting.GeoFences {
			dist := haversine(
				req.Device.Geo.Lat, req.Device.Geo.Lon,
				fence.Lat, fence.Lon,
			)
			if dist <= fence.Radius {
				matchGeo = true
				break
			}
		}
		if !matchGeo {
			return false
		}
	}

	// Check device type
	if len(c.Targeting.Devices) > 0 {
		if !contains(c.Targeting.Devices, req.Device.Type) {
			return false
		}
	}

	// Check OS
	if len(c.Targeting.OS) > 0 {
		if !contains(c.Targeting.OS, req.Device.OS) {
			return false
		}
	}

	// Check age
	if c.Targeting.MinAge > 0 && req.User.Age < c.Targeting.MinAge {
		return false
	}
	if c.Targeting.MaxAge > 0 && req.User.Age > c.Targeting.MaxAge {
		return false
	}

	// Check categories (interest overlap)
	if len(c.Targeting.Categories) > 0 && len(req.User.Categories) > 0 {
		if !hasOverlap(c.Targeting.Categories, req.User.Categories) {
			return false
		}
	}

	return true
}

// haversine calculates distance between two points in km
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func hasOverlap(slice1, slice2 []string) bool {
	for _, item := range slice1 {
		if contains(slice2, item) {
			return true
		}
	}
	return false
}

type OpenRTBNative struct {
	Request string `json:"request"`
}

// OpenRTBNativeRequest is the parsed "request" string (Native 1.2 Layout)
// Spec: https://www.iab.com/wp-content/uploads/2018/03/OpenRTB-Native-Ads-Specification-Final-1.2.pdf
type OpenRTBNativeRequest struct {
	Ver         string        `json:"ver,omitempty"`
	Context     int           `json:"context,omitempty"`
	ContextSub  int           `json:"contextsubtype,omitempty"`
	PlcmtType   int           `json:"plcmttype,omitempty"`
	PlcmtCnt    int           `json:"plcmtcnt,omitempty"`
	Seq         int           `json:"seq,omitempty"`
	Assets      []NativeAsset `json:"assets"`
	AUrlSupport int           `json:"aurlsupport,omitempty"`
	DUrlSupport int           `json:"durlsupport,omitempty"`
}

// NativeAsset represents "assets" array in Native Request
// Each asset has an ID and one of {title, img, video, data}
type NativeAsset struct {
	ID       int             `json:"id"`
	Required int             `json:"required,omitempty"`
	Title    *NativeTitleReq `json:"title,omitempty"`
	Img      *NativeImgReq   `json:"img,omitempty"`
	Video    *NativeVideoReq `json:"video,omitempty"`
	Data     *NativeDataReq  `json:"data,omitempty"`
}

type NativeTitleReq struct {
	Len int `json:"len"`
}

type NativeImgReq struct {
	Type  int      `json:"type,omitempty"` // 1=Icon, 3=Main
	W     int      `json:"w,omitempty"`
	WMin  int      `json:"wmin,omitempty"`
	H     int      `json:"h,omitempty"`
	HMin  int      `json:"hmin,omitempty"`
	Mimes []string `json:"mimes,omitempty"`
}

type NativeVideoReq struct {
	Mimes       []string `json:"mimes,omitempty"`
	MinDuration int      `json:"minduration,omitempty"`
	MaxDuration int      `json:"maxduration,omitempty"`
	Protocols   []int    `json:"protocols,omitempty"`
}

// Data Type: 1=sponsored, 2=desc, 3=rating, 12=cta
type NativeDataReq struct {
	Type int `json:"type"`
	Len  int `json:"len,omitempty"`
}

// Touchpoint represents a single interaction in a user's conversion journey
type Touchpoint struct {
	Type       string    `json:"type"`       // "impression", "click", "view"
	RequestID  string    `json:"requestId"`  // Original bid request ID
	CampaignID string    `json:"campaignId"` // Campaign that served the ad
	Timestamp  time.Time `json:"timestamp"`  // When the interaction occurred
	Position   int       `json:"position"`   // Position in the journey (1 = first, -1 = last)
}

// AttributionCredit represents credit assigned to a touchpoint in multi-touch attribution
type AttributionCredit struct {
	Touchpoint Touchpoint `json:"touchpoint"`
	Credit     float64    `json:"credit"` // 0.0 to 1.0, sum of all credits = 1.0
	Model      string     `json:"model"`  // "linear", "time_decay", "position_based", "last_touch", "first_touch"
}

// DaypartingOptimization configures automatic bid adjustments by hour of day
type DaypartingOptimization struct {
	Enabled           bool                       `json:"enabled"`                     // Enable automatic daypart optimization
	HourlyMultipliers map[int]float64            `json:"hourlyMultipliers,omitempty"` // Manual hour-of-day multipliers (0-23 -> multiplier)
	AutoOptimize      bool                       `json:"autoOptimize,omitempty"`      // Use ML-like hourly performance data to auto-adjust
	LookbackDays      int                        `json:"lookbackDays,omitempty"`      // Days of historical data for auto-optimization (default: 14)
	MinMultiplier     float64                    `json:"minMultiplier,omitempty"`     // Minimum hourly multiplier (default: 0.3)
	MaxMultiplier     float64                    `json:"maxMultiplier,omitempty"`     // Maximum hourly multiplier (default: 2.0)
	Timezone          string                     `json:"timezone,omitempty"`          // Timezone for hour calculation (default: UTC)
	DaySpecific       map[string]map[int]float64 `json:"daySpecific,omitempty"`       // Day-of-week specific multipliers: {"monday": {9: 1.5, ...}}
}

// DaypartingResult represents the result of dayparting optimization
type DaypartingResult struct {
	Hour          int     `json:"hour"`
	DayOfWeek     string  `json:"dayOfWeek"`
	Multiplier    float64 `json:"multiplier"`
	HistoricalCTR float64 `json:"historicalCtr,omitempty"`
	HistoricalCVR float64 `json:"historicalCvr,omitempty"`
	IsOptimalHour bool    `json:"isOptimalHour"`
	Reason        string  `json:"reason,omitempty"`
}

// AudienceModeling configures lookalike expansion and audience suppression
type AudienceModeling struct {
	// Lookalike Expansion
	LookalikeEnabled    bool     `json:"lookalikeEnabled,omitempty"`    // Enable lookalike audience expansion
	SeedSegments        []string `json:"seedSegments,omitempty"`        // Seed audience segment IDs for lookalike modeling
	LookalikeExpansion  float64  `json:"lookalikeExpansion,omitempty"`  // Expansion factor (1-10, higher = broader reach, lower = more similar)
	SimilarityThreshold float64  `json:"similarityThreshold,omitempty"` // Min similarity score to include (0-1, default: 0.7)
	LookalikeBoost      float64  `json:"lookalikeBoost,omitempty"`      // Bid multiplier for lookalike matches (default: 1.3)
	LookalikeFeatures   []string `json:"lookalikeFeatures,omitempty"`   // Features for similarity: "demographics", "interests", "behavior", "geo", "device"
	// Audience Suppression
	SuppressionEnabled    bool     `json:"suppressionEnabled,omitempty"`    // Enable audience suppression
	SuppressionSegments   []string `json:"suppressionSegments,omitempty"`   // Segments to suppress (e.g., existing customers)
	SuppressionEvents     []string `json:"suppressionEvents,omitempty"`     // Events that trigger suppression (e.g., "purchase", "signup")
	SuppressionWindowDays int      `json:"suppressionWindowDays,omitempty"` // How many days to suppress (default: 30)
	// Audience Scoring
	ScoringEnabled  bool            `json:"scoringEnabled,omitempty"`  // Enable propensity scoring
	ScoringModel    string          `json:"scoringModel,omitempty"`    // "propensity", "ltv", "churn_risk"
	MinScore        float64         `json:"minScore,omitempty"`        // Minimum score to bid on (0-1)
	ScoreBidMapping []ScoreBidRange `json:"scoreBidMapping,omitempty"` // Score-to-bid multiplier mapping
}

// ScoreBidRange maps an audience score range to a bid multiplier
type ScoreBidRange struct {
	MinScore   float64 `json:"minScore"`   // Minimum score (inclusive)
	MaxScore   float64 `json:"maxScore"`   // Maximum score (exclusive)
	Multiplier float64 `json:"multiplier"` // Bid multiplier for this range
}

// AudienceModelingResult represents the result of audience modeling evaluation
type AudienceModelingResult struct {
	Matched         bool    `json:"matched"`
	Suppressed      bool    `json:"suppressed"`
	Multiplier      float64 `json:"multiplier"`
	IsLookalike     bool    `json:"isLookalike,omitempty"`
	SimilarityScore float64 `json:"similarityScore,omitempty"`
	PropensityScore float64 `json:"propensityScore,omitempty"`
	AudienceTier    string  `json:"audienceTier,omitempty"` // "seed", "lookalike_high", "lookalike_medium", "prospecting"
	Reason          string  `json:"reason,omitempty"`
}

// ============================================================================
// BID LANDSCAPE ANALYSIS
// ============================================================================

// BidLandscape represents historical bid distribution analysis
type BidLandscape struct {
	Enabled           bool                `json:"enabled,omitempty"`
	AnalysisWindow    int                 `json:"analysisWindow,omitempty"`    // Hours of data to analyze
	MinSampleSize     int                 `json:"minSampleSize,omitempty"`     // Minimum bids for valid analysis
	Percentiles       []BidPercentile     `json:"percentiles,omitempty"`       // Win rate at various bid levels
	OptimalBidRange   *OptimalBidRange    `json:"optimalBidRange,omitempty"`   // Recommended bid range
	CompetitorDensity []CompetitorDensity `json:"competitorDensity,omitempty"` // Competitor concentration zones
}

// BidPercentile represents win probability at a bid percentile
type BidPercentile struct {
	Percentile  int     `json:"percentile"`  // 10, 25, 50, 75, 90
	BidPrice    float64 `json:"bidPrice"`    // Price at this percentile
	WinRate     float64 `json:"winRate"`     // Historical win rate
	AvgClearPrc float64 `json:"avgClearPrc"` // Average clearing price when winning
}

// OptimalBidRange suggests optimal bidding strategy
type OptimalBidRange struct {
	MinBid          float64 `json:"minBid"`          // Floor for competitive bids
	MaxBid          float64 `json:"maxBid"`          // Ceiling before diminishing returns
	SweetSpot       float64 `json:"sweetSpot"`       // Best value bid
	ExpectedWinRate float64 `json:"expectedWinRate"` // Predicted win rate at sweet spot
	ExpectedCPM     float64 `json:"expectedCPM"`     // Predicted effective CPM
}

// CompetitorDensity shows where competitors concentrate bids
type CompetitorDensity struct {
	PriceFloor float64 `json:"priceFloor"`
	PriceCeil  float64 `json:"priceCeil"`
	Density    float64 `json:"density"`    // 0-1 concentration
	Aggressive bool    `json:"aggressive"` // High competition zone
}

// BidLandscapeResult contains analysis results
type BidLandscapeResult struct {
	Analyzed        bool             `json:"analyzed"`
	SampleSize      int              `json:"sampleSize"`
	RecommendedBid  float64          `json:"recommendedBid"`
	BidMultiplier   float64          `json:"bidMultiplier"`
	Confidence      float64          `json:"confidence"` // 0-1 confidence in recommendation
	OptimalRange    *OptimalBidRange `json:"optimalRange,omitempty"`
	MarketCondition string           `json:"marketCondition"` // "soft", "competitive", "aggressive"
	Reason          string           `json:"reason,omitempty"`
}

// ============================================================================
// CREATIVE OPTIMIZATION
// ============================================================================

// CreativeOptimization configures automatic creative selection
type CreativeOptimization struct {
	Enabled             bool                    `json:"enabled,omitempty"`
	OptimizationGoal    string                  `json:"optimizationGoal,omitempty"`    // "ctr", "cvr", "engagement", "viewability"
	ExplorationRate     float64                 `json:"explorationRate,omitempty"`     // 0-1, rate of testing vs exploitation
	MinImpressions      int                     `json:"minImpressions,omitempty"`      // Min impressions before optimization
	CreativePool        []CreativeVariant       `json:"creativePool,omitempty"`        // Available creatives
	PlacementRules      []PlacementCreativeRule `json:"placementRules,omitempty"`      // Placement-specific rules
	AutoPause           bool                    `json:"autoPause,omitempty"`           // Auto-pause underperformers
	PauseThreshold      float64                 `json:"pauseThreshold,omitempty"`      // Performance threshold for pausing
	DynamicCreative     *DynamicCreativeConfig  `json:"dynamicCreative,omitempty"`     // DCO settings
	MultivariateTesting bool                    `json:"multivariateTesting,omitempty"` // Test creative elements
}

// CreativeVariant represents a creative variant for testing
type CreativeVariant struct {
	ID          string        `json:"id"`
	Name        string        `json:"name,omitempty"`
	Creative    *Creative     `json:"creative"`
	Weight      float64       `json:"weight,omitempty"`      // Selection weight (0-1)
	Performance *CreativePerf `json:"performance,omitempty"` // Historical performance
	Status      string        `json:"status,omitempty"`      // "active", "paused", "testing"
	Segments    []string      `json:"segments,omitempty"`    // Target audience segments
	Placements  []string      `json:"placements,omitempty"`  // Preferred placements
}

// CreativePerf tracks creative performance metrics
type CreativePerf struct {
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Conversions    int64   `json:"conversions"`
	CTR            float64 `json:"ctr"`
	CVR            float64 `json:"cvr"`
	Viewability    float64 `json:"viewability"`
	EngagementRate float64 `json:"engagementRate"`
	AvgTimeViewed  float64 `json:"avgTimeViewed"`
	Score          float64 `json:"score"` // Composite optimization score
}

// PlacementCreativeRule maps placements to optimal creatives
type PlacementCreativeRule struct {
	PlacementType string   `json:"placementType"` // "banner", "video", "native", "interstitial"
	Sizes         []string `json:"sizes,omitempty"`
	CreativeIDs   []string `json:"creativeIds"`
	Boost         float64  `json:"boost,omitempty"` // Performance boost for this combo
}

// DynamicCreativeConfig configures Dynamic Creative Optimization
type DynamicCreativeConfig struct {
	Enabled       bool     `json:"enabled,omitempty"`
	Headlines     []string `json:"headlines,omitempty"`
	Descriptions  []string `json:"descriptions,omitempty"`
	Images        []string `json:"images,omitempty"`
	CTAs          []string `json:"ctas,omitempty"`
	Personalize   bool     `json:"personalize,omitempty"`   // User-level personalization
	WeatherBased  bool     `json:"weatherBased,omitempty"`  // Weather-responsive creative
	LocationBased bool     `json:"locationBased,omitempty"` // Location-based content
}

// CreativeOptimizationResult contains creative selection results
type CreativeOptimizationResult struct {
	SelectedCreativeID string   `json:"selectedCreativeId"`
	SelectionMethod    string   `json:"selectionMethod"` // "optimized", "exploration", "rule_based"
	PredictedCTR       float64  `json:"predictedCtr,omitempty"`
	PredictedCVR       float64  `json:"predictedCvr,omitempty"`
	Confidence         float64  `json:"confidence"`
	AlternativeIDs     []string `json:"alternativeIds,omitempty"`
	Reason             string   `json:"reason,omitempty"`
}

// ============================================================================
// INCREMENTALITY TESTING
// ============================================================================

// IncrementalityConfig configures lift measurement experiments
type IncrementalityConfig struct {
	Enabled           bool     `json:"enabled,omitempty"`
	ExperimentID      string   `json:"experimentId,omitempty"`
	ExperimentName    string   `json:"experimentName,omitempty"`
	ControlPercent    float64  `json:"controlPercent,omitempty"`    // % of users in control (no ads)
	HoldoutType       string   `json:"holdoutType,omitempty"`       // "user", "geo", "time"
	GeoHoldouts       []string `json:"geoHoldouts,omitempty"`       // Holdout regions
	ConversionWindow  int      `json:"conversionWindow,omitempty"`  // Days to track conversions
	MinSampleSize     int      `json:"minSampleSize,omitempty"`     // Min users per group
	ConfidenceLevel   float64  `json:"confidenceLevel,omitempty"`   // Required statistical confidence
	Metrics           []string `json:"metrics,omitempty"`           // Metrics to measure lift
	SegmentBreakdowns []string `json:"segmentBreakdowns,omitempty"` // Segments for sub-analysis
}

// IncrementalityResult contains experiment results
type IncrementalityResult struct {
	ExperimentID       string        `json:"experimentId"`
	Status             string        `json:"status"` // "running", "complete", "insufficient_data"
	ControlGroupSize   int           `json:"controlGroupSize"`
	TestGroupSize      int           `json:"testGroupSize"`
	Lift               float64       `json:"lift"`               // Incremental lift %
	LiftConfidence     float64       `json:"liftConfidence"`     // Statistical confidence
	IncrementalConv    int           `json:"incrementalConv"`    // Additional conversions from ads
	IncrementalRevenue float64       `json:"incrementalRevenue"` // Additional revenue
	ROAS               float64       `json:"roas"`               // Incremental ROAS
	IROAS              float64       `json:"iroas"`              // Incremental ROAS
	SegmentLifts       []SegmentLift `json:"segmentLifts,omitempty"`
	Recommendation     string        `json:"recommendation,omitempty"`
	UserInControlGroup bool          `json:"userInControlGroup"` // For bid-time decisions
}

// SegmentLift shows lift by segment
type SegmentLift struct {
	Segment    string  `json:"segment"`
	Lift       float64 `json:"lift"`
	Confidence float64 `json:"confidence"`
	SampleSize int     `json:"sampleSize"`
}

// ============================================================================
// PRIVACY SANDBOX SUPPORT
// ============================================================================

// PrivacySandbox configures Privacy Sandbox API integration
type PrivacySandbox struct {
	Enabled            bool                  `json:"enabled,omitempty"`
	TopicsAPI          *TopicsAPIConfig      `json:"topicsApi,omitempty"`
	AttributionAPI     *AttributionAPIConfig `json:"attributionApi,omitempty"`
	FledgeEnabled      bool                  `json:"fledgeEnabled,omitempty"` // Protected Audience API
	SharedStorage      bool                  `json:"sharedStorage,omitempty"`
	PrivateAggregation bool                  `json:"privateAggregation,omitempty"`
	FallbackStrategy   string                `json:"fallbackStrategy,omitempty"` // "contextual", "cohort", "none"
}

// TopicsAPIConfig configures Topics API usage
type TopicsAPIConfig struct {
	Enabled            bool            `json:"enabled,omitempty"`
	TopicTaxonomy      string          `json:"topicTaxonomy,omitempty"`      // "iab", "chrome"
	TargetTopics       []int           `json:"targetTopics,omitempty"`       // Topic IDs to target
	ExcludeTopics      []int           `json:"excludeTopics,omitempty"`      // Topic IDs to exclude
	TopicBidBoosts     []TopicBidBoost `json:"topicBidBoosts,omitempty"`     // Bid adjustments per topic
	MinTopicConfidence float64         `json:"minTopicConfidence,omitempty"` // Min confidence to use topic
}

// TopicBidBoost adjusts bids based on Topics API
type TopicBidBoost struct {
	TopicID    int     `json:"topicId"`
	TopicName  string  `json:"topicName,omitempty"`
	Multiplier float64 `json:"multiplier"`
}

// AttributionAPIConfig configures Attribution Reporting API
type AttributionAPIConfig struct {
	Enabled            bool     `json:"enabled,omitempty"`
	ReportingOrigin    string   `json:"reportingOrigin,omitempty"`
	SourceEventID      string   `json:"sourceEventId,omitempty"`
	TriggerData        []int    `json:"triggerData,omitempty"`
	AggregatableValues []string `json:"aggregatableValues,omitempty"`
	DebugMode          bool     `json:"debugMode,omitempty"`
}

// PrivacySandboxResult contains Privacy Sandbox evaluation
type PrivacySandboxResult struct {
	TopicsAvailable    bool    `json:"topicsAvailable"`
	UserTopics         []int   `json:"userTopics,omitempty"`
	TopicMatch         bool    `json:"topicMatch"`
	TopicMultiplier    float64 `json:"topicMultiplier"`
	AttributionEnabled bool    `json:"attributionEnabled"`
	FledgeEligible     bool    `json:"fledgeEligible"`
	FallbackUsed       bool    `json:"fallbackUsed"`
	FallbackMethod     string  `json:"fallbackMethod,omitempty"`
	Reason             string  `json:"reason,omitempty"`
}

// ============================================================================
// CONTEXTUAL AI
// ============================================================================

// ContextualAI configures ML-based contextual analysis
type ContextualAI struct {
	Enabled            bool                 `json:"enabled,omitempty"`
	ModelVersion       string               `json:"modelVersion,omitempty"`
	AnalyzeContent     bool                 `json:"analyzeContent,omitempty"`    // Page content analysis
	AnalyzeSentiment   bool                 `json:"analyzeSentiment,omitempty"`  // Sentiment detection
	AnalyzeEntities    bool                 `json:"analyzeEntities,omitempty"`   // Named entity recognition
	AnalyzeEmotion     bool                 `json:"analyzeEmotion,omitempty"`    // Emotional tone
	TargetCategories   []ContextualCategory `json:"targetCategories,omitempty"`  // Categories to target
	ExcludeCategories  []ContextualCategory `json:"excludeCategories,omitempty"` // Categories to exclude
	SentimentTargeting *SentimentTargeting  `json:"sentimentTargeting,omitempty"`
	EntityTargeting    []EntityTarget       `json:"entityTargeting,omitempty"`
	SemanticTargeting  *SemanticTargeting   `json:"semanticTargeting,omitempty"`
	MinConfidence      float64              `json:"minConfidence,omitempty"` // Min AI confidence
}

// ContextualCategory represents an IAB or custom category
type ContextualCategory struct {
	ID         string  `json:"id"`
	Name       string  `json:"name,omitempty"`
	Taxonomy   string  `json:"taxonomy,omitempty"` // "iab", "custom"
	Confidence float64 `json:"confidence,omitempty"`
	Multiplier float64 `json:"multiplier,omitempty"`
}

// SentimentTargeting configures sentiment-based targeting
type SentimentTargeting struct {
	TargetPositive    bool    `json:"targetPositive,omitempty"`
	TargetNeutral     bool    `json:"targetNeutral,omitempty"`
	TargetNegative    bool    `json:"targetNegative,omitempty"`
	PositiveBoost     float64 `json:"positiveBoost,omitempty"`
	NegativePenalty   float64 `json:"negativePenalty,omitempty"`
	MinSentimentScore float64 `json:"minSentimentScore,omitempty"` // -1 to 1
}

// EntityTarget targets specific named entities
type EntityTarget struct {
	EntityType string   `json:"entityType"` // "person", "organization", "location", "product", "event"
	Entities   []string `json:"entities"`
	Multiplier float64  `json:"multiplier,omitempty"`
	Exclude    bool     `json:"exclude,omitempty"`
}

// SemanticTargeting uses semantic similarity matching
type SemanticTargeting struct {
	Enabled             bool     `json:"enabled,omitempty"`
	SeedContent         []string `json:"seedContent,omitempty"`         // Reference content
	SimilarityThreshold float64  `json:"similarityThreshold,omitempty"` // 0-1
	UseEmbeddings       bool     `json:"useEmbeddings,omitempty"`       // Use vector embeddings
}

// ContextualAIResult contains AI analysis results
type ContextualAIResult struct {
	Analyzed       bool                 `json:"analyzed"`
	Categories     []ContextualCategory `json:"categories,omitempty"`
	Sentiment      string               `json:"sentiment,omitempty"` // "positive", "neutral", "negative"
	SentimentScore float64              `json:"sentimentScore"`      // -1 to 1
	Entities       []DetectedEntity     `json:"entities,omitempty"`
	Emotion        string               `json:"emotion,omitempty"` // "joy", "trust", "fear", etc.
	ContentQuality float64              `json:"contentQuality"`    // 0-1
	BrandSafe      bool                 `json:"brandSafe"`
	SemanticMatch  float64              `json:"semanticMatch,omitempty"` // 0-1 similarity
	BidMultiplier  float64              `json:"bidMultiplier"`
	Confidence     float64              `json:"confidence"`
	Reason         string               `json:"reason,omitempty"`
}

// DetectedEntity represents an entity found in content
type DetectedEntity struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Sentiment  string  `json:"sentiment,omitempty"`
}

// ============================================================================
// REAL-TIME ALERTS
// ============================================================================

// AlertConfig configures real-time alerting
type AlertConfig struct {
	Enabled              bool               `json:"enabled,omitempty"`
	BudgetAlerts         *BudgetAlerts      `json:"budgetAlerts,omitempty"`
	PerformanceAlerts    *PerformanceAlerts `json:"performanceAlerts,omitempty"`
	AnomalyDetection     *AnomalyDetection  `json:"anomalyDetection,omitempty"`
	PacingAlerts         *PacingAlerts      `json:"pacingAlerts,omitempty"`
	NotificationChannels []string           `json:"notificationChannels,omitempty"` // "email", "slack", "webhook"
	AlertCooldown        int                `json:"alertCooldown,omitempty"`        // Minutes between alerts
}

// BudgetAlerts configures budget-related alerts
type BudgetAlerts struct {
	Enabled            bool    `json:"enabled,omitempty"`
	WarnAtPercent      float64 `json:"warnAtPercent,omitempty"`      // Warn when X% spent
	CriticalAtPercent  float64 `json:"criticalAtPercent,omitempty"`  // Critical when X% spent
	ProjectedOverspend bool    `json:"projectedOverspend,omitempty"` // Alert on projected overspend
	UnexpectedSpike    bool    `json:"unexpectedSpike,omitempty"`    // Alert on spend spikes
	SpikeThreshold     float64 `json:"spikeThreshold,omitempty"`     // % increase to trigger
}

// PerformanceAlerts configures performance-related alerts
type PerformanceAlerts struct {
	Enabled            bool    `json:"enabled,omitempty"`
	CTRDropPercent     float64 `json:"ctrDropPercent,omitempty"`     // Alert on CTR drop
	CVRDropPercent     float64 `json:"cvrDropPercent,omitempty"`     // Alert on CVR drop
	CPAIncreasePercent float64 `json:"cpaIncreasePercent,omitempty"` // Alert on CPA increase
	WinRateDropPercent float64 `json:"winRateDropPercent,omitempty"` // Alert on win rate drop
	ViewabilityDrop    float64 `json:"viewabilityDrop,omitempty"`    // Alert on viewability drop
	ComparisonWindow   int     `json:"comparisonWindow,omitempty"`   // Hours to compare against
}

// AnomalyDetection configures ML-based anomaly alerts
type AnomalyDetection struct {
	Enabled          bool     `json:"enabled,omitempty"`
	Sensitivity      string   `json:"sensitivity,omitempty"` // "low", "medium", "high"
	MetricsToMonitor []string `json:"metricsToMonitor,omitempty"`
	AutoPause        bool     `json:"autoPause,omitempty"`      // Auto-pause on severe anomaly
	LearningPeriod   int      `json:"learningPeriod,omitempty"` // Days for baseline
}

// PacingAlerts configures delivery pacing alerts
type PacingAlerts struct {
	Enabled            bool    `json:"enabled,omitempty"`
	UnderPacingPercent float64 `json:"underPacingPercent,omitempty"` // Alert if under-pacing
	OverPacingPercent  float64 `json:"overPacingPercent,omitempty"`  // Alert if over-pacing
	CheckFrequency     int     `json:"checkFrequency,omitempty"`     // Minutes between checks
}

// Alert represents an active alert
type Alert struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`     // "budget", "performance", "anomaly", "pacing"
	Severity     string    `json:"severity"` // "info", "warning", "critical"
	CampaignID   string    `json:"campaignId"`
	CampaignName string    `json:"campaignName,omitempty"`
	Message      string    `json:"message"`
	Metric       string    `json:"metric,omitempty"`
	CurrentValue float64   `json:"currentValue,omitempty"`
	Threshold    float64   `json:"threshold,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Acknowledged bool      `json:"acknowledged"`
	AutoAction   string    `json:"autoAction,omitempty"` // Action taken if any
}

// AlertResult for bid-time alert checks
type AlertResult struct {
	HasActiveAlerts  bool     `json:"hasActiveAlerts"`
	CriticalAlerts   int      `json:"criticalAlerts"`
	ShouldPauseBid   bool     `json:"shouldPauseBid"`
	BidAdjustment    float64  `json:"bidAdjustment"`
	ActiveAlertTypes []string `json:"activeAlertTypes,omitempty"`
	Reason           string   `json:"reason,omitempty"`
}

// ============================================================================
// COMPETITIVE INTELLIGENCE
// ============================================================================

// CompetitiveIntelligence configures competitor analysis
type CompetitiveIntelligence struct {
	Enabled           bool              `json:"enabled,omitempty"`
	TrackCompetitors  []string          `json:"trackCompetitors,omitempty"` // Competitor advertiser IDs/domains
	AnalyzeWinLoss    bool              `json:"analyzeWinLoss,omitempty"`   // Win/loss analysis
	AnalyzeCreatives  bool              `json:"analyzeCreatives,omitempty"` // Track competitor creatives
	AnalyzeBidding    bool              `json:"analyzeBidding,omitempty"`   // Bid pattern analysis
	MarketShareGoal   float64           `json:"marketShareGoal,omitempty"`  // Target share of voice
	CompetitiveMode   string            `json:"competitiveMode,omitempty"`  // "aggressive", "balanced", "defensive"
	AutoAdjust        bool              `json:"autoAdjust,omitempty"`       // Auto-adjust to competition
	IntelligenceRules []CompetitiveRule `json:"intelligenceRules,omitempty"`
}

// CompetitiveRule defines response to competitor activity
type CompetitiveRule struct {
	CompetitorID  string  `json:"competitorId,omitempty"`
	Condition     string  `json:"condition"` // "outbid", "high_activity", "new_creative"
	Response      string  `json:"response"`  // "increase_bid", "match", "ignore"
	MaxMultiplier float64 `json:"maxMultiplier,omitempty"`
}

// CompetitorProfile represents intelligence on a competitor
type CompetitorProfile struct {
	ID             string    `json:"id"`
	Name           string    `json:"name,omitempty"`
	Domain         string    `json:"domain,omitempty"`
	EstimatedSpend float64   `json:"estimatedSpend,omitempty"`
	ShareOfVoice   float64   `json:"shareOfVoice,omitempty"`
	AvgBidPrice    float64   `json:"avgBidPrice,omitempty"`
	WinRateAgainst float64   `json:"winRateAgainst,omitempty"` // Our win rate vs them
	TopPlacements  []string  `json:"topPlacements,omitempty"`
	BiddingPattern string    `json:"biddingPattern,omitempty"` // "aggressive", "conservative"
	PeakHours      []int     `json:"peakHours,omitempty"`
	CreativeCount  int       `json:"creativeCount,omitempty"`
	LastSeen       time.Time `json:"lastSeen,omitempty"`
}

// CompetitiveIntelResult contains competitive analysis results
type CompetitiveIntelResult struct {
	Analyzed           bool                `json:"analyzed"`
	CompetitorsActive  int                 `json:"competitorsActive"`
	MarketCondition    string              `json:"marketCondition"` // "low", "medium", "high" competition
	OurShareOfVoice    float64             `json:"ourShareOfVoice"`
	LeadingCompetitor  string              `json:"leadingCompetitor,omitempty"`
	RecommendedAction  string              `json:"recommendedAction"`
	BidAdjustment      float64             `json:"bidAdjustment"`
	CompetitorProfiles []CompetitorProfile `json:"competitorProfiles,omitempty"`
	Reason             string              `json:"reason,omitempty"`
}

// ============================================================================
// UNIFIED ID SUPPORT
// ============================================================================

// UnifiedIDConfig configures cross-platform identity resolution
type UnifiedIDConfig struct {
	Enabled         bool         `json:"enabled,omitempty"`
	Providers       []IDProvider `json:"providers,omitempty"`
	FallbackOrder   []string     `json:"fallbackOrder,omitempty"`   // Priority order for ID resolution
	EnrichProfiles  bool         `json:"enrichProfiles,omitempty"`  // Enrich with provider data
	CrossDeviceSync bool         `json:"crossDeviceSync,omitempty"` // Sync across devices
	ConsentRequired bool         `json:"consentRequired,omitempty"` // Require explicit consent
	IDGraphEnabled  bool         `json:"idGraphEnabled,omitempty"`  // Use identity graph
	MatchRateTarget float64      `json:"matchRateTarget,omitempty"` // Target match rate
}

// IDProvider represents a unified ID provider
type IDProvider struct {
	Name      string  `json:"name"` // "uid2", "id5", "rampid", "liveramp", "zeotap"
	Enabled   bool    `json:"enabled,omitempty"`
	Priority  int     `json:"priority,omitempty"`
	Endpoint  string  `json:"endpoint,omitempty"`
	APIKey    string  `json:"apiKey,omitempty"`
	MatchRate float64 `json:"matchRate,omitempty"` // Historical match rate
	BidBoost  float64 `json:"bidBoost,omitempty"`  // Bid boost when matched
	Timeout   int     `json:"timeout,omitempty"`   // MS timeout
}

// UnifiedID represents a resolved identity
type UnifiedID struct {
	Provider      string            `json:"provider"`
	ID            string            `json:"id"`
	Confidence    float64           `json:"confidence"`
	LinkedIDs     []LinkedID        `json:"linkedIds,omitempty"`
	Segments      []string          `json:"segments,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	ConsentStatus string            `json:"consentStatus,omitempty"` // "granted", "denied", "unknown"
	LastRefreshed time.Time         `json:"lastRefreshed,omitempty"`
}

// LinkedID represents a linked identity from ID graph
type LinkedID struct {
	Provider   string  `json:"provider"`
	ID         string  `json:"id"`
	DeviceType string  `json:"deviceType,omitempty"`
	Confidence float64 `json:"confidence"`
}

// UnifiedIDResult contains identity resolution results
type UnifiedIDResult struct {
	Resolved         bool        `json:"resolved"`
	PrimaryID        *UnifiedID  `json:"primaryId,omitempty"`
	AlternateIDs     []UnifiedID `json:"alternateIds,omitempty"`
	DeviceCount      int         `json:"deviceCount"`
	MatchedProviders []string    `json:"matchedProviders,omitempty"`
	EnrichedProfile  bool        `json:"enrichedProfile"`
	BidMultiplier    float64     `json:"bidMultiplier"`
	AudienceSegments []string    `json:"audienceSegments,omitempty"`
	HasConsent       bool        `json:"hasConsent"`
	Reason           string      `json:"reason,omitempty"`
}
