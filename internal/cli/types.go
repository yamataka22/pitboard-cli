package cli

// API のレスポンスのうち、コマンドが中身を見る部分だけ型にする
type spaceSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type meData struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Token struct {
		Name  string `json:"name"`
		Scope string `json:"scope"`
	} `json:"token"`
	Spaces []spaceSummary `json:"spaces"`
}

type spaceDetail struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	DefaultKind string `json:"default_kind"`
	Me          struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"me"`
	Progresses []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Done bool   `json:"done"`
	} `json:"progresses"`
	Projects []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"projects"`
	Labels []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"labels"`
	Members []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		Assignee bool   `json:"assignee"`
		Me       bool   `json:"me"`
	} `json:"members"`
}

type ref struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type progressRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

type taskSummary struct {
	Number            int          `json:"number"`
	ID                int          `json:"id"`
	Name              string       `json:"name"`
	State             string       `json:"state"`
	Progress          *progressRef `json:"progress"`
	Kind              string       `json:"kind"`
	Point             *string      `json:"point"`
	PointDays         *float64     `json:"point_days"`
	Assignee          *ref         `json:"assignee"`
	Owner             *ref         `json:"owner"`
	Project           *ref         `json:"project"`
	Labels            []ref        `json:"labels"`
	Archived          bool         `json:"archived"`
	ProgressChangedAt *string      `json:"progress_changed_at"`
	TodoCount         int          `json:"todo_count"`
	TodoDoneCount     int          `json:"todo_done_count"`
	CreatedAt         string       `json:"created_at"`
	UpdatedAt         string       `json:"updated_at"`
	URL               string       `json:"url"`
}

type taskComment struct {
	ID        int           `json:"id"`
	Author    ref           `json:"author"`
	Content   string        `json:"content"`
	CreatedAt string        `json:"created_at"`
	Replies   []taskComment `json:"replies"`
}

type taskDetail struct {
	taskSummary
	Document *string `json:"document"`
	Todos    []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Done bool   `json:"done"`
	} `json:"todos"`
	Comments     []taskComment `json:"comments"`
	RelatedTasks []struct {
		Number int    `json:"number"`
		Name   string `json:"name"`
		State  string `json:"state"`
	} `json:"related_tasks"`
}

type pagination struct {
	Page    int  `json:"page"`
	PerPage int  `json:"per_page"`
	HasNext bool `json:"has_next"`
	Total   int  `json:"total"`
}
