package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"easy8-cli/internal/api"
	"easy8-cli/internal/config"
)

// Version is set at build time via -ldflags "-X easy8-cli/internal/cli.Version=..."
var Version = "dev"

func Run(args []string) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		return 1
	}

	if len(args) == 0 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "issue":
		return runIssue(args[1:], cfg)
	case "pbi":
		return runPBI(args[1:], cfg)
	case "version", "--version", "-v":
		fmt.Println(Version)
		return 0
	case "help", "-h", "--help":
		printUsage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", args[0])
		printUsage()
		return 2
	}
}

func runIssue(args []string, cfg config.Config) int {
	if len(args) == 0 {
		printIssueUsage()
		return 2
	}

	client := api.NewClient(cfg)

	switch args[0] {
	case "create":
		return runIssueCreate(args[1:], cfg, client)
	case "show":
		return runIssueShow(args[1:], cfg, client)
	case "list":
		return runIssueList(args[1:], cfg, client)
	case "search":
		return runIssueSearch(args[1:], cfg, client)
	case "update":
		return runIssueUpdate(args[1:], cfg, client)
	case "help", "-h", "--help":
		printIssueUsage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown issue command:", args[0])
		printIssueUsage()
		return 2
	}
}

func runIssueCreate(args []string, cfg config.Config, client *api.Client) int {
	fs := flag.NewFlagSet("issue create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	subject := fs.String("subject", "", "Issue subject (required)")
	description := fs.String("description", "", "Issue description")
	projectID := fs.Int("project-id", cfg.Defaults.ProjectID, "Project ID")
	trackerID := fs.Int("tracker-id", cfg.Defaults.TrackerID, "Tracker ID")
	statusID := fs.Int("status-id", cfg.Defaults.StatusID, "Status ID")
	priorityID := fs.Int("priority-id", cfg.Defaults.PriorityID, "Priority ID")
	authorID := fs.Int("author-id", cfg.Defaults.AuthorID, "Author ID")
	assignedToID := fs.Int("assigned-to-id", cfg.Defaults.AssignedToID, "Assigned to user ID")
	startDate := fs.String("start-date", "", "Start date (YYYY-MM-DD)")
	dueDate := fs.String("due-date", "", "Due date (YYYY-MM-DD)")
	var doneRatio optionalInt
	fs.Var(&doneRatio, "done-ratio", "Done ratio (0-100)")
	jsonOut := fs.Bool("json", false, "JSON output")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if err := requireString("subject", *subject); err != nil {
		return usageError(err)
	}
	if err := requireInt("project-id", *projectID); err != nil {
		return usageError(err)
	}
	if err := requireInt("tracker-id", *trackerID); err != nil {
		return usageError(err)
	}
	if err := requireInt("status-id", *statusID); err != nil {
		return usageError(err)
	}
	if err := requireInt("priority-id", *priorityID); err != nil {
		return usageError(err)
	}
	if err := requireInt("author-id", *authorID); err != nil {
		return usageError(err)
	}
	if err := requireInt("assigned-to-id", *assignedToID); err != nil {
		return usageError(err)
	}

	input := api.IssueInput{
		Subject:      stringPtr(*subject),
		ProjectID:    intPtr(*projectID),
		TrackerID:    intPtr(*trackerID),
		StatusID:     intPtr(*statusID),
		PriorityID:   intPtr(*priorityID),
		AuthorID:     intPtr(*authorID),
		AssignedToID: intPtr(*assignedToID),
	}
	if strings.TrimSpace(*description) != "" {
		input.Description = stringPtr(*description)
	}
	if strings.TrimSpace(*startDate) != "" {
		input.StartDate = stringPtr(*startDate)
	}
	if strings.TrimSpace(*dueDate) != "" {
		input.DueDate = stringPtr(*dueDate)
	}
	if doneRatio.set {
		if doneRatio.value < 0 || doneRatio.value > 100 {
			return usageError(fmt.Errorf("--done-ratio must be between 0 and 100"))
		}
		input.DoneRatio = intPtr(doneRatio.value)
	}

	resp, err := client.CreateIssue(context.Background(), input)
	if err != nil {
		return apiError(err)
	}

	if *jsonOut {
		return outputJSON(resp)
	}
	return outputIssues([]api.Issue{resp.Issue})
}

func runIssueShow(args []string, cfg config.Config, client *api.Client) int {
	normalizedArgs, positionalID, hasPositionalID, err := normalizeIDArgs(args)
	if err != nil {
		return usageError(err)
	}
	explicitID, hasExplicitID, err := extractExplicitIDArg(args)
	if err != nil {
		return usageError(err)
	}

	fs := flag.NewFlagSet("issue show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	id := fs.Int("id", 0, "Issue ID (required)")
	include := fs.String("include", "", "Include fields (comma-separated, e.g. journals,attachments)")
	jsonOut := fs.Bool("json", false, "JSON output")

	if err := fs.Parse(normalizedArgs); err != nil {
		return 2
	}

	if hasPositionalID && hasExplicitID && positionalID != explicitID {
		return usageError(fmt.Errorf("positional id %d does not match --id %d", positionalID, explicitID))
	}

	if *id == 0 {
		return usageError(fmt.Errorf("id is required (use '<id>' or --id)"))
	}

	var includes []string
	if strings.TrimSpace(*include) != "" {
		includes = splitComma(*include)
	}

	resp, err := client.GetIssue(context.Background(), *id, includes)
	if err != nil {
		return apiError(err)
	}

	if *jsonOut {
		return outputJSON(resp)
	}
	return outputIssueDetail(resp.Issue)
}

func runIssueList(args []string, cfg config.Config, client *api.Client) int {
	fs := flag.NewFlagSet("issue list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	limit := fs.Int("limit", 25, "Limit (max 100)")
	offset := fs.Int("offset", 0, "Offset")
	sort := fs.String("sort", "", "Sort expression")
	query := fs.String("q", "", "Free-text query (easy_query_q)")
	include := fs.String("include", "", "Include fields (comma-separated)")
	jsonOut := fs.Bool("json", false, "JSON output")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	params := api.IssueListParams{
		Limit:  *limit,
		Offset: *offset,
		Sort:   strings.TrimSpace(*sort),
		Query:  strings.TrimSpace(*query),
	}
	if strings.TrimSpace(*include) != "" {
		params.Include = splitComma(*include)
	}

	resp, err := client.ListIssues(context.Background(), params)
	if err != nil {
		return apiError(err)
	}
	if *jsonOut {
		return outputJSON(resp)
	}
	return outputIssues(resp.Issues)
}

func runIssueSearch(args []string, cfg config.Config, client *api.Client) int {
	fs := flag.NewFlagSet("issue search", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	query := fs.String("q", "", "Search query")
	limit := fs.Int("limit", 25, "Limit (max 100)")
	offset := fs.Int("offset", 0, "Offset")
	sort := fs.String("sort", "", "Sort expression")
	include := fs.String("include", "", "Include fields (comma-separated)")
	var assigneeID optionalInt
	var statusID optionalInt
	var priorityID optionalInt
	var taskTypeID optionalInt
	var projectID optionalInt
	fs.Var(&assigneeID, "assignee-id", "Assignee user ID")
	fs.Var(&statusID, "status-id", "Status ID")
	fs.Var(&priorityID, "priority-id", "Priority ID")
	fs.Var(&taskTypeID, "task-type-id", "Task type (tracker) ID")
	fs.Var(&projectID, "project-id", "Project ID")
	var dueDate string
	var subject string
	var assignee string
	var status string
	var priority string
	var taskType string
	var project string
	fs.StringVar(&dueDate, "due-date", "", "Due date (YYYY-MM-DD)")
	fs.StringVar(&subject, "subject", "", "Subject filter")
	fs.StringVar(&assignee, "assignee", "", "Assignee login or name")
	fs.StringVar(&status, "status", "", "Status name")
	fs.StringVar(&priority, "priority", "", "Priority name")
	fs.StringVar(&taskType, "task-type", "", "Task type (tracker) name")
	fs.StringVar(&project, "project", "", "Project name")
	jsonOut := fs.Bool("json", false, "JSON output")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	resolvedAssigneeID, err := resolveAssigneeID(context.Background(), client, assigneeID, assignee)
	if err != nil {
		return usageError(err)
	}
	resolvedStatusID, err := resolveStatusID(context.Background(), client, statusID, status)
	if err != nil {
		return usageError(err)
	}
	resolvedPriorityID, err := resolvePriorityID(context.Background(), client, priorityID, priority)
	if err != nil {
		return usageError(err)
	}
	resolvedTaskTypeID, err := resolveTaskTypeID(context.Background(), client, taskTypeID, taskType)
	if err != nil {
		return usageError(err)
	}
	resolvedProjectID, err := resolveProjectID(context.Background(), client, projectID, project)
	if err != nil {
		return usageError(err)
	}

	queryValue := strings.TrimSpace(*query)
	if queryValue == "" && resolvedAssigneeID == 0 && resolvedStatusID == 0 && resolvedPriorityID == 0 && resolvedTaskTypeID == 0 && resolvedProjectID == 0 && strings.TrimSpace(dueDate) == "" && strings.TrimSpace(subject) == "" {
		return usageError(fmt.Errorf("at least one filter is required (e.g. --q, --status, --assignee)"))
	}

	params := api.IssueListParams{
		Limit:      *limit,
		Offset:     *offset,
		Sort:       strings.TrimSpace(*sort),
		Query:      queryValue,
		DueDate:    strings.TrimSpace(dueDate),
		Subject:    strings.TrimSpace(subject),
		AssigneeID: resolvedAssigneeID,
		StatusID:   resolvedStatusID,
		PriorityID: resolvedPriorityID,
		TaskTypeID: resolvedTaskTypeID,
		ProjectID:  resolvedProjectID,
	}
	if strings.TrimSpace(*include) != "" {
		params.Include = splitComma(*include)
	}

	resp, err := client.ListIssues(context.Background(), params)
	if err != nil {
		return apiError(err)
	}
	if *jsonOut {
		return outputJSON(resp)
	}
	return outputIssues(resp.Issues)
}

func runIssueUpdate(args []string, cfg config.Config, client *api.Client) int {
	normalizedArgs, positionalID, hasPositionalID, err := normalizeIDArgs(args)
	if err != nil {
		return usageError(err)
	}
	explicitID, hasExplicitID, err := extractExplicitIDArg(args)
	if err != nil {
		return usageError(err)
	}

	fs := flag.NewFlagSet("issue update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	id := fs.Int("id", 0, "Issue ID (required)")
	subject := fs.String("subject", "", "Issue subject")
	description := fs.String("description", "", "Issue description")
	var statusID optionalInt
	var priorityID optionalInt
	var assignedToID optionalInt
	var doneRatio optionalInt
	fs.Var(&statusID, "status-id", "Status ID")
	fs.Var(&priorityID, "priority-id", "Priority ID")
	fs.Var(&assignedToID, "assigned-to-id", "Assigned to user ID")
	fs.Var(&doneRatio, "done-ratio", "Done ratio (0-100)")
	notes := fs.String("notes", "", "Notes (journal entry)")
	jsonOut := fs.Bool("json", false, "JSON output")

	if err := fs.Parse(normalizedArgs); err != nil {
		return 2
	}

	if hasPositionalID && hasExplicitID && positionalID != explicitID {
		return usageError(fmt.Errorf("positional id %d does not match --id %d", positionalID, explicitID))
	}

	if *id == 0 {
		return usageError(fmt.Errorf("id is required (use '<id>' or --id)"))
	}

	input := api.IssueInput{}
	if strings.TrimSpace(*subject) != "" {
		input.Subject = stringPtr(*subject)
	}
	if strings.TrimSpace(*description) != "" {
		input.Description = stringPtr(*description)
	}
	if statusID.set {
		input.StatusID = intPtr(statusID.value)
	}
	if priorityID.set {
		input.PriorityID = intPtr(priorityID.value)
	}
	if assignedToID.set {
		input.AssignedToID = intPtr(assignedToID.value)
	}
	if doneRatio.set {
		if doneRatio.value < 0 || doneRatio.value > 100 {
			return usageError(fmt.Errorf("--done-ratio must be between 0 and 100"))
		}
		input.DoneRatio = intPtr(doneRatio.value)
	}
	if strings.TrimSpace(*notes) != "" {
		input.Notes = stringPtr(*notes)
	}

	resp, err := client.UpdateIssue(context.Background(), *id, input)
	if err != nil {
		return apiError(err)
	}
	// Redmine PUT typically returns 200 with empty body.
	// Fetch the updated issue so the user sees the current state.
	if resp.Issue.ID == 0 {
		getResp, getErr := client.GetIssue(context.Background(), *id, nil)
		if getErr != nil {
			fmt.Fprintf(os.Stdout, "Issue #%d updated.\n", *id)
			return 0
		}
		resp = getResp
	}
	if *jsonOut {
		return outputJSON(resp)
	}
	return outputIssues([]api.Issue{resp.Issue})
}

// --- PBI commands ---

func runPBI(args []string, cfg config.Config) int {
	if len(args) == 0 {
		printPBIUsage()
		return 2
	}

	client := api.NewClient(cfg)

	switch args[0] {
	case "list":
		return runPBIList(args[1:], cfg, client)
	case "show":
		return runPBIShow(args[1:], cfg, client)
	case "update":
		return runPBIUpdate(args[1:], cfg, client)
	case "help", "-h", "--help":
		printPBIUsage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "unknown pbi command:", args[0])
		printPBIUsage()
		return 2
	}
}

func runPBIList(args []string, cfg config.Config, client *api.Client) int {
	fs := flag.NewFlagSet("pbi list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	limit := fs.Int("limit", 25, "Max items")
	offset := fs.Int("offset", 0, "Offset")
	sort := fs.String("sort", "", "Sort (e.g. updated_at:desc)")
	query := fs.String("q", "", "Fulltext search")
	status := fs.String("status", "", "Filter by status (to_do, realization, done, deleted)")
	authorID := fs.Int("author-id", 0, "Filter by author ID")
	boardID := fs.Int("board-id", 0, "Filter by board ID")
	jsonOut := fs.Bool("json", false, "JSON output")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	params := api.PBIListParams{
		Limit:    *limit,
		Offset:   *offset,
		Sort:     *sort,
		Query:    *query,
		Status:   *status,
		AuthorID: *authorID,
		BoardID:  *boardID,
	}

	resp, err := client.ListPBIs(context.Background(), params)
	if err != nil {
		return apiError(err)
	}

	if *jsonOut {
		return outputJSON(resp)
	}
	return outputPBIs(resp.PBIs)
}

func runPBIShow(args []string, cfg config.Config, client *api.Client) int {
	normalizedArgs, positionalID, hasPositionalID, err := normalizeIDArgs(args)
	if err != nil {
		return usageError(err)
	}
	explicitID, hasExplicitID, err := extractExplicitIDArg(args)
	if err != nil {
		return usageError(err)
	}

	fs := flag.NewFlagSet("pbi show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	id := fs.Int("id", 0, "PBI ID (required)")
	jsonOut := fs.Bool("json", false, "JSON output")

	if err := fs.Parse(normalizedArgs); err != nil {
		return 2
	}

	if hasPositionalID && hasExplicitID && positionalID != explicitID {
		return usageError(fmt.Errorf("positional id %d does not match --id %d", positionalID, explicitID))
	}

	if *id == 0 {
		return usageError(fmt.Errorf("id is required (use '<id>' or --id)"))
	}

	resp, err := client.GetPBI(context.Background(), *id)
	if err != nil {
		return apiError(err)
	}

	// Fetch full issue details if PBI has linked issues
	var issues []api.Issue
	if len(resp.PBI.Issues) > 0 {
		ids := make([]int, len(resp.PBI.Issues))
		for i, issue := range resp.PBI.Issues {
			ids[i] = issue.ID
		}
		issueResp, err := client.ListIssues(context.Background(), api.IssueListParams{
			IssueIDs: ids,
			Limit:    len(ids),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not fetch issue details: %v\n", err)
		} else {
			issues = issueResp.Issues
		}
	}

	if *jsonOut {
		return outputJSON(resp)
	}
	return outputPBIDetail(resp.PBI, issues)
}

func runPBIUpdate(args []string, cfg config.Config, client *api.Client) int {
	normalizedArgs, positionalID, hasPositionalID, err := normalizeIDArgs(args)
	if err != nil {
		return usageError(err)
	}
	explicitID, hasExplicitID, err := extractExplicitIDArg(args)
	if err != nil {
		return usageError(err)
	}

	fs := flag.NewFlagSet("pbi update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	id := fs.Int("id", 0, "PBI ID (required)")
	name := fs.String("name", "", "New name")
	description := fs.String("description", "", "New description")
	status := fs.String("status", "", "New status (to_do, realization, done, deleted)")
	estimate := fs.String("estimate", "", "New estimate")

	if err := fs.Parse(normalizedArgs); err != nil {
		return 2
	}

	if hasPositionalID && hasExplicitID && positionalID != explicitID {
		return usageError(fmt.Errorf("positional id %d does not match --id %d", positionalID, explicitID))
	}

	if *id == 0 {
		return usageError(fmt.Errorf("id is required (use '<id>' or --id)"))
	}

	var input api.PBIInput
	if strings.TrimSpace(*name) != "" {
		input.Name = name
	}
	if strings.TrimSpace(*description) != "" {
		input.Description = description
	}
	if strings.TrimSpace(*status) != "" {
		input.Status = status
	}
	if strings.TrimSpace(*estimate) != "" {
		input.Estimate = estimate
	}

	err = client.UpdatePBI(context.Background(), *id, input)
	if err != nil {
		return apiError(err)
	}

	fmt.Fprintf(os.Stdout, "PBI #%d updated.\n", *id)
	return 0
}

func usageError(err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	return 2
}

func requireString(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("--%s is required", name)
	}
	return nil
}

func requireInt(name string, value int) error {
	if value == 0 {
		return fmt.Errorf("--%s is required", name)
	}
	return nil
}

func printUsage() {
	lines := []string{
		"easy8-cli",
		"",
		"Usage:",
		"  easy8 issue <command> [flags]",
		"  easy8 pbi <command> [flags]",
		"  easy8 version",
		"",
		"Commands:",
		"  issue create   Create a new issue",
		"  issue show     Show issue detail",
		"  issue list     List issues",
		"  issue search   Fulltext search",
		"  issue update   Update an issue",
		"  pbi list       List product backlog items",
		"  pbi show       Show PBI detail",
		"  pbi update     Update a PBI",
		"  version        Print version",
		"",
		"Use 'easy8 issue --help' or 'easy8 pbi --help' for details.",
	}
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
}

func printIssueUsage() {
	lines := []string{
		"easy8 issue",
		"",
		"Usage:",
		"  easy8 issue create [flags]",
		"  easy8 issue show <id> [flags]",
		"  easy8 issue list [flags]",
		"  easy8 issue search [flags]",
		"  easy8 issue update <id> [flags]",
		"",
		"Examples:",
		"  easy8 issue show 123",
		"  easy8 issue show 123 --include journals,attachments --json",
		"  easy8 issue show --id 123",
		"  easy8 issue list --limit 10",
		"  easy8 issue search --q \"onboarding\"",
		"  easy8 issue search --q \"petr\" --assignee-id 51 --status-id 2 --priority-id 3",
		"  easy8 issue search --q \"petr\" --assignee \"Alice Doe\" --status \"New\" --priority \"High\" --task-type \"Task\" --project \"Project A\"",
		"  easy8 issue create --subject \"Fix login\" --project-id 1 --tracker-id 1 --status-id 1 --priority-id 1 --author-id 1 --assigned-to-id 2",
		"  easy8 issue update 123 --status-id 5",
		"  easy8 issue update --id 123 --status-id 5",
	}
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
}

func printPBIUsage() {
	lines := []string{
		"easy8 pbi",
		"",
		"Usage:",
		"  easy8 pbi list [flags]",
		"  easy8 pbi show <id> [flags]",
		"  easy8 pbi update <id> [flags]",
		"",
		"Examples:",
		"  easy8 pbi list --limit 10",
		"  easy8 pbi list --status to_do --board-id 17",
		"  easy8 pbi list --q \"design\" --author-id 51",
		"  easy8 pbi show 42",
		"  easy8 pbi show 42 --json",
		"  easy8 pbi show --id 42",
		"  easy8 pbi update 42 --status done",
		"  easy8 pbi update --id 42 --status done",
		"  easy8 pbi update --id 42 --name \"New name\" --estimate 5",
	}
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
}

type optionalInt struct {
	set   bool
	value int
}

func (flagValue *optionalInt) String() string {
	if !flagValue.set {
		return ""
	}
	return fmt.Sprintf("%d", flagValue.value)
}

func (flagValue *optionalInt) Set(value string) error {
	parsed, err := parseInt(value)
	if err != nil {
		return err
	}
	flagValue.value = parsed
	flagValue.set = true
	return nil
}

func parseInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid int: %s", value)
	}
	return parsed, nil
}

func normalizeIDArgs(args []string) ([]string, int, bool, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return args, 0, false, nil
	}

	id, err := parseInt(args[0])
	if err != nil {
		return nil, 0, false, fmt.Errorf("invalid id: %s", args[0])
	}

	normalized := make([]string, 0, len(args)+1)
	normalized = append(normalized, "--id", args[0])
	normalized = append(normalized, args[1:]...)
	return normalized, id, true, nil
}

func extractExplicitIDArg(args []string) (int, bool, error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--id" || arg == "-id":
			if i+1 >= len(args) {
				return 0, false, fmt.Errorf("--id requires a value")
			}
			id, err := parseInt(args[i+1])
			if err != nil {
				return 0, false, fmt.Errorf("invalid id: %s", args[i+1])
			}
			return id, true, nil
		case strings.HasPrefix(arg, "--id="):
			value := strings.TrimPrefix(arg, "--id=")
			id, err := parseInt(value)
			if err != nil {
				return 0, false, fmt.Errorf("invalid id: %s", value)
			}
			return id, true, nil
		case strings.HasPrefix(arg, "-id="):
			value := strings.TrimPrefix(arg, "-id=")
			id, err := parseInt(value)
			if err != nil {
				return 0, false, fmt.Errorf("invalid id: %s", value)
			}
			return id, true, nil
		}
	}

	return 0, false, nil
}

func splitComma(input string) []string {
	parts := strings.Split(input, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func apiError(err error) int {
	var apiErr api.APIError
	if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "api error %d: %s\n", apiErr.StatusCode, apiErr.Body)
		return 1
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	return 1
}

type nameID struct {
	ID   int
	Name string
}

func resolveAssigneeID(ctx context.Context, client *api.Client, id optionalInt, name string) (int, error) {
	if strings.TrimSpace(name) == "" {
		if id.set {
			return id.value, nil
		}
		return 0, nil
	}

	users, err := client.ListUsers(ctx)
	if err != nil {
		return 0, err
	}

	needle := normalizeName(name)
	var matches []api.User
	for _, user := range users {
		if matchesUser(user, needle) {
			matches = append(matches, user)
		}
	}

	if len(matches) == 0 {
		return 0, fmt.Errorf("assignee not found: %s", name)
	}
	if len(matches) > 1 {
		return 0, fmt.Errorf("assignee matches multiple users: %s", name)
	}
	match := matches[0]
	if id.set && id.value != match.ID {
		return 0, fmt.Errorf("assignee-id does not match assignee name")
	}
	return match.ID, nil
}

func resolveStatusID(ctx context.Context, client *api.Client, id optionalInt, name string) (int, error) {
	if strings.TrimSpace(name) == "" {
		if id.set {
			return id.value, nil
		}
		return 0, nil
	}
	items, err := client.ListIssueStatuses(ctx)
	if err != nil {
		return 0, err
	}
	return resolveNameID(id, name, toNameIDsStatus(items), "status")
}

func resolvePriorityID(ctx context.Context, client *api.Client, id optionalInt, name string) (int, error) {
	if strings.TrimSpace(name) == "" {
		if id.set {
			return id.value, nil
		}
		return 0, nil
	}
	items, err := client.ListIssuePriorities(ctx)
	if err != nil {
		return 0, err
	}
	return resolveNameID(id, name, toNameIDsPriority(items), "priority")
}

func resolveTaskTypeID(ctx context.Context, client *api.Client, id optionalInt, name string) (int, error) {
	if strings.TrimSpace(name) == "" {
		if id.set {
			return id.value, nil
		}
		return 0, nil
	}
	items, err := client.ListTrackers(ctx)
	if err != nil {
		return 0, err
	}
	return resolveNameID(id, name, toNameIDsTracker(items), "task-type")
}

func resolveProjectID(ctx context.Context, client *api.Client, id optionalInt, name string) (int, error) {
	if strings.TrimSpace(name) == "" {
		if id.set {
			return id.value, nil
		}
		return 0, nil
	}
	items, err := client.ListProjects(ctx)
	if err != nil {
		return 0, err
	}
	return resolveNameID(id, name, toNameIDsProject(items), "project")
}

func resolveNameID(id optionalInt, name string, items []nameID, label string) (int, error) {
	needle := normalizeName(name)
	var matches []nameID
	for _, item := range items {
		if normalizeName(item.Name) == needle {
			matches = append(matches, item)
		}
	}
	if len(matches) == 0 {
		return 0, fmt.Errorf("%s not found: %s", label, name)
	}
	if len(matches) > 1 {
		return 0, fmt.Errorf("%s matches multiple entries: %s", label, name)
	}
	match := matches[0]
	if id.set && id.value != match.ID {
		return 0, fmt.Errorf("%s-id does not match %s name", label, label)
	}
	return match.ID, nil
}

func normalizeName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func matchesUser(user api.User, needle string) bool {
	if normalizeName(user.Login) == needle {
		return true
	}
	full := strings.TrimSpace(user.Firstname + " " + user.Lastname)
	if normalizeName(full) == needle {
		return true
	}
	return false
}

func toNameIDsStatus(items []api.IssueStatus) []nameID {
	result := make([]nameID, 0, len(items))
	for _, item := range items {
		result = append(result, nameID{ID: item.ID, Name: item.Name})
	}
	return result
}

func toNameIDsPriority(items []api.IssuePriority) []nameID {
	result := make([]nameID, 0, len(items))
	for _, item := range items {
		result = append(result, nameID{ID: item.ID, Name: item.Name})
	}
	return result
}

func toNameIDsTracker(items []api.Tracker) []nameID {
	result := make([]nameID, 0, len(items))
	for _, item := range items {
		result = append(result, nameID{ID: item.ID, Name: item.Name})
	}
	return result
}

func toNameIDsProject(items []api.Project) []nameID {
	result := make([]nameID, 0, len(items))
	for _, item := range items {
		result = append(result, nameID{ID: item.ID, Name: item.Name})
	}
	return result
}
