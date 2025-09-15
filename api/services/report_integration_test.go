package services

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ZiplEix/scrabble/api/database"
	"github.com/ZiplEix/scrabble/api/models/request"
)

func resetReports(t *testing.T) {
	t.Helper()
	_, err := database.Exec("DELETE FROM game_message_reads")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM messages")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM game_moves")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM game_players")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM games")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM push_subscriptions")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM reports")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM users")
	require.NoError(t, err)
}

func TestCreateAndGetReport(t *testing.T) {
	resetReports(t)
	u, err := CreateUser("rep_user", "pwd")
	require.NoError(t, err)

	id, err := CreateReport(u.ID, "Titre", "Contenu", "suggestion")
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	r, err := GetReportByID(strconv.FormatInt(id, 10))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, id, r.ID)
	assert.Equal(t, "Titre", r.Title)
	assert.Equal(t, "Contenu", r.Content)
	assert.Equal(t, "open", r.Status)
	assert.Equal(t, "suggestion", r.Type)
	assert.NotEmpty(t, r.CreatedAt)
	assert.NotEmpty(t, r.UpdatedAt)
	assert.Equal(t, "rep_user", r.Username)
}

func TestGetAllReports(t *testing.T) {
	resetReports(t)
	u, err := CreateUser("rep_all", "pwd")
	require.NoError(t, err)

	// create multiple reports
	_, _ = CreateReport(u.ID, "R1", "C1", "bug")
	_, _ = CreateReport(u.ID, "R2", "C2", "bug")
	_, _ = CreateReport(u.ID, "R3", "C3", "bug")

	all, err := GetAllReports()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(all), 3)
	// spot check fields
	assert.NotEmpty(t, all[0].Title)
	assert.NotEmpty(t, all[0].Username)
}

func TestUpdateReportStatus(t *testing.T) {
	resetReports(t)
	u, err := CreateUser("rep_status", "pwd")
	require.NoError(t, err)
	id, err := CreateReport(u.ID, "Status", "Body", "bug")
	require.NoError(t, err)

	err = UpdateReportStatus(strconv.FormatInt(id, 10), "in_progress")
	require.NoError(t, err)

	r, err := GetReportByID(strconv.FormatInt(id, 10))
	require.NoError(t, err)
	assert.Equal(t, "in_progress", r.Status)

	// non-existent
	err = UpdateReportStatus("999999", "resolved")
	require.Error(t, err)
}

func TestUpdateReport_Fields(t *testing.T) {
	resetReports(t)
	u, err := CreateUser("rep_update", "pwd")
	require.NoError(t, err)
	id, err := CreateReport(u.ID, "OldT", "OldC", "bug")
	require.NoError(t, err)

	err = UpdateReport(strconv.FormatInt(id, 10), request.UpdateReportRequest{
		Title:   "NewT",
		Content: "NewC",
		Status:  "resolved",
	})
	require.NoError(t, err)

	r, err := GetReportByID(strconv.FormatInt(id, 10))
	require.NoError(t, err)
	assert.Equal(t, "NewT", r.Title)
	assert.Equal(t, "NewC", r.Content)
	assert.Equal(t, "resolved", r.Status)

	// No-op update
	err = UpdateReport(strconv.FormatInt(id, 10), request.UpdateReportRequest{})
	require.Error(t, err)
}

func TestDeleteReport(t *testing.T) {
	resetReports(t)
	u, err := CreateUser("rep_delete", "pwd")
	require.NoError(t, err)
	id, err := CreateReport(u.ID, "T", "C", "bug")
	require.NoError(t, err)

	err = DeleteReport(strconv.FormatInt(id, 10))
	require.NoError(t, err)

	// verify it is gone
	_, err = GetReportByID(strconv.FormatInt(id, 10))
	require.Error(t, err)

	// delete missing
	err = DeleteReport("999999")
	require.Error(t, err)
}

func TestGetReportsByUserID(t *testing.T) {
	resetReports(t)
	u1, err := CreateUser("rep_u1", "pwd")
	require.NoError(t, err)
	u2, err := CreateUser("rep_u2", "pwd")
	require.NoError(t, err)

	// reports for u1
	_, _ = CreateReport(u1.ID, "u1 R1", "C", "bug")
	_, _ = CreateReport(u1.ID, "u1 R2", "C", "bug")
	// report for u2
	_, _ = CreateReport(u2.ID, "u2 R1", "C", "bug")

	list1, err := GetReportsByUserID(u1.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list1), 2)
	for _, m := range list1 {
		assert.Contains(t, m["title"], "u1")
		// username is a *string in the map, check presence and value
		if uname, ok := m["username"].(*string); ok && uname != nil {
			assert.Equal(t, "rep_u1", *uname)
		}
	}

	list2, err := GetReportsByUserID(u2.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list2), 1)
	for _, m := range list2 {
		assert.Contains(t, m["title"], "u2")
		if uname, ok := m["username"].(*string); ok && uname != nil {
			assert.Equal(t, "rep_u2", *uname)
		}
	}
}
