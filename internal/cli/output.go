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

func outputIssueDetail(issue api.Issue) int {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%d\n", issue.ID)
	fmt.Fprintf(w, "Subject:\t%s\n", issue.Subject)
	fmt.Fprintf(w, "Project:\t%s\n", nameOrEmpty(issue.Project))
	fmt.Fprintf(w, "Tracker:\t%s\n", nameOrEmpty(issue.Tracker))
	fmt.Fprintf(w, "Status:\t%s\n", nameOrEmpty(issue.Status))
	fmt.Fprintf(w, "Priority:\t%s\n", nameOrEmpty(issue.Priority))
	fmt.Fprintf(w, "Author:\t%s\n", nameOrEmpty(issue.Author))
	fmt.Fprintf(w, "Assignee:\t%s\n", nameOrEmpty(issue.AssignedTo))
	fmt.Fprintf(w, "Start date:\t%s\n", issue.StartDate)
	fmt.Fprintf(w, "Due date:\t%s\n", issue.DueDate)
	fmt.Fprintf(w, "Done:\t%d%%\n", issue.DoneRatio)
	fmt.Fprintf(w, "Created:\t%s\n", issue.CreatedOn)
	fmt.Fprintf(w, "Updated:\t%s\n", issue.UpdatedOn)
	if issue.Description != "" {
		fmt.Fprintf(w, "\nDescription:\n%s\n", issue.Description)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "output error:", err)
		return 1
	}
	return 0
}

func outputPBIs(pbis []api.PBI) int {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tStatus\tEstimate\tAuthor\tUpdated")
	for _, pbi := range pbis {
		author := nameOrEmpty(pbi.Author)
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", pbi.ID, pbi.Name, pbi.Status, pbi.Estimate, author, pbi.UpdatedAt)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(os.Stderr, "output error:", err)
		return 1
	}
	return 0
}

func outputPBIDetail(pbi api.PBI, issues []api.Issue) int {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%d\n", pbi.ID)
	fmt.Fprintf(w, "Name:\t%s\n", pbi.Name)
	fmt.Fprintf(w, "Status:\t%s\n", pbi.Status)
	fmt.Fprintf(w, "Estimate:\t%s\n", pbi.Estimate)
	fmt.Fprintf(w, "Board:\t%s\n", nameOrEmpty(pbi.Board))
	fmt.Fprintf(w, "Author:\t%s\n", nameOrEmpty(pbi.Author))
	fmt.Fprintf(w, "Created:\t%s\n", pbi.CreatedAt)
	fmt.Fprintf(w, "Updated:\t%s\n", pbi.UpdatedAt)
	if len(issues) > 0 {
		fmt.Fprintf(w, "\nIssues:\n")
		fmt.Fprintf(w, "  ID\tSubject\tStatus\tAssignee\tDone\n")
		for _, issue := range issues {
			status := nameOrEmpty(issue.Status)
			assignee := nameOrEmpty(issue.AssignedTo)
			fmt.Fprintf(w, "  #%d\t%s\t%s\t%s\t%d%%\n", issue.ID, issue.Subject, status, assignee, issue.DoneRatio)
		}
	} else if len(pbi.Issues) > 0 {
		// Fallback: show basic info if full details were not fetched
		fmt.Fprintf(w, "\nIssues:\n")
		for _, issue := range pbi.Issues {
			fmt.Fprintf(w, "  #%d\t%s\n", issue.ID, issue.Subject)
		}
	}
	if len(pbi.StickyNotes) > 0 {
		fmt.Fprintf(w, "\nSticky notes:\n")
		for _, note := range pbi.StickyNotes {
			fmt.Fprintf(w, "  %s\t[%s]\n", note.Name, note.Status)
		}
	}
	if pbi.Description != "" {
		fmt.Fprintf(w, "\nDescription:\n%s\n", pbi.Description)
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
