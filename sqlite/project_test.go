package sqlite_test

// import (
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"github.com/qualidafial/gtd-tui"
// )

// func TestDB_CreateProject(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		input   gtd.Project
// 		wantErr bool
// 	}{
// 		{
// 			name: "minimal project",
// 			input: gtd.Project{
// 				Title:  "Launch website",
// 				Status: gtd.ProjectStatusActive,
// 			},
// 		},
// 		{
// 			name: "full project",
// 			input: gtd.Project{
// 				Title:       "Launch website",
// 				Outcome:     "Website is live and accepting users",
// 				Description: "Covers design, dev, and deployment",
// 				Status:      gtd.ProjectStatusDeferred,
// 				Due:         new(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)),
// 			},
// 		},
// 		{
// 			name:    "missing title",
// 			input:   gtd.Project{Status: gtd.ProjectStatusActive},
// 			wantErr: true,
// 		},
// 		{
// 			name:    "missing status",
// 			input:   gtd.Project{Title: "Launch website"},
// 			wantErr: true,
// 		},
// 		{
// 			name:    "invalid status",
// 			input:   gtd.Project{Title: "Launch website", Status: "bogus"},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := openTestDB(t)
// 			c := ctx(t)

// 			got, err := db.CreateProject(c, tt.input)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				return
// 			}
// 			require.NoError(t, err)

// 			assert.NotZero(t, got.ID)
// 			assert.False(t, got.CreatedAt.IsZero())
// 			assert.False(t, got.UpdatedAt.IsZero())
// 			assert.Equal(t, tt.input.Title, got.Title)
// 			assert.Equal(t, tt.input.Outcome, got.Outcome)
// 			assert.Equal(t, tt.input.Description, got.Description)
// 			assert.Equal(t, tt.input.Status, got.Status)

// 			fetched, err := db.Project(c, got.ID)
// 			require.NoError(t, err)
// 			assert.Equal(t, got, fetched)
// 		})
// 	}
// }

// func TestDB_UpdateProject(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		setup  gtd.Project
// 		update func(gtd.Project) gtd.Project
// 		check  func(*testing.T, gtd.Project)
// 	}{
// 		{
// 			name:  "update title",
// 			setup: gtd.Project{Title: "Old title", Status: gtd.ProjectStatusActive},
// 			update: func(p gtd.Project) gtd.Project {
// 				p.Title = "New title"
// 				return p
// 			},
// 			check: func(t *testing.T, p gtd.Project) {
// 				assert.Equal(t, "New title", p.Title)
// 			},
// 		},
// 		{
// 			name:  "update status",
// 			setup: gtd.Project{Title: "Project", Status: gtd.ProjectStatusActive},
// 			update: func(p gtd.Project) gtd.Project {
// 				p.Status = gtd.ProjectStatusDone
// 				return p
// 			},
// 			check: func(t *testing.T, p gtd.Project) {
// 				assert.Equal(t, gtd.ProjectStatusDone, p.Status)
// 			},
// 		},
// 		{
// 			name:  "set due date",
// 			setup: gtd.Project{Title: "Project", Status: gtd.ProjectStatusActive},
// 			update: func(p gtd.Project) gtd.Project {
// 				p.Due = new(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
// 				return p
// 			},
// 			check: func(t *testing.T, p gtd.Project) {
// 				require.NotNil(t, p.Due)
// 				assert.True(t, p.Due.Equal(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)))
// 			},
// 		},
// 		{
// 			name:  "clear due date",
// 			setup: gtd.Project{Title: "Project", Status: gtd.ProjectStatusActive, Due: new(time.Now())},
// 			update: func(p gtd.Project) gtd.Project {
// 				p.Due = nil
// 				return p
// 			},
// 			check: func(t *testing.T, p gtd.Project) {
// 				assert.Nil(t, p.Due)
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := openTestDB(t)
// 			c := ctx(t)

// 			created, err := db.CreateProject(c, tt.setup)
// 			require.NoError(t, err)

// 			_, err = db.UpdateProject(c, tt.update(created))
// 			require.NoError(t, err)

// 			fetched, err := db.Project(c, created.ID)
// 			require.NoError(t, err)
// 			tt.check(t, fetched)
// 		})
// 	}
// }

// func TestDB_DeleteProject(t *testing.T) {
// 	db := openTestDB(t)
// 	c := ctx(t)

// 	created, err := db.CreateProject(c, gtd.Project{Title: "To delete", Status: gtd.ProjectStatusActive})
// 	require.NoError(t, err)

// 	require.NoError(t, db.DeleteProject(c, created.ID))

// 	_, err = db.Project(c, created.ID)
// 	assert.Error(t, err)
// }

// func TestDB_Projects(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		seed   []gtd.Project
// 		filter gtd.ProjectFilter
// 		want   []string
// 	}{
// 		{
// 			name: "all projects",
// 			seed: []gtd.Project{
// 				{Title: "Alpha", Status: gtd.ProjectStatusActive},
// 				{Title: "Beta", Status: gtd.ProjectStatusActive},
// 			},
// 			filter: gtd.ProjectFilter{},
// 			want:   []string{"Alpha", "Beta"},
// 		},
// 		{
// 			name: "filter by status",
// 			seed: []gtd.Project{
// 				{Title: "Active", Status: gtd.ProjectStatusActive},
// 				{Title: "Deferred", Status: gtd.ProjectStatusDeferred},
// 			},
// 			filter: gtd.ProjectFilter{Status: new(gtd.ProjectStatusDeferred)},
// 			want:   []string{"Deferred"},
// 		},
// 		{
// 			name: "filter by query matches title",
// 			seed: []gtd.Project{
// 				{Title: "Launch website", Status: gtd.ProjectStatusActive},
// 				{Title: "Hire engineer", Status: gtd.ProjectStatusActive},
// 			},
// 			filter: gtd.ProjectFilter{Query: "website"},
// 			want:   []string{"Launch website"},
// 		},
// 		{
// 			name: "filter by query matches outcome",
// 			seed: []gtd.Project{
// 				{Title: "Project A", Outcome: "Website is live", Status: gtd.ProjectStatusActive},
// 				{Title: "Project B", Outcome: "Team is fully staffed", Status: gtd.ProjectStatusActive},
// 			},
// 			filter: gtd.ProjectFilter{Query: "live"},
// 			want:   []string{"Project A"},
// 		},
// 		{
// 			name:   "empty result",
// 			seed:   []gtd.Project{{Title: "Alpha", Status: gtd.ProjectStatusActive}},
// 			filter: gtd.ProjectFilter{Query: "zzz"},
// 			want:   nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := openTestDB(t)
// 			c := ctx(t)

// 			for _, p := range tt.seed {
// 				_, err := db.CreateProject(c, p)
// 				require.NoError(t, err)
// 			}

// 			got, err := db.Projects(c, tt.filter)
// 			require.NoError(t, err)

// 			var titles []string
// 			for _, p := range got {
// 				titles = append(titles, p.Title)
// 			}
// 			assert.Equal(t, tt.want, titles)
// 		})
// 	}
// }
