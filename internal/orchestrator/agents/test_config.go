package agents

import (
	"bufio"
	"os"
	"strings"
)

// LoadAPIConfig loads API configuration from api.txt for integration tests
func LoadAPIConfig() map[string]string {
	config := make(map[string]string)
	
	file, err := os.Open("../../../api.txt")
	if err != nil {
		return config
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			config[key] = value
		}
	}
	
	return config
}

// IsAPIConfigured checks if API is configured for integration tests
func IsAPIConfigured() bool {
	config := LoadAPIConfig()
	return config["OPENROUTER_API_KEY"] != "" || 
	       config["OPENAI_API_KEY"] != "" ||
	       config["ANTHROPIC_API_KEY"] != ""
}
