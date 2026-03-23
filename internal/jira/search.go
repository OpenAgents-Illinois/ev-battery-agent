package jira

import "fmt"

// findExistingTicket searches for an open (non-Done) ticket whose summary contains the VIN.
// Returns the ticket key if found, empty string otherwise.
func (c *Client) findExistingTicket(vin string) string {
	jql := fmt.Sprintf(`project = "%s" AND summary ~ "%s" AND statusCategory != Done`, c.projectKey, vin)

	issues, _, err := c.api.Issue.Search(jql, &jiraSearchOptions)
	if err != nil || len(issues) == 0 {
		return ""
	}
	return issues[0].Key
}
