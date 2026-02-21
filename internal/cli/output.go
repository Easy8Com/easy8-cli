package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"easy8-cli/internal/api"
)

func outputJSON(value any) int {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, "output error:", err)
		return 1
	}
	return 0
}

func outputIssues(issues []api.Issue) int {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSubject\tStatus\tAssignee\tUpdated")
	for _, issue := range issues {
		status := nameOrEmpty(issue.Status)
		assignee := nameOrEmpty(issue.AssignedTo)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", issue.ID, issue.Subject, status, assignee, issue.UpdatedOn)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "output error:", err)
		return 1
	}
	return 0
}

func outputSearch(results []api.SearchResult) int {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tType\tTitle\tURL")
	for _, result := range results {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", result.ID, result.Type, result.Title, result.URL)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "output error:", err)
		return 1
	}
	return 0
}

func nameOrEmpty(ref *api.NamedRef) string {
	if ref == nil {
		return ""
	}
	return ref.Name
}
