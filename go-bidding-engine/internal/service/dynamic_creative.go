package service

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taskirx/go-bidding-engine/internal/cache"
)

// DynamicCreativeService provides dynamic creative optimization (DCO)
// for personalized ad content generation
type DynamicCreativeService struct {
	cache           cache.Cache
	templates       sync.Map // templateID -> *CreativeTemplate
	elements        sync.Map // elementID -> *CreativeElement
	combinations    sync.Map // combinationKey -> *CreativeCombination
	userPreferences sync.Map // userID -> *UserCreativePreference
	performance     sync.Map // combinationID -> *CombinationPerformance
	mu              sync.RWMutex
	config          DCOConfig
}

// DCOConfig holds configuration for dynamic creative optimization
type DCOConfig struct {
	MaxElementsPerSlot     int     `json:"max_elements_per_slot"`
	ExplorationRate        float64 `json:"exploration_rate"`
	MinImpressionsForStats int64   `json:"min_impressions_for_stats"`
	PersonalizationWeight  float64 `json:"personalization_weight"`
	ContextWeight          float64 `json:"context_weight"`
	PerformanceWeight      float64 `json:"performance_weight"`
	EnableAutoOptimization bool    `json:"enable_auto_optimization"`
}

// CreativeTemplate represents a creative template with slots for dynamic elements
type CreativeTemplate struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Format      string                   `json:"format"` // banner, video, native
	Dimensions  Dimensions               `json:"dimensions"`
	Slots       map[string]*TemplateSlot `json:"slots"`
	BaseHTML    string                   `json:"base_html"`
	BaseCSS     string                   `json:"base_css"`
	Rules       []PersonalizationRule    `json:"rules"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// Dimensions represents creative dimensions
type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TemplateSlot represents a slot in the template for dynamic content
type TemplateSlot struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"` // headline, image, cta, description, logo
	Required     bool     `json:"required"`
	MaxLength    int      `json:"max_length"`
	AllowedTypes []string `json:"allowed_types"`
	DefaultValue string   `json:"default_value"`
	Position     Position `json:"position"`
}

// Position represents element position in the creative
type Position struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Anchor string `json:"anchor"` // top-left, center, etc.
}

// PersonalizationRule defines rules for dynamic content selection
type PersonalizationRule struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Priority   int             `json:"priority"`
	Conditions []RuleCondition `json:"conditions"`
	Actions    []RuleAction    `json:"actions"`
	Enabled    bool            `json:"enabled"`
}

// RuleCondition defines a condition for rule evaluation
type RuleCondition struct {
	Field    string `json:"field"`    // user.segment, context.category, device.type
	Operator string `json:"operator"` // equals, contains, in, gt, lt
	Value    any    `json:"value"`
}

// RuleAction defines an action when rule conditions are met
type RuleAction struct {
	Type   string `json:"type"` // select_element, set_property, apply_style
	SlotID string `json:"slot_id"`
	Value  any    `json:"value"`
}

// CreativeElement represents a dynamic element that can fill a slot
type CreativeElement struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"` // headline, image, cta, etc.
	Content     string              `json:"content"`
	ImageURL    string              `json:"image_url,omitempty"`
	CTAText     string              `json:"cta_text,omitempty"`
	CTAColor    string              `json:"cta_color,omitempty"`
	Attributes  map[string]string   `json:"attributes"`
	Tags        []string            `json:"tags"`
	Segments    []string            `json:"segments"` // Target segments
	Performance *ElementPerformance `json:"performance"`
	CreatedAt   time.Time           `json:"created_at"`
}

// ElementPerformance tracks element performance
type ElementPerformance struct {
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Conversions    int64   `json:"conversions"`
	CTR            float64 `json:"ctr"`
	ConversionRate float64 `json:"conversion_rate"`
	Revenue        float64 `json:"revenue"`
	Score          float64 `json:"score"`
	mu             sync.Mutex
}

// CreativeCombination represents a specific combination of elements
type CreativeCombination struct {
	ID          string                  `json:"id"`
	TemplateID  string                  `json:"template_id"`
	Elements    map[string]string       `json:"elements"` // slotID -> elementID
	Hash        string                  `json:"hash"`
	Performance *CombinationPerformance `json:"performance"`
	CreatedAt   time.Time               `json:"created_at"`
}

// CombinationPerformance tracks combination performance
type CombinationPerformance struct {
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Conversions    int64   `json:"conversions"`
	CTR            float64 `json:"ctr"`
	ConversionRate float64 `json:"conversion_rate"`
	Revenue        float64 `json:"revenue"`
	UCBScore       float64 `json:"ucb_score"`
	mu             sync.Mutex
}

// UserCreativePreference stores user preferences for creative elements
type UserCreativePreference struct {
	UserID             string             `json:"user_id"`
	PreferredColors    []string           `json:"preferred_colors"`
	PreferredCTAs      []string           `json:"preferred_ctas"`
	EngagedElements    map[string]int     `json:"engaged_elements"` // elementID -> engagement count
	ConvertedElements  map[string]int     `json:"converted_elements"`
	ClickedHeadlines   []string           `json:"clicked_headlines"`
	ContextPreferences map[string]float64 `json:"context_preferences"`
	LastUpdated        time.Time          `json:"last_updated"`
}

// DCORequest represents a request to generate optimized creative
type DCORequest struct {
	TemplateID  string         `json:"template_id"`
	UserID      string         `json:"user_id"`
	Context     DCOContext     `json:"context"`
	Constraints DCOConstraints `json:"constraints"`
}

// DCOContext provides context for creative optimization
type DCOContext struct {
	PageCategory    string            `json:"page_category"`
	ContentKeywords []string          `json:"content_keywords"`
	UserSegments    []string          `json:"user_segments"`
	DeviceType      string            `json:"device_type"`
	TimeOfDay       string            `json:"time_of_day"`
	DayOfWeek       string            `json:"day_of_week"`
	GeoLocation     string            `json:"geo_location"`
	Weather         string            `json:"weather"`
	CustomData      map[string]string `json:"custom_data"`
}

// DCOConstraints defines constraints for creative generation
type DCOConstraints struct {
	RequiredElements  []string `json:"required_elements"`
	ExcludedElements  []string `json:"excluded_elements"`
	MaxCombinations   int      `json:"max_combinations"`
	PreferExploration bool     `json:"prefer_exploration"`
}

// DCOResponse represents the optimized creative response
type DCOResponse struct {
	CombinationID        string                      `json:"combination_id"`
	TemplateID           string                      `json:"template_id"`
	Elements             map[string]*CreativeElement `json:"elements"`
	RenderedHTML         string                      `json:"rendered_html"`
	PersonalizationScore float64                     `json:"personalization_score"`
	PredictedCTR         float64                     `json:"predicted_ctr"`
	SelectionMethod      string                      `json:"selection_method"`
	RulesApplied         []string                    `json:"rules_applied"`
}

// NewDynamicCreativeService creates a new DCO service
func NewDynamicCreativeService(c cache.Cache) *DynamicCreativeService {
	return &DynamicCreativeService{
		cache: c,
		config: DCOConfig{
			MaxElementsPerSlot:     10,
			ExplorationRate:        0.1,
			MinImpressionsForStats: 100,
			PersonalizationWeight:  0.3,
			ContextWeight:          0.3,
			PerformanceWeight:      0.4,
			EnableAutoOptimization: true,
		},
	}
}

// CreateTemplate creates a new creative template
func (s *DynamicCreativeService) CreateTemplate(template *CreativeTemplate) (*CreativeTemplate, error) {
	if template.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}

	if len(template.Slots) == 0 {
		return nil, fmt.Errorf("template must have at least one slot")
	}

	now := time.Now()
	template.ID = uuid.New().String()
	template.CreatedAt = now
	template.UpdatedAt = now

	s.templates.Store(template.ID, template)
	return template, nil
}

// GetTemplate retrieves a template by ID
func (s *DynamicCreativeService) GetTemplate(templateID string) (*CreativeTemplate, error) {
	val, ok := s.templates.Load(templateID)
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	return val.(*CreativeTemplate), nil
}

// CreateElement creates a new creative element
func (s *DynamicCreativeService) CreateElement(element *CreativeElement) (*CreativeElement, error) {
	if element.Type == "" {
		return nil, fmt.Errorf("element type is required")
	}

	element.ID = uuid.New().String()
	element.CreatedAt = time.Now()
	element.Performance = &ElementPerformance{}

	s.elements.Store(element.ID, element)
	return element, nil
}

// GetElement retrieves an element by ID
func (s *DynamicCreativeService) GetElement(elementID string) (*CreativeElement, error) {
	val, ok := s.elements.Load(elementID)
	if !ok {
		return nil, fmt.Errorf("element not found: %s", elementID)
	}
	return val.(*CreativeElement), nil
}

// GetElementsByType returns all elements of a specific type
func (s *DynamicCreativeService) GetElementsByType(elementType string) []*CreativeElement {
	var elements []*CreativeElement
	s.elements.Range(func(key, value any) bool {
		elem := value.(*CreativeElement)
		if elem.Type == elementType {
			elements = append(elements, elem)
		}
		return true
	})
	return elements
}

// GenerateOptimizedCreative generates an optimized creative for the given context
func (s *DynamicCreativeService) GenerateOptimizedCreative(req DCORequest) (*DCOResponse, error) {
	template, err := s.GetTemplate(req.TemplateID)
	if err != nil {
		return nil, err
	}

	// Get user preferences
	userPref := s.getUserPreference(req.UserID)

	// Select elements for each slot
	selectedElements := make(map[string]*CreativeElement)
	var rulesApplied []string

	for slotID, slot := range template.Slots {
		// Get candidate elements for this slot
		candidates := s.getCandidateElements(slot.Type, req.Context, req.Constraints)

		if len(candidates) == 0 {
			// Use default if available
			if slot.DefaultValue != "" {
				continue
			}
			if slot.Required {
				return nil, fmt.Errorf("no elements available for required slot: %s", slotID)
			}
			continue
		}

		// Score and select element
		selectedElement, method := s.selectElement(candidates, userPref, req.Context, template.Rules)
		selectedElements[slotID] = selectedElement

		if method != "performance" {
			rulesApplied = append(rulesApplied, method)
		}
	}

	// Create or get combination
	combination := s.getOrCreateCombination(template.ID, selectedElements)

	// Calculate scores
	personalScore := s.calculatePersonalizationScore(selectedElements, userPref)
	predictedCTR := s.predictCTR(combination, req.Context)

	// Render HTML
	renderedHTML := s.renderCreative(template, selectedElements)

	return &DCOResponse{
		CombinationID:        combination.ID,
		TemplateID:           template.ID,
		Elements:             selectedElements,
		RenderedHTML:         renderedHTML,
		PersonalizationScore: personalScore,
		PredictedCTR:         predictedCTR,
		SelectionMethod:      s.determineSelectionMethod(rulesApplied),
		RulesApplied:         rulesApplied,
	}, nil
}

// RecordImpression records an impression for a combination
func (s *DynamicCreativeService) RecordImpression(combinationID string) error {
	val, ok := s.combinations.Load(combinationID)
	if !ok {
		return fmt.Errorf("combination not found: %s", combinationID)
	}

	combo := val.(*CreativeCombination)
	combo.Performance.mu.Lock()
	combo.Performance.Impressions++
	combo.Performance.mu.Unlock()

	// Update element performances
	for _, elemID := range combo.Elements {
		if elemVal, exists := s.elements.Load(elemID); exists {
			elem := elemVal.(*CreativeElement)
			elem.Performance.mu.Lock()
			elem.Performance.Impressions++
			elem.Performance.mu.Unlock()
		}
	}

	return nil
}

// RecordClick records a click for a combination
func (s *DynamicCreativeService) RecordClick(combinationID, userID string) error {
	val, ok := s.combinations.Load(combinationID)
	if !ok {
		return fmt.Errorf("combination not found: %s", combinationID)
	}

	combo := val.(*CreativeCombination)
	combo.Performance.mu.Lock()
	combo.Performance.Clicks++
	if combo.Performance.Impressions > 0 {
		combo.Performance.CTR = float64(combo.Performance.Clicks) / float64(combo.Performance.Impressions)
	}
	combo.Performance.mu.Unlock()

	// Update element performances
	for _, elemID := range combo.Elements {
		if elemVal, exists := s.elements.Load(elemID); exists {
			elem := elemVal.(*CreativeElement)
			elem.Performance.mu.Lock()
			elem.Performance.Clicks++
			if elem.Performance.Impressions > 0 {
				elem.Performance.CTR = float64(elem.Performance.Clicks) / float64(elem.Performance.Impressions)
			}
			elem.Performance.mu.Unlock()
		}
	}

	// Update user preferences
	s.updateUserPreferenceOnClick(userID, combo)

	return nil
}

// RecordConversion records a conversion for a combination
func (s *DynamicCreativeService) RecordConversion(combinationID string, revenue float64) error {
	val, ok := s.combinations.Load(combinationID)
	if !ok {
		return fmt.Errorf("combination not found: %s", combinationID)
	}

	combo := val.(*CreativeCombination)
	combo.Performance.mu.Lock()
	combo.Performance.Conversions++
	combo.Performance.Revenue += revenue
	if combo.Performance.Impressions > 0 {
		combo.Performance.ConversionRate = float64(combo.Performance.Conversions) / float64(combo.Performance.Impressions)
	}
	combo.Performance.mu.Unlock()

	// Update element performances
	for _, elemID := range combo.Elements {
		if elemVal, exists := s.elements.Load(elemID); exists {
			elem := elemVal.(*CreativeElement)
			elem.Performance.mu.Lock()
			elem.Performance.Conversions++
			elem.Performance.Revenue += revenue / float64(len(combo.Elements))
			if elem.Performance.Impressions > 0 {
				elem.Performance.ConversionRate = float64(elem.Performance.Conversions) / float64(elem.Performance.Impressions)
			}
			elem.Performance.mu.Unlock()
		}
	}

	return nil
}

// GetTopCombinations returns the top performing combinations
func (s *DynamicCreativeService) GetTopCombinations(templateID string, limit int) []*CreativeCombination {
	var combinations []*CreativeCombination

	s.combinations.Range(func(key, value any) bool {
		combo := value.(*CreativeCombination)
		if combo.TemplateID == templateID {
			combinations = append(combinations, combo)
		}
		return true
	})

	// Sort by CTR
	sort.Slice(combinations, func(i, j int) bool {
		return combinations[i].Performance.CTR > combinations[j].Performance.CTR
	})

	if len(combinations) > limit {
		combinations = combinations[:limit]
	}

	return combinations
}

// GetDCOStats returns DCO statistics
func (s *DynamicCreativeService) GetDCOStats() map[string]any {
	stats := map[string]any{
		"total_templates":    0,
		"total_elements":     0,
		"total_combinations": 0,
		"total_impressions":  int64(0),
		"total_clicks":       int64(0),
		"total_conversions":  int64(0),
		"avg_ctr":            0.0,
		"elements_by_type":   make(map[string]int),
	}

	s.templates.Range(func(key, value any) bool {
		stats["total_templates"] = stats["total_templates"].(int) + 1
		return true
	})

	elementsByType := make(map[string]int)
	s.elements.Range(func(key, value any) bool {
		elem := value.(*CreativeElement)
		stats["total_elements"] = stats["total_elements"].(int) + 1
		elementsByType[elem.Type]++
		return true
	})
	stats["elements_by_type"] = elementsByType

	var totalImpressions, totalClicks, totalConversions int64
	s.combinations.Range(func(key, value any) bool {
		combo := value.(*CreativeCombination)
		stats["total_combinations"] = stats["total_combinations"].(int) + 1
		totalImpressions += combo.Performance.Impressions
		totalClicks += combo.Performance.Clicks
		totalConversions += combo.Performance.Conversions
		return true
	})

	stats["total_impressions"] = totalImpressions
	stats["total_clicks"] = totalClicks
	stats["total_conversions"] = totalConversions

	if totalImpressions > 0 {
		stats["avg_ctr"] = float64(totalClicks) / float64(totalImpressions)
	}

	return stats
}

// UpdateConfig updates DCO configuration
func (s *DynamicCreativeService) UpdateConfig(config DCOConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// GetConfig returns current DCO configuration
func (s *DynamicCreativeService) GetConfig() DCOConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// Helper functions

func (s *DynamicCreativeService) getUserPreference(userID string) *UserCreativePreference {
	if val, ok := s.userPreferences.Load(userID); ok {
		return val.(*UserCreativePreference)
	}

	// Create new preference record
	pref := &UserCreativePreference{
		UserID:             userID,
		EngagedElements:    make(map[string]int),
		ConvertedElements:  make(map[string]int),
		ContextPreferences: make(map[string]float64),
		LastUpdated:        time.Now(),
	}
	s.userPreferences.Store(userID, pref)
	return pref
}

func (s *DynamicCreativeService) getCandidateElements(elementType string, context DCOContext, constraints DCOConstraints) []*CreativeElement {
	var candidates []*CreativeElement

	excludeSet := make(map[string]bool)
	for _, id := range constraints.ExcludedElements {
		excludeSet[id] = true
	}

	s.elements.Range(func(key, value any) bool {
		elem := value.(*CreativeElement)
		if elem.Type != elementType {
			return true
		}
		if excludeSet[elem.ID] {
			return true
		}

		// Check segment match
		if len(elem.Segments) > 0 && len(context.UserSegments) > 0 {
			matched := false
			for _, seg := range elem.Segments {
				for _, userSeg := range context.UserSegments {
					if seg == userSeg {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				return true
			}
		}

		candidates = append(candidates, elem)
		return true
	})

	return candidates
}

func (s *DynamicCreativeService) selectElement(candidates []*CreativeElement, userPref *UserCreativePreference, context DCOContext, rules []PersonalizationRule) (*CreativeElement, string) {
	if len(candidates) == 0 {
		return nil, ""
	}

	// Check for exploration
	if rand.Float64() < s.config.ExplorationRate {
		return candidates[rand.Intn(len(candidates))], "exploration"
	}

	// Apply rules first
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if s.evaluateRule(rule, context) {
			for _, action := range rule.Actions {
				if action.Type == "select_element" {
					elemID := action.Value.(string)
					for _, c := range candidates {
						if c.ID == elemID {
							return c, rule.Name
						}
					}
				}
			}
		}
	}

	// Score-based selection
	type scoredElement struct {
		element *CreativeElement
		score   float64
	}

	scored := make([]scoredElement, len(candidates))
	for i, elem := range candidates {
		score := s.scoreElement(elem, userPref, context)
		scored[i] = scoredElement{element: elem, score: score}
	}

	// Sort by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return scored[0].element, "performance"
}

func (s *DynamicCreativeService) evaluateRule(rule PersonalizationRule, context DCOContext) bool {
	for _, cond := range rule.Conditions {
		if !s.evaluateCondition(cond, context) {
			return false
		}
	}
	return true
}

func (s *DynamicCreativeService) evaluateCondition(cond RuleCondition, context DCOContext) bool {
	var fieldValue any

	switch cond.Field {
	case "context.category":
		fieldValue = context.PageCategory
	case "device.type":
		fieldValue = context.DeviceType
	case "time.day":
		fieldValue = context.TimeOfDay
	case "geo.location":
		fieldValue = context.GeoLocation
	default:
		return false
	}

	switch cond.Operator {
	case "equals":
		return fieldValue == cond.Value
	case "contains":
		if s, ok := fieldValue.(string); ok {
			if v, ok := cond.Value.(string); ok {
				return s == v || len(s) > 0 && len(v) > 0
			}
		}
	case "in":
		if vals, ok := cond.Value.([]string); ok {
			for _, v := range vals {
				if v == fieldValue {
					return true
				}
			}
		}
	}

	return false
}

func (s *DynamicCreativeService) scoreElement(elem *CreativeElement, userPref *UserCreativePreference, context DCOContext) float64 {
	score := 0.0

	// Performance score (UCB-like)
	if elem.Performance.Impressions >= s.config.MinImpressionsForStats {
		performanceScore := elem.Performance.CTR
		exploration := math.Sqrt(2 * math.Log(float64(s.getTotalImpressions())) / float64(elem.Performance.Impressions))
		score += s.config.PerformanceWeight * (performanceScore + 0.1*exploration)
	} else {
		// Exploration bonus for new elements
		score += s.config.PerformanceWeight * 0.5
	}

	// Personalization score
	if engagements, ok := userPref.EngagedElements[elem.ID]; ok {
		score += s.config.PersonalizationWeight * math.Min(float64(engagements)/10.0, 1.0)
	}

	// Context match score
	for _, tag := range elem.Tags {
		for _, keyword := range context.ContentKeywords {
			if tag == keyword {
				score += s.config.ContextWeight * 0.2
			}
		}
	}

	return score
}

func (s *DynamicCreativeService) getTotalImpressions() int64 {
	var total int64
	s.combinations.Range(func(key, value any) bool {
		combo := value.(*CreativeCombination)
		total += combo.Performance.Impressions
		return true
	})
	if total < 1 {
		return 1
	}
	return total
}

func (s *DynamicCreativeService) getOrCreateCombination(templateID string, elements map[string]*CreativeElement) *CreativeCombination {
	// Create element map
	elemMap := make(map[string]string)
	for slotID, elem := range elements {
		elemMap[slotID] = elem.ID
	}

	// Create hash for combination
	hash := s.hashCombination(templateID, elemMap)

	// Check if exists
	if val, ok := s.combinations.Load(hash); ok {
		return val.(*CreativeCombination)
	}

	// Create new combination
	combo := &CreativeCombination{
		ID:          uuid.New().String(),
		TemplateID:  templateID,
		Elements:    elemMap,
		Hash:        hash,
		Performance: &CombinationPerformance{},
		CreatedAt:   time.Now(),
	}

	s.combinations.Store(hash, combo)
	s.combinations.Store(combo.ID, combo)

	return combo
}

func (s *DynamicCreativeService) hashCombination(templateID string, elements map[string]string) string {
	// Simple hash - could be improved
	hash := templateID
	var keys []string
	for k := range elements {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		hash += ":" + k + "=" + elements[k]
	}
	return hash
}

func (s *DynamicCreativeService) calculatePersonalizationScore(elements map[string]*CreativeElement, userPref *UserCreativePreference) float64 {
	if len(elements) == 0 {
		return 0
	}

	score := 0.0
	for _, elem := range elements {
		if engagements, ok := userPref.EngagedElements[elem.ID]; ok {
			score += math.Min(float64(engagements)/5.0, 1.0)
		}
		if conversions, ok := userPref.ConvertedElements[elem.ID]; ok {
			score += math.Min(float64(conversions)*2.0, 2.0)
		}
	}

	return math.Min(score/float64(len(elements)), 1.0)
}

func (s *DynamicCreativeService) predictCTR(combo *CreativeCombination, context DCOContext) float64 {
	// Base CTR from combination performance
	baseCTR := 0.01
	if combo.Performance.Impressions >= s.config.MinImpressionsForStats {
		baseCTR = combo.Performance.CTR
	}

	// Context adjustments
	adjustment := 1.0

	// Device type adjustment
	switch context.DeviceType {
	case "mobile":
		adjustment *= 1.1
	case "desktop":
		adjustment *= 0.95
	case "tablet":
		adjustment *= 1.0
	}

	// Time of day adjustment
	switch context.TimeOfDay {
	case "morning":
		adjustment *= 1.05
	case "evening":
		adjustment *= 1.1
	case "night":
		adjustment *= 0.9
	}

	return math.Min(baseCTR*adjustment, 1.0)
}

func (s *DynamicCreativeService) renderCreative(template *CreativeTemplate, elements map[string]*CreativeElement) string {
	// Simple rendering - in production this would be more sophisticated
	html := template.BaseHTML
	if html == "" {
		html = "<div class='dco-creative'>"
		for slotID, elem := range elements {
			html += fmt.Sprintf("<div class='slot-%s'>%s</div>", slotID, elem.Content)
		}
		html += "</div>"
	}
	return html
}

func (s *DynamicCreativeService) determineSelectionMethod(rulesApplied []string) string {
	if len(rulesApplied) > 0 {
		for _, rule := range rulesApplied {
			if rule == "exploration" {
				return "exploration"
			}
		}
		return "rule_based"
	}
	return "ml_optimized"
}

func (s *DynamicCreativeService) updateUserPreferenceOnClick(userID string, combo *CreativeCombination) {
	pref := s.getUserPreference(userID)

	for _, elemID := range combo.Elements {
		if pref.EngagedElements == nil {
			pref.EngagedElements = make(map[string]int)
		}
		pref.EngagedElements[elemID]++
	}
	pref.LastUpdated = time.Now()
}
