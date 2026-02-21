package api

type NamedRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Issue struct {
	ID          int       `json:"id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description,omitempty"`
	DoneRatio   int       `json:"done_ratio,omitempty"`
	StartDate   string    `json:"start_date,omitempty"`
	DueDate     string    `json:"due_date,omitempty"`
	UpdatedOn   string    `json:"updated_on,omitempty"`
	CreatedOn   string    `json:"created_on,omitempty"`
	Project     *NamedRef `json:"project,omitempty"`
	Tracker     *NamedRef `json:"tracker,omitempty"`
	Status      *NamedRef `json:"status,omitempty"`
	Priority    *NamedRef `json:"priority,omitempty"`
	Author      *NamedRef `json:"author,omitempty"`
	AssignedTo  *NamedRef `json:"assigned_to,omitempty"`
}

type IssueInput struct {
	Subject      *string `json:"subject,omitempty"`
	ProjectID    *int    `json:"project_id,omitempty"`
	TrackerID    *int    `json:"tracker_id,omitempty"`
	StatusID     *int    `json:"status_id,omitempty"`
	PriorityID   *int    `json:"priority_id,omitempty"`
	AuthorID     *int    `json:"author_id,omitempty"`
	AssignedToID *int    `json:"assigned_to_id,omitempty"`
	Description  *string `json:"description,omitempty"`
	StartDate    *string `json:"start_date,omitempty"`
	DueDate      *string `json:"due_date,omitempty"`
	DoneRatio    *int    `json:"done_ratio,omitempty"`
	Notes        *string `json:"notes,omitempty"`
}

type IssueRequest struct {
	Issue IssueInput `json:"issue"`
}

type IssueResponse struct {
	Issue Issue `json:"issue"`
}

type IssueListResponse struct {
	Issues     []Issue `json:"issues"`
	TotalCount int     `json:"total_count"`
	Offset     int     `json:"offset"`
	Limit      int     `json:"limit"`
}

type SearchResult struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Datetime    string `json:"datetime"`
}

type SearchResponse struct {
	Results    []SearchResult `json:"results"`
	TotalCount int            `json:"total_count"`
	Offset     int            `json:"offset"`
	Limit      int            `json:"limit"`
}
