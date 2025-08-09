package i18n

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	// Global message catalog registry
	globalCatalog MessageCatalog = NewMemoryCatalog()
	
	// Parameter interpolation regex for {param} style
	interpolationRegex = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
)

// MemoryCatalog is an in-memory implementation of MessageCatalog.
type MemoryCatalog struct {
	mu       sync.RWMutex
	messages map[string]Message
}

// NewMemoryCatalog creates a new in-memory message catalog.
func NewMemoryCatalog() *MemoryCatalog {
	return &MemoryCatalog{
		messages: make(map[string]Message),
	}
}

// Register adds a message to the catalog.
func (c *MemoryCatalog) Register(msg Message) error {
	if msg.Key == "" {
		return &CatalogError{
			Operation: "register",
			Key:       msg.Key,
			Err:       fmt.Errorf("message key cannot be empty"),
		}
	}
	
	if msg.Default == "" {
		return &CatalogError{
			Operation: "register",
			Key:       msg.Key,
			Err:       fmt.Errorf("message default text cannot be empty"),
		}
	}
	
	// Auto-detect parameters if not specified
	if len(msg.Params) == 0 {
		validator := NewStandardParamValidator()
		msg.Params = validator.RequiredParams(msg.Default)
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Check for duplicate keys
	if _, exists := c.messages[msg.Key]; exists {
		return &CatalogError{
			Operation: "register",
			Key:       msg.Key,
			Err:       fmt.Errorf("message key already exists"),
		}
	}
	
	c.messages[msg.Key] = msg
	return nil
}

// RegisterBulk adds multiple messages to the catalog.
func (c *MemoryCatalog) RegisterBulk(messages []Message) error {
	var errors []error
	
	for _, msg := range messages {
		if err := c.Register(msg); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		var errStrs []string
		for _, err := range errors {
			errStrs = append(errStrs, err.Error())
		}
		return fmt.Errorf("bulk registration failed: %s", strings.Join(errStrs, "; "))
	}
	
	return nil
}

// Get retrieves a message by key.
func (c *MemoryCatalog) Get(key string) (Message, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	msg, exists := c.messages[key]
	return msg, exists
}

// Keys returns all registered message keys, sorted alphabetically.
func (c *MemoryCatalog) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keys := make([]string, 0, len(c.messages))
	for key := range c.messages {
		keys = append(keys, key)
	}
	
	sort.Strings(keys)
	return keys
}

// Count returns the number of registered messages.
func (c *MemoryCatalog) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.messages)
}

// Validate checks if all expected parameters are provided for a message.
func (c *MemoryCatalog) Validate(key string, params Params) error {
	msg, exists := c.Get(key)
	if !exists {
		return &CatalogError{
			Operation: "validate",
			Key:       key,
			Err:       fmt.Errorf("message not found in catalog"),
		}
	}
	
	validator := NewStandardParamValidator()
	if err := validator.ValidateParams(msg.Params, params); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			validationErr.Key = key
			return validationErr
		}
		return err
	}
	
	return nil
}

// Clear removes all messages from the catalog.
func (c *MemoryCatalog) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.messages = make(map[string]Message)
}

// Global catalog functions

// Register adds a message to the global catalog.
func Register(key, defaultText string, params ...string) error {
	msg := Message{
		Key:     key,
		Default: defaultText,
		Params:  params,
	}
	return globalCatalog.Register(msg)
}

// RegisterMessage adds a complete message to the global catalog.
func RegisterMessage(msg Message) error {
	return globalCatalog.Register(msg)
}

// RegisterMessages adds multiple messages to the global catalog.
func RegisterMessages(messages []Message) error {
	return globalCatalog.RegisterBulk(messages)
}

// Get retrieves a message from the global catalog.
func Get(key string) (Message, bool) {
	return globalCatalog.Get(key)
}

// Keys returns all keys from the global catalog.
func Keys() []string {
	return globalCatalog.Keys()
}

// Count returns the number of messages in the global catalog.
func Count() int {
	return globalCatalog.Count()
}

// Validate validates parameters against the global catalog.
func Validate(key string, params Params) error {
	return globalCatalog.Validate(key, params)
}

// Clear clears the global catalog (useful for testing).
func Clear() {
	globalCatalog.Clear()
}

// Message lookup and interpolation

// Lookup retrieves a message by key with optional parameter validation.
func Lookup(key string, params Params, strict bool) (Message, string, error) {
	msg, exists := Get(key)
	if !exists {
		return Message{}, "", &CatalogError{
			Operation: "lookup",
			Key:       key,
			Err:       fmt.Errorf("message not found"),
		}
	}
	
	// Validate parameters if strict mode is enabled
	if strict && len(msg.Params) > 0 {
		if err := Validate(key, params); err != nil {
			return msg, msg.Default, err
		}
	}
	
	// Interpolate parameters
	text, err := InterpolateParams(msg.Default, params)
	if err != nil {
		return msg, msg.Default, &InterpolationError{
			Key:      key,
			Template: msg.Default,
			Err:      err,
		}
	}
	
	return msg, text, nil
}

// InterpolateParams performs {param} style parameter interpolation.
func InterpolateParams(text string, params Params) (string, error) {
	if len(params) == 0 {
		return text, nil
	}
	
	// Track missing parameters for error reporting
	var missingParams []string
	
	result := interpolationRegex.ReplaceAllStringFunc(text, func(match string) string {
		// Extract parameter name (remove { and })
		paramName := match[1 : len(match)-1]
		
		if value, exists := params[paramName]; exists {
			return fmt.Sprintf("%v", value)
		}
		
		// Track missing parameter
		missingParams = append(missingParams, paramName)
		return match // Keep original placeholder
	})
	
	if len(missingParams) > 0 {
		return result, &InterpolationError{
			Template:     text,
			MissingParam: missingParams[0], // Report first missing param
		}
	}
	
	return result, nil
}

// Helper functions for message registration

// MustRegister registers a message and panics on error.
// Useful for init functions where registration errors should be fatal.
func MustRegister(key, defaultText string, params ...string) {
	if err := Register(key, defaultText, params...); err != nil {
		panic(fmt.Sprintf("failed to register message '%s': %v", key, err))
	}
}

// MustRegisterMessage registers a message and panics on error.
func MustRegisterMessage(msg Message) {
	if err := RegisterMessage(msg); err != nil {
		panic(fmt.Sprintf("failed to register message '%s': %v", msg.Key, err))
	}
}

// MustRegisterMessages registers multiple messages and panics on error.
func MustRegisterMessages(messages []Message) {
	if err := RegisterMessages(messages); err != nil {
		panic(fmt.Sprintf("failed to register messages: %v", err))
	}
}

// Utility functions

// HasMessage checks if a message key exists in the catalog.
func HasMessage(key string) bool {
	_, exists := Get(key)
	return exists
}

// GetDefault returns the default text for a message key.
func GetDefault(key string) string {
	msg, exists := Get(key)
	if !exists {
		return ""
	}
	return msg.Default
}

// GetParams returns the expected parameters for a message key.
func GetParams(key string) []string {
	msg, exists := Get(key)
	if !exists {
		return nil
	}
	return msg.Params
}

// SearchKeys searches for message keys containing the given substring.
func SearchKeys(substring string) []string {
	allKeys := Keys()
	var matches []string
	
	substring = strings.ToLower(substring)
	for _, key := range allKeys {
		if strings.Contains(strings.ToLower(key), substring) {
			matches = append(matches, key)
		}
	}
	
	return matches
}

// GroupByPrefix groups message keys by their prefix (part before first dot).
func GroupByPrefix() map[string][]string {
	allKeys := Keys()
	groups := make(map[string][]string)
	
	for _, key := range allKeys {
		parts := strings.SplitN(key, ".", 2)
		prefix := parts[0]
		
		groups[prefix] = append(groups[prefix], key)
	}
	
	// Sort each group
	for prefix := range groups {
		sort.Strings(groups[prefix])
	}
	
	return groups
}

// Stats returns statistics about the message catalog.
type CatalogStats struct {
	TotalMessages      int
	MessagesByPrefix   map[string]int
	MessagesWithParams int
	AverageParamCount  float64
	TopPrefixes        []string // Top 5 prefixes by message count
}

// GetStats returns statistics about the global catalog.
func GetStats() CatalogStats {
	allKeys := Keys()
	prefixes := make(map[string]int)
	totalParams := 0
	messagesWithParams := 0
	
	for _, key := range allKeys {
		// Count by prefix
		parts := strings.SplitN(key, ".", 2)
		prefix := parts[0]
		prefixes[prefix]++
		
		// Count parameters
		msg, _ := Get(key)
		paramCount := len(msg.Params)
		if paramCount > 0 {
			messagesWithParams++
			totalParams += paramCount
		}
	}
	
	// Calculate average params
	avgParams := 0.0
	if messagesWithParams > 0 {
		avgParams = float64(totalParams) / float64(messagesWithParams)
	}
	
	// Get top prefixes
	type prefixCount struct {
		prefix string
		count  int
	}
	
	var prefixCounts []prefixCount
	for prefix, count := range prefixes {
		prefixCounts = append(prefixCounts, prefixCount{prefix, count})
	}
	
	sort.Slice(prefixCounts, func(i, j int) bool {
		return prefixCounts[i].count > prefixCounts[j].count
	})
	
	topPrefixes := make([]string, 0, 5)
	for i, pc := range prefixCounts {
		if i >= 5 {
			break
		}
		topPrefixes = append(topPrefixes, pc.prefix)
	}
	
	return CatalogStats{
		TotalMessages:      len(allKeys),
		MessagesByPrefix:   prefixes,
		MessagesWithParams: messagesWithParams,
		AverageParamCount:  avgParams,
		TopPrefixes:        topPrefixes,
	}
}
