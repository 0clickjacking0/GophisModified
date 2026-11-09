package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
)

func getFirstCampaign(t *testing.T) models.Campaign {
	campaigns, err := models.GetCampaigns(1)
	if err != nil {
		t.Fatalf("error getting first campaign from database: %v", err)
	}
	return campaigns[0]
}

func getFirstEmailRequest(t *testing.T) models.EmailRequest {
	campaign := getFirstCampaign(t)
	req := models.EmailRequest{
		TemplateId:    campaign.TemplateId,
		Template:      campaign.Template,
		PageId:        campaign.PageId,
		Page:          campaign.Page,
		URL:           "http://localhost.localdomain",
		UserId:        1,
		BaseRecipient: campaign.Results[0].BaseRecipient,
		SMTP:          campaign.SMTP,
		FromAddress:   campaign.SMTP.FromAddress,
	}
	err := models.PostEmailRequest(&req)
	if err != nil {
		t.Fatalf("error creating email request: %v", err)
	}
	return req
}

func openEmail(t *testing.T, ctx *testContext, rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/track?%s=%s", ctx.phishServer.URL, models.RecipientParameter, rid))
	if err != nil {
		t.Fatalf("error requesting /track endpoint: %v", err)
	}
	defer resp.Body.Close()
	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body from /track endpoint: %v", err)
	}
	expected, err := ioutil.ReadFile("static/images/pixel.png")
	if err != nil {
		t.Fatalf("error reading local transparent pixel: %v", err)
	}
	if !bytes.Equal(got, expected) {
		t.Fatalf("unexpected tracking pixel data received. expected %#v got %#v", expected, got)
	}
}

func openEmail404(t *testing.T, ctx *testContext, rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/track?%s=%s", ctx.phishServer.URL, models.RecipientParameter, rid))
	if err != nil {
		t.Fatalf("error requesting /track endpoint: %v", err)
	}
	defer resp.Body.Close()
	got := resp.StatusCode
	expected := http.StatusNotFound
	if got != expected {
		t.Fatalf("invalid status code received for /track endpoint. expected %d got %d", expected, got)
	}
}

func reportedEmail(t *testing.T, ctx *testContext, rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/report?%s=%s", ctx.phishServer.URL, models.RecipientParameter, rid))
	if err != nil {
		t.Fatalf("error requesting /report endpoint: %v", err)
	}
	got := resp.StatusCode
	expected := http.StatusNoContent
	if got != expected {
		t.Fatalf("invalid status code received for /report endpoint. expected %d got %d", expected, got)
	}
}

func reportEmail404(t *testing.T, ctx *testContext, rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/report?%s=%s", ctx.phishServer.URL, models.RecipientParameter, rid))
	if err != nil {
		t.Fatalf("error requesting /report endpoint: %v", err)
	}
	got := resp.StatusCode
	expected := http.StatusNotFound
	if got != expected {
		t.Fatalf("invalid status code received for /report endpoint. expected %d got %d", expected, got)
	}
}

func clickLink(t *testing.T, ctx *testContext, rid string, expectedHTML string) {
	clickLinkWithUserAgent(t, ctx, rid, expectedHTML, "")
}

func clickLinkWithUserAgent(t *testing.T, ctx *testContext, rid string, expectedHTML string, userAgent string) {
	resp := getWithUserAgent(t, fmt.Sprintf("%s/?%s=%s", ctx.phishServer.URL, models.RecipientParameter, rid), userAgent)
	defer resp.Body.Close()
	got, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading payload from / endpoint response: %v", err)
	}
	if !bytes.Equal(got, []byte(expectedHTML)) {
		t.Fatalf("invalid response received from / endpoint. expected %s got %s", got, expectedHTML)
	}
}

func getWithUserAgent(t *testing.T, requestURL string, userAgent string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		t.Fatalf("error creating GET request: %v", err)
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("error sending GET request: %v", err)
	}
	return resp
}

func clickLink404(t *testing.T, ctx *testContext, rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/?%s=%s", ctx.phishServer.URL, models.RecipientParameter, rid))
	if err != nil {
		t.Fatalf("error requesting / endpoint: %v", err)
	}
	defer resp.Body.Close()
	got := resp.StatusCode
	expected := http.StatusNotFound
	if got != expected {
		t.Fatalf("invalid status code received for / endpoint. expected %d got %d", expected, got)
	}
}

func transparencyRequest(t *testing.T, ctx *testContext, r models.Result, rid, path string) {
	resp, err := http.Get(fmt.Sprintf("%s%s?%s=%s", ctx.phishServer.URL, path, models.RecipientParameter, rid))
	if err != nil {
		t.Fatalf("error requesting %s endpoint: %v", path, err)
	}
	defer resp.Body.Close()
	got := resp.StatusCode
	expected := http.StatusOK
	if got != expected {
		t.Fatalf("invalid status code received for / endpoint. expected %d got %d", expected, got)
	}
	tr := &TransparencyResponse{}
	err = json.NewDecoder(resp.Body).Decode(tr)
	if err != nil {
		t.Fatalf("error unmarshaling transparency request: %v", err)
	}
	expectedResponse := &TransparencyResponse{
		ContactAddress: ctx.config.ContactAddress,
		SendDate:       r.SendDate,
		Server:         config.ServerName,
	}
	if !reflect.DeepEqual(tr, expectedResponse) {
		t.Fatalf("unexpected transparency response received. expected %v got %v", expectedResponse, tr)
	}
}

func TestOpenedPhishingEmail(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	campaign := getFirstCampaign(t)
	result := campaign.Results[0]
	if result.Status != models.StatusSending {
		t.Fatalf("unexpected result status received. expected %s got %s", models.StatusSending, result.Status)
	}

	openEmail(t, ctx, result.RId)

	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	if result.Status != models.EventOpened {
		t.Fatalf("unexpected result status received. expected %s got %s", models.EventOpened, result.Status)
	}
	if lastEvent.Message != models.EventOpened {
		t.Fatalf("unexpected event status received. expected %s got %s", lastEvent.Message, models.EventOpened)
	}
	if result.ModifiedDate != lastEvent.Time {
		t.Fatalf("unexpected result modified date received. expected %s got %s", lastEvent.Time, result.ModifiedDate)
	}
}

func TestReportedPhishingEmail(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	campaign := getFirstCampaign(t)
	result := campaign.Results[0]
	if result.Status != models.StatusSending {
		t.Fatalf("unexpected result status received. expected %s got %s", models.StatusSending, result.Status)
	}

	reportedEmail(t, ctx, result.RId)

	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]

	if result.Reported != true {
		t.Fatalf("unexpected result report status received. expected %v got %v", true, result.Reported)
	}
	if lastEvent.Message != models.EventReported {
		t.Fatalf("unexpected event status received. expected %s got %s", lastEvent.Message, models.EventReported)
	}
	if result.ModifiedDate != lastEvent.Time {
		t.Fatalf("unexpected result modified date received. expected %s got %s", lastEvent.Time, result.ModifiedDate)
	}
}

func TestClickedPhishingLinkAfterOpen(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	campaign := getFirstCampaign(t)
	result := campaign.Results[0]
	if result.Status != models.StatusSending {
		t.Fatalf("unexpected result status received. expected %s got %s", models.StatusSending, result.Status)
	}

	openEmail(t, ctx, result.RId)
	clickLink(t, ctx, result.RId, campaign.Page.HTML)

	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	if result.Status != models.EventClicked {
		t.Fatalf("unexpected result status received. expected %s got %s", models.EventClicked, result.Status)
	}
	if lastEvent.Message != models.EventClicked {
		t.Fatalf("unexpected event status received. expected %s got %s", lastEvent.Message, models.EventClicked)
	}
	if result.ModifiedDate != lastEvent.Time {
		t.Fatalf("unexpected result modified date received. expected %s got %s", lastEvent.Time, result.ModifiedDate)
	}
}

func TestGatewayScannerLinkIsNotCounted(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	campaign := getFirstCampaign(t)
	result := campaign.Results[0]
	if err := result.HandleEmailSent(); err != nil {
		t.Fatalf("error marking email as sent: %v", err)
	}
	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	initialEventCount := len(campaign.Events)
	initialModifiedDate := result.ModifiedDate

	trackerURL := fmt.Sprintf("%s/resource/image/pixel.png?%s=%s", ctx.phishServer.URL, models.RecipientParameter, result.RId)
	trackerResponse := getWithUserAgent(t, trackerURL, gatewayScannerUserAgent)
	trackerResponse.Body.Close()

	clickLinkWithUserAgent(t, ctx, result.RId, campaign.Page.HTML, gatewayScannerUserAgent)

	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	if result.Status != models.EventSent {
		t.Fatalf("gateway scanner changed result status. expected %s got %s", models.EventSent, result.Status)
	}
	if result.IP != "" {
		t.Fatalf("gateway scanner changed result IP. expected empty IP got %s", result.IP)
	}
	if result.Latitude != 0 || result.Longitude != 0 {
		t.Fatalf("gateway scanner changed result location. got latitude %f longitude %f", result.Latitude, result.Longitude)
	}
	if !result.ModifiedDate.Equal(initialModifiedDate) {
		t.Fatalf("gateway scanner changed modified date. expected %s got %s", initialModifiedDate, result.ModifiedDate)
	}
	if len(campaign.Events) != initialEventCount {
		t.Fatalf("gateway scanner created an event. expected %d events got %d", initialEventCount, len(campaign.Events))
	}

	humanUserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36 Edg/151.0.0.0"
	clickLinkWithUserAgent(t, ctx, result.RId, campaign.Page.HTML, humanUserAgent)
	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	if result.Status != models.EventClicked {
		t.Fatalf("human click after gateway scan was not counted. expected %s got %s", models.EventClicked, result.Status)
	}
	if len(campaign.Events) != initialEventCount+1 {
		t.Fatalf("unexpected event count after human click. expected %d got %d", initialEventCount+1, len(campaign.Events))
	}
}

func TestSimilarBrowserUserAgentIsCounted(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	campaign := getFirstCampaign(t)
	result := campaign.Results[0]
	if err := result.HandleEmailSent(); err != nil {
		t.Fatalf("error marking email as sent: %v", err)
	}
	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.1 Safari/537.36"

	clickLinkWithUserAgent(t, ctx, result.RId, campaign.Page.HTML, userAgent)

	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	if result.Status != models.EventClicked {
		t.Fatalf("similar browser user agent was not counted. expected %s got %s", models.EventClicked, result.Status)
	}
	if lastEvent.Message != models.EventClicked {
		t.Fatalf("unexpected event status. expected %s got %s", models.EventClicked, lastEvent.Message)
	}
}

func TestGatewayScannerRequestDetection(t *testing.T) {
	sentAt := time.Date(2026, time.August, 14, 1, 15, 33, 0, time.UTC)
	result := models.Result{Status: models.EventSent, SendDate: sentAt}
	request, err := http.NewRequest(http.MethodGet, "http://example.test/?id=test", nil)
	if err != nil {
		t.Fatalf("error creating gateway request: %v", err)
	}
	request.Header.Set("User-Agent", gatewayScannerUserAgent)

	tests := []struct {
		name       string
		method     string
		userAgent  string
		result     models.Result
		receivedAt time.Time
		expected   bool
	}{
		{name: "inside window", method: http.MethodGet, userAgent: gatewayScannerUserAgent, result: result, receivedAt: sentAt.Add(350 * time.Millisecond), expected: true},
		{name: "window boundary", method: http.MethodGet, userAgent: gatewayScannerUserAgent, result: result, receivedAt: sentAt.Add(gatewayScannerWindow), expected: true},
		{name: "after window", method: http.MethodGet, userAgent: gatewayScannerUserAgent, result: result, receivedAt: sentAt.Add(gatewayScannerWindow + time.Nanosecond), expected: false},
		{name: "before sent", method: http.MethodGet, userAgent: gatewayScannerUserAgent, result: result, receivedAt: sentAt.Add(-time.Nanosecond), expected: false},
		{name: "not sent", method: http.MethodGet, userAgent: gatewayScannerUserAgent, result: models.Result{Status: models.StatusSending, SendDate: sentAt}, receivedAt: sentAt, expected: false},
		{name: "zero send date", method: http.MethodGet, userAgent: gatewayScannerUserAgent, result: models.Result{Status: models.EventSent}, receivedAt: sentAt, expected: false},
		{name: "post request", method: http.MethodPost, userAgent: gatewayScannerUserAgent, result: result, receivedAt: sentAt, expected: false},
		{name: "different user agent", method: http.MethodGet, userAgent: "Mozilla/5.0 Chrome/151.0.0.0", result: result, receivedAt: sentAt, expected: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request.Method = test.method
			request.Header.Set("User-Agent", test.userAgent)
			if got := isGatewayScannerRequest(request, test.result, test.receivedAt); got != test.expected {
				t.Fatalf("unexpected gateway scanner detection. expected %t got %t", test.expected, got)
			}
		})
	}
}

func TestNoRecipientID(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	resp, err := http.Get(fmt.Sprintf("%s/track", ctx.phishServer.URL))
	if err != nil {
		t.Fatalf("error requesting /track endpoint: %v", err)
	}
	got := resp.StatusCode
	expected := http.StatusNotFound
	if got != expected {
		t.Fatalf("invalid status code received for /track endpoint. expected %d got %d", expected, got)
	}

	resp, err = http.Get(ctx.phishServer.URL)
	if err != nil {
		t.Fatalf("error requesting /track endpoint: %v", err)
	}
	got = resp.StatusCode
	if got != expected {
		t.Fatalf("invalid status code received for / endpoint. expected %d got %d", expected, got)
	}
}

func TestInvalidRecipientID(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	rid := "XXXXXXXXXX"
	openEmail404(t, ctx, rid)
	clickLink404(t, ctx, rid)
	reportEmail404(t, ctx, rid)
}

func TestCompletedCampaignClick(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	campaign := getFirstCampaign(t)
	result := campaign.Results[0]
	if result.Status != models.StatusSending {
		t.Fatalf("unexpected result status received. expected %s got %s", models.StatusSending, result.Status)
	}

	openEmail(t, ctx, result.RId)

	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	if result.Status != models.EventOpened {
		t.Fatalf("unexpected result status received. expected %s got %s", models.EventOpened, result.Status)
	}

	models.CompleteCampaign(campaign.Id, 1)
	openEmail404(t, ctx, result.RId)
	clickLink404(t, ctx, result.RId)

	campaign = getFirstCampaign(t)
	result = campaign.Results[0]
	if result.Status != models.EventOpened {
		t.Fatalf("unexpected result status received. expected %s got %s", models.EventOpened, result.Status)
	}
}

func TestRobotsHandler(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	resp, err := http.Get(fmt.Sprintf("%s/robots.txt", ctx.phishServer.URL))
	if err != nil {
		t.Fatalf("error requesting /robots.txt endpoint: %v", err)
	}
	defer resp.Body.Close()
	got := resp.StatusCode
	expectedStatus := http.StatusOK
	if got != expectedStatus {
		t.Fatalf("invalid status code received for /track endpoint. expected %d got %d", expectedStatus, got)
	}
	expected := []byte("User-agent: *\nDisallow: /\n")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body from /robots.txt endpoint: %v", err)
	}
	if !bytes.Equal(body, expected) {
		t.Fatalf("invalid robots.txt response received. expected %s got %s", expected, body)
	}
}

func TestInvalidPreviewID(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	bogusRId := fmt.Sprintf("%sbogus", models.PreviewPrefix)
	openEmail404(t, ctx, bogusRId)
	clickLink404(t, ctx, bogusRId)
	reportEmail404(t, ctx, bogusRId)
}

func TestPreviewTrack(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	req := getFirstEmailRequest(t)
	openEmail(t, ctx, req.RId)
}

func TestPreviewClick(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	req := getFirstEmailRequest(t)
	clickLink(t, ctx, req.RId, req.Page.HTML)
}

func TestInvalidTransparencyRequest(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	bogusRId := fmt.Sprintf("bogus%s", TransparencySuffix)
	openEmail404(t, ctx, bogusRId)
	clickLink404(t, ctx, bogusRId)
	reportEmail404(t, ctx, bogusRId)
}

func TestTransparencyRequest(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	campaign := getFirstCampaign(t)
	result := campaign.Results[0]
	rid := fmt.Sprintf("%s%s", result.RId, TransparencySuffix)
	transparencyRequest(t, ctx, result, rid, "/")
	transparencyRequest(t, ctx, result, rid, "/track")
	transparencyRequest(t, ctx, result, rid, "/report")

	// And check with the URL encoded version of a +
	rid = fmt.Sprintf("%s%s", result.RId, "%2b")
	transparencyRequest(t, ctx, result, rid, "/")
	transparencyRequest(t, ctx, result, rid, "/track")
	transparencyRequest(t, ctx, result, rid, "/report")
}

func TestRedirectTemplating(t *testing.T) {
	ctx := setupTest(t)
	defer tearDown(t, ctx)
	p := models.Page{
		Name:        "Redirect Page",
		HTML:        "<html>Test</html>",
		UserId:      1,
		RedirectURL: "http://example.com/{{.RId}}",
	}
	err := models.PostPage(&p)
	if err != nil {
		t.Fatalf("error posting new page: %v", err)
	}
	smtp, _ := models.GetSMTP(1, 1)
	template, _ := models.GetTemplate(1, 1)
	group, _ := models.GetGroup(1, 1)

	campaign := models.Campaign{Name: "Redirect campaign"}
	campaign.UserId = 1
	campaign.Template = template
	campaign.Page = p
	campaign.SMTP = smtp
	campaign.Groups = []models.Group{group}
	err = models.PostCampaign(&campaign, campaign.UserId)
	if err != nil {
		t.Fatalf("error creating campaign: %v", err)
	}

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	result := campaign.Results[0]
	resp, err := client.PostForm(fmt.Sprintf("%s/?%s=%s", ctx.phishServer.URL, models.RecipientParameter, result.RId), url.Values{"username": {"test"}, "password": {"test"}})
	if err != nil {
		t.Fatalf("error requesting / endpoint: %v", err)
	}
	defer resp.Body.Close()
	got := resp.StatusCode
	expectedStatus := http.StatusFound
	if got != expectedStatus {
		t.Fatalf("invalid status code received for /track endpoint. expected %d got %d", expectedStatus, got)
	}
	expectedURL := fmt.Sprintf("http://example.com/%s", result.RId)
	gotURL, err := resp.Location()
	if err != nil {
		t.Fatalf("error getting Location header from response: %v", err)
	}
	if gotURL.String() != expectedURL {
		t.Fatalf("invalid redirect received. expected %s got %s", expectedURL, gotURL)
	}
}
