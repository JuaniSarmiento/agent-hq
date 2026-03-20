package profiles

import (
	"bufio"
	"embed"
	"fmt"
	"strings"
)

//go:embed agents/*.md
var AgentProfiles embed.FS

// ProfileInfo holds a profile's name and role description.
type ProfileInfo struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// List returns all available agent profiles with their name and role.
func List() ([]ProfileInfo, error) {
	entries, err := AgentProfiles.ReadDir("agents")
	if err != nil {
		return nil, fmt.Errorf("read profiles dir: %w", err)
	}

	var profiles []ProfileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		role, err := extractRole(entry.Name())
		if err != nil {
			role = ""
		}

		profiles = append(profiles, ProfileInfo{
			Name: name,
			Role: role,
		})
	}

	return profiles, nil
}

// Get returns the full markdown content of a profile by name.
func Get(name string) (string, error) {
	filename := name + ".md"
	data, err := AgentProfiles.ReadFile("agents/" + filename)
	if err != nil {
		return "", fmt.Errorf("profile %q not found: %w", name, err)
	}
	return string(data), nil
}

// extractRole reads the YAML frontmatter to find the role field.
func extractRole(filename string) (string, error) {
	data, err := AgentProfiles.ReadFile("agents/" + filename)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "role:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "role:")), nil
		}
	}

	return "", fmt.Errorf("no role found in %s", filename)
}
