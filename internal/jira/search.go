package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// findExistingTicket searches for an open (non-Done) ticket whose summary contains the VIN.
// Returns the ticket key if found, empty string otherwise.
func (c *Client) findExistingTicket(vin string) string {
	jql := fmt.Sprintf(`project = "%s" AND summary ~ "%s" AND statusCategory != Done`, c.projectKey, vin)
	apiURL := fmt.Sprintf("https://%s/rest/api/3/search?jql=%s&maxResults=1&fields=key,summary",
		c.domain, url.QueryEscape(jql))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Basic "+c.encodedAuth())
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	if total, _ := result["total"].(float64); total == 0 {
		return ""
	}
	issues, ok := result["issues"].([]any)
	if !ok || len(issues) == 0 {
		return ""
	}
	issue, ok := issues[0].(map[string]any)
	if !ok {
		return ""
	}
	key, _ := issue["key"].(string)
	return key
}
