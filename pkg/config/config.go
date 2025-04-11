type Config struct {
	ActivePersona string               `json:"activePersona"`
	Personas      map[string]AIPersona `json:"personas"`
}

// AIPersona defines an AI assistant personality
type AIPersona struct {
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
}

// GetCurrentPersona returns the currently active persona
func (c *Config) GetCurrentPersona() AIPersona {
	// Check if active persona exists
	if persona, ok := c.Personas[c.ActivePersona]; ok {
		return persona
	}

	// Return default persona if active one doesn't exist
	return c.Personas["default"]
}

// ListPersonas returns all available personas
func (c *Config) ListPersonas() map[string]AIPersona {
	// Initialize personas if nil
	if c.Personas == nil {
		c.initDefaultPersonas()
	}

	return c.Personas
}

// SetPersona sets the active persona by name
func (c *Config) SetPersona(name string) error {
	// Check if persona exists
	if _, ok := c.Personas[name]; !ok {
		return fmt.Errorf("persona '%s' not found", name)
	}

	// Set active persona
	c.ActivePersona = name

	// Save config
	return c.Save()
}

// AddCustomPersona adds a new custom persona
func (c *Config) AddCustomPersona(name string, persona AIPersona) error {
	// Initialize personas if nil
	if c.Personas == nil {
		c.initDefaultPersonas()
	}

	// Check if name is reserved
	if name == "default" || name == "kubernetes-expert" || name == "devops-engineer" {
		return fmt.Errorf("cannot override built-in persona '%s'", name)
	}

	// Add persona
	c.Personas[name] = persona

	// Save config
	return c.Save()
}

// RemoveCustomPersona removes a custom persona
func (c *Config) RemoveCustomPersona(name string) error {
	// Check if persona exists
	if _, ok := c.Personas[name]; !ok {
		return fmt.Errorf("persona '%s' not found", name)
	}

	// Check if trying to remove built-in persona
	if name == "default" || name == "kubernetes-expert" || name == "devops-engineer" {
		return fmt.Errorf("cannot remove built-in persona '%s'", name)
	}

	// Check if trying to remove active persona
	if name == c.ActivePersona {
		return fmt.Errorf("cannot remove active persona, switch to another persona first")
	}

	// Remove persona
	delete(c.Personas, name)

	// Save config
	return c.Save()
}

// initDefaultPersonas initializes the default set of personas
func (c *Config) initDefaultPersonas() {
	if c.Personas == nil {
		c.Personas = make(map[string]AIPersona)
	}

	// Set default persona if not present
	if _, ok := c.Personas["default"]; !ok {
		c.Personas["default"] = AIPersona{
			Description:  "General assistant for Kubernetes operations",
			SystemPrompt: "You are a helpful AI assistant for Kubernetes operations. Provide concise, accurate responses to user queries about Kubernetes resources, configurations, and best practices.",
		}
	}

	// Add Kubernetes expert persona if not present
	if _, ok := c.Personas["kubernetes-expert"]; !ok {
		c.Personas["kubernetes-expert"] = AIPersona{
			Description:  "Specialized in Kubernetes architecture and advanced topics",
			SystemPrompt: "You are a Kubernetes expert with deep knowledge of architecture, networking, security, and advanced topics. Prioritize best practices and consider performance, security, and operational aspects in your responses.",
		}
	}

	// Add DevOps engineer persona if not present
	if _, ok := c.Personas["devops-engineer"]; !ok {
		c.Personas["devops-engineer"] = AIPersona{
			Description:  "Focused on DevOps practices and CI/CD integration",
			SystemPrompt: "You are a DevOps engineer specializing in Kubernetes CI/CD pipelines, GitOps workflows, and infrastructure automation. Focus on implementation details, practical advice, and automation strategies in your responses.",
		}
	}

	// Set active persona if not set
	if c.ActivePersona == "" {
		c.ActivePersona = "default"
	}
}

// Load loads configuration from file
func (c *Config) Load() error {
	// ... existing code ...

	// Initialize default personas if needed
	c.initDefaultPersonas()

	// ... existing code ...
} 