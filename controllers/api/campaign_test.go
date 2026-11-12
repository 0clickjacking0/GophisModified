package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gophishctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
)

func TestCampaignsBatchDelete(t *testing.T) {
	testCtx := setupTest(t)
	createTestData(t)

	campaigns, err := models.GetCampaigns(testCtx.admin.Id)
	if err != nil {
		t.Fatalf("error getting campaigns: %v", err)
	}
	if len(campaigns) == 0 {
		t.Fatal("expected at least one campaign")
	}

	ids := make([]int64, 0, len(campaigns))
	for _, campaign := range campaigns {
		ids = append(ids, campaign.Id)
	}
	body, err := json.Marshal(campaignDeleteRequest{IDs: ids})
	if err != nil {
		t.Fatalf("error encoding request: %v", err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/campaigns/batch-delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = gophishctx.Set(req, "user_id", testCtx.admin.Id)
	response := httptest.NewRecorder()

	testCtx.apiServer.CampaignsBatchDelete(response, req)

	if response.Code != http.StatusOK {
		t.Fatalf("incorrect status code received. expected %d got %d", http.StatusOK, response.Code)
	}
	got := models.Response{}
	err = json.NewDecoder(response.Body).Decode(&got)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if !got.Success {
		t.Fatalf("expected successful response: %+v", got)
	}

	remaining, err := models.GetCampaigns(testCtx.admin.Id)
	if err != nil {
		t.Fatalf("error getting remaining campaigns: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected no campaigns to remain, got %d", len(remaining))
	}
}
