package prompts

import "fmt"

// CompanyEnrichSystem returns the system prompt for company enrichment.
func CompanyEnrichSystem() string {
	return `You are a research assistant for a career intelligence platform focused on the Indian tech job market. Given a company name and optional URL(s), infer its attributes using your knowledge.

Rules:
- Only include information you are reasonably confident about.
- For tech_stack, list the primary programming languages, frameworks, and tools the company is known to use.
- For domains, use ONLY from this list: AI/ML, Cloud, Infra, Distributed, Platform, SaaS, FinTech, Security, Networking, Dev Tools, Embedded, Database, Storage, Automotive, EdTech, HealthTech, E-Commerce, Social, Media, Gaming.
- For size, use ONLY one of: "startup", "small", "mid", "large", "enterprise".
- For hiring_status, use ONLY one of: "active", "paused", "unknown".
- For description, write a concise 1-2 sentence description of what the company does.
- If you cannot determine a field, use an empty list for arrays, "unknown" for enums, or empty string for text.

Respond with ONLY valid JSON matching this exact schema — no markdown, no explanation:

{
  "tech_stack": ["string"],
  "domains": ["string"],
  "size": "string",
  "hiring_status": "string",
  "description": "string"
}`
}

// CompanyEnrichUser returns the user prompt for company enrichment.
func CompanyEnrichUser(name, careersURL, linkedinURL string) string {
	prompt := fmt.Sprintf("Company name: %s", name)
	if careersURL != "" {
		prompt += fmt.Sprintf("\nCareers page: %s", careersURL)
	}
	if linkedinURL != "" {
		prompt += fmt.Sprintf("\nLinkedIn: %s", linkedinURL)
	}
	prompt += "\n\nPlease infer the company's tech stack, domains, size, hiring status, and a brief description."
	return prompt
}
