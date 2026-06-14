package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
)

func TestProjectService_ConvertTaskToProject(t *testing.T) {
	t.Run("promotes task to project as first action", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "Plan offsite", Description: "Q3 team", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		project, reframed, err := projSvc.ConvertTaskToProject(ctx, task.ID,
			gtd.Project{Outcome: "Team gathered offsite"},
			gtd.Task{Title: "Book the venue", Description: "near downtown"})
		require.NoError(t, err)

		// Project inherits title/description from the task; outcome is supplied.
		assert.Equal(t, "Plan offsite", project.Title)
		assert.Equal(t, "Q3 team", project.Description)
		assert.Equal(t, "Team gathered offsite", project.Outcome)
		assert.Equal(t, gtd.ProjectStatusOpen, project.Status)
		assert.NotZero(t, project.ID)

		// The original task is re-parented and reframed.
		assert.Equal(t, task.ID, reframed.ID)
		require.NotNil(t, reframed.ProjectID)
		assert.Equal(t, project.ID, *reframed.ProjectID)
		assert.Equal(t, "Book the venue", reframed.Title)
		assert.Equal(t, "near downtown", reframed.Description)
		assert.Equal(t, gtd.TaskStatusOpen, reframed.Status)

		// Persisted: the task belongs to the project.
		got, err := taskSvc.GetTask(ctx, task.ID)
		require.NoError(t, err)
		require.NotNil(t, got.ProjectID)
		assert.Equal(t, project.ID, *got.ProjectID)
		assert.Equal(t, "Book the venue", got.Title)
	})

	t.Run("project field overrides honored", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "task title", Description: "task desc", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		project, _, err := projSvc.ConvertTaskToProject(ctx, task.ID,
			gtd.Project{Title: "explicit project", Description: "explicit desc", Outcome: "done"},
			gtd.Task{Title: "first step"})
		require.NoError(t, err)
		assert.Equal(t, "explicit project", project.Title)
		assert.Equal(t, "explicit desc", project.Description)
	})

	t.Run("rejects task already in a project", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		host, err := projSvc.CreateProject(ctx, gtd.Project{Title: "host", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "t", Status: gtd.TaskStatusOpen, ProjectID: &host.ID})
		require.NoError(t, err)

		_, _, err = projSvc.ConvertTaskToProject(ctx, task.ID, gtd.Project{Outcome: "o"}, gtd.Task{Title: "x"})
		require.Error(t, err)
	})

	t.Run("rejects missing task without creating a project", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		projSvc := service.NewProjectService(db)

		_, _, err := projSvc.ConvertTaskToProject(ctx, 999, gtd.Project{Outcome: "o"}, gtd.Task{Title: "x"})
		require.Error(t, err)

		// Atomicity: the failed conversion left no project behind.
		projects, err := projSvc.ListProjects(ctx, gtd.ProjectFilter{})
		require.NoError(t, err)
		assert.Empty(t, projects)
	})
}

func TestProjectService_ConvertProjectToTask(t *testing.T) {
	t.Run("collapses empty open project to standalone task", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		due := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		project, err := projSvc.CreateProject(ctx, gtd.Project{Title: "tiny", Description: "desc", Due: &due, Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)

		task, err := projSvc.ConvertProjectToTask(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, "tiny", task.Title)
		assert.Equal(t, "desc", task.Description)
		assert.Equal(t, gtd.TaskStatusOpen, task.Status)
		assert.Nil(t, task.ProjectID)
		require.NotNil(t, task.Due)
		assert.True(t, task.Due.Equal(due))

		// The project is gone.
		_, err = projSvc.GetProject(ctx, project.ID)
		require.Error(t, err)

		// The task is a standalone task in default views.
		tasks, err := taskSvc.ListTasks(ctx, gtd.TaskFilter{})
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, task.ID, tasks[0].ID)
	})

	t.Run("folds outcome into description", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		projSvc := service.NewProjectService(db)

		project, err := projSvc.CreateProject(ctx, gtd.Project{Title: "t", Description: "body", Outcome: "the goal", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)

		task, err := projSvc.ConvertProjectToTask(ctx, project.ID)
		require.NoError(t, err)
		assert.Equal(t, "body\n\nthe goal", task.Description)
	})

	t.Run("rejects non-open project", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		projSvc := service.NewProjectService(db)

		project, err := projSvc.CreateProject(ctx, gtd.Project{Title: "someday", Status: gtd.ProjectStatusSomeday})
		require.NoError(t, err)

		_, err = projSvc.ConvertProjectToTask(ctx, project.ID)
		require.Error(t, err)
	})

	t.Run("rejects project with tasks of any status", func(t *testing.T) {
		statuses := []gtd.TaskStatus{gtd.TaskStatusOpen, gtd.TaskStatusDone, gtd.TaskStatusDropped}
		for _, st := range statuses {
			db := openTestDB(t)
			ctx := t.Context()
			taskSvc := service.NewTaskService(db)
			projSvc := service.NewProjectService(db)

			project, err := projSvc.CreateProject(ctx, gtd.Project{Title: "p", Status: gtd.ProjectStatusOpen})
			require.NoError(t, err)
			created, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "t", Status: gtd.TaskStatusOpen, ProjectID: &project.ID})
			require.NoError(t, err)
			// Move the task into the target terminal status where needed.
			switch st {
			case gtd.TaskStatusDone:
				_, err = taskSvc.CompleteTask(ctx, created.ID, time.Now())
				require.NoError(t, err)
			case gtd.TaskStatusDropped:
				_, err = taskSvc.DropTask(ctx, created.ID, time.Now())
				require.NoError(t, err)
			}

			_, err = projSvc.ConvertProjectToTask(ctx, project.ID)
			require.Errorf(t, err, "status %s should block conversion", st)

			// Atomicity: the project still exists.
			_, err = projSvc.GetProject(ctx, project.ID)
			require.NoError(t, err)
		}
	})
}

func TestProjectService_LinkTaskToProject(t *testing.T) {
	t.Run("links standalone task and orders it last", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		project, err := projSvc.CreateProject(ctx, gtd.Project{Title: "p", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		existing, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "existing", Status: gtd.TaskStatusOpen, ProjectID: &project.ID})
		require.NoError(t, err)
		standalone, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "standalone", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		linked, err := projSvc.LinkTaskToProject(ctx, standalone.ID, project.ID)
		require.NoError(t, err)
		require.NotNil(t, linked.ProjectID)
		assert.Equal(t, project.ID, *linked.ProjectID)

		// Ordered last within the project's tasks.
		tasks, err := taskSvc.ListTasks(ctx, gtd.TaskFilter{}.WithProjectID(project.ID))
		require.NoError(t, err)
		require.Len(t, tasks, 2)
		assert.Equal(t, existing.ID, tasks[0].ID)
		assert.Equal(t, standalone.ID, tasks[1].ID)
	})

	t.Run("rejects non-standalone task", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		p1, err := projSvc.CreateProject(ctx, gtd.Project{Title: "p1", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		p2, err := projSvc.CreateProject(ctx, gtd.Project{Title: "p2", Status: gtd.ProjectStatusOpen})
		require.NoError(t, err)
		task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "t", Status: gtd.TaskStatusOpen, ProjectID: &p1.ID})
		require.NoError(t, err)

		_, err = projSvc.LinkTaskToProject(ctx, task.ID, p2.ID)
		require.Error(t, err)
	})

	t.Run("rejects invalid project", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "t", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		_, err = projSvc.LinkTaskToProject(ctx, task.ID, 999)
		require.Error(t, err)

		// The task remains standalone after the failed link.
		got, err := taskSvc.GetTask(ctx, task.ID)
		require.NoError(t, err)
		assert.Nil(t, got.ProjectID)
	})

	t.Run("linking into a someday project hides the task until reopened", func(t *testing.T) {
		db := openTestDB(t)
		ctx := t.Context()
		taskSvc := service.NewTaskService(db)
		projSvc := service.NewProjectService(db)

		project, err := projSvc.CreateProject(ctx, gtd.Project{Title: "later", Status: gtd.ProjectStatusSomeday})
		require.NoError(t, err)
		task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "t", Status: gtd.TaskStatusOpen})
		require.NoError(t, err)

		linked, err := projSvc.LinkTaskToProject(ctx, task.ID, project.ID)
		require.NoError(t, err)
		assert.Equal(t, gtd.TaskStatusOpen, linked.Status, "status is unchanged by linking")

		// Excluded from default views (IncludeSomedayProjects=false).
		def, err := taskSvc.ListTasks(ctx, gtd.TaskFilter{})
		require.NoError(t, err)
		assert.Empty(t, def)

		// Included when the someday filter is on.
		incl, err := taskSvc.ListTasks(ctx, gtd.TaskFilter{IncludeSomedayProjects: true})
		require.NoError(t, err)
		require.Len(t, incl, 1)

		// Reopening the project restores the task to default views.
		_, err = projSvc.ReopenProject(ctx, project.ID, time.Now())
		require.NoError(t, err)
		restored, err := taskSvc.ListTasks(ctx, gtd.TaskFilter{})
		require.NoError(t, err)
		require.Len(t, restored, 1)
		assert.Equal(t, task.ID, restored[0].ID)
	})
}
