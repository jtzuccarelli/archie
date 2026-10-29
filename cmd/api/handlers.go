package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jtzuccarelli/archie/internal/call"
	"github.com/jtzuccarelli/archie/internal/store"
)

func (app *application) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type createCallRequest struct {
	Audio           string `json:"audio"`
	TdURL           string `json:"td_url"`
	AgentName       string `json:"agent_name"`
	OfferName       string `json:"offer_name"`
	Disposition     string `json:"disposition"`
	AgentTalkTime   string `json:"agent_talk_time"`
	ForwardDuration string `json:"forward_duration"`
}

type createCallResponse struct {
	ID int64 `json:"id"`
}

type callDetailResponse struct {
	ID              int64    `json:"id"`
	AudioFileURL    string   `json:"audio_file_url"`
	Transcript      *string  `json:"transcript"`
	Flags           []string `json:"flags"`
	FlagCount       int      `json:"flag_count"`
	AgentName       *string  `json:"agent_name"`
	TrackdriveURL   *string  `json:"trackdrive_url"`
	Disposition     string   `json:"disposition"`
	OfferName       string   `json:"offer_name"`
	AgentTalkTime   *int     `json:"agent_talk_time"`
	ForwardDuration *int     `json:"forward_duration"`
}

func (app *application) getCallHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.InfoContext(r.Context(), "get call handler hit", "call_id", r.PathValue("id"))

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "id must be a positive integer", http.StatusBadRequest)
		return
	}

	detail, err := app.store.GetCall(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "call not found", http.StatusNotFound)
		return
	}
	if err != nil {
		app.logger.ErrorContext(r.Context(), "retrieving call", "call_id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	flags := detail.Flags
	if flags == nil {
		flags = []string{}
	}
	response := callDetailResponse{
		ID:              detail.ID,
		AudioFileURL:    detail.AudioFileURL,
		Transcript:      detail.Transcript,
		Flags:           flags,
		FlagCount:       detail.FlagCount,
		AgentName:       detail.AgentName,
		TrackdriveURL:   detail.TrackdriveURL,
		Disposition:     detail.Disposition,
		OfferName:       detail.OfferName,
		AgentTalkTime:   detail.AgentTalkTime,
		ForwardDuration: detail.ForwardDuration,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		app.logger.ErrorContext(r.Context(), "writing call response", "call_id", id, "error", err)
	}
}

type callSummaryResponse struct {
	ID            int64       `json:"id"`
	TdCallID      *string     `json:"td_call_id"`
	Status        call.Status `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	AgentName     *string     `json:"agent_name"`
	OfferName     string      `json:"offer_name"`
	Disposition   string      `json:"disposition"`
	FlagCount     int         `json:"flag_count"`
	TrackdriveURL *string     `json:"trackdrive_url"`
}

type callsListResponse struct {
	Calls  []callSummaryResponse `json:"calls"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

func (app *application) getCallsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()

	limit := 20
	if raw := q.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 20 {
			http.Error(w, "limit must be between 1 and 20", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	offset := 0
	if raw := q.Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			http.Error(w, "offset must be a nonnegative integer", http.StatusBadRequest)
			return
		}
		offset = parsed
	}

	tdCallID := q.Get("td_call_id")

	if tdCallID != "" {
		id, err := uuid.Parse(tdCallID)
		if err != nil {
			http.Error(w, "td_call_id must be a valid UUID", http.StatusBadRequest)
			return
		}
		tdCallID = id.String()
	}

	dateFrom, dateTo, err := parseCallDates(q.Get("date_from"), q.Get("date_to"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filters := call.CallFilters{
		AgentName:   q.Get("agent_name"),
		Disposition: q.Get("disposition"),
		SortBy:      q.Get("sort_by"),
		SortOrder:   q.Get("sort_order"),
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		TdCallID:    tdCallID,
		Limit:       limit,
		Offset:      offset,
	}

	calls, total, err := app.store.GetCalls(r.Context(), filters)

	if err != nil {
		app.logger.ErrorContext(r.Context(), "retrieving calls", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	summaries := make([]callSummaryResponse, 0, len(calls))
	for _, c := range calls {
		summaries = append(summaries, callSummaryResponse{
			ID:            c.ID,
			TdCallID:      c.TdCallID,
			Status:        c.Status,
			CreatedAt:     c.CreatedAt,
			AgentName:     c.AgentName,
			OfferName:     c.OfferName,
			Disposition:   c.Disposition,
			FlagCount:     c.FlagCount,
			TrackdriveURL: c.TrackdriveURL,
		})
	}

	callResponse := callsListResponse{
		Calls:  summaries,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(callResponse); err != nil {
		app.logger.ErrorContext(r.Context(), "writing calls response", "error", err)
	}

}

func (app *application) createCallsHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.InfoContext(r.Context(), "call ingestion started...")

	var req createCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Audio == "" {
		http.Error(w, "audio is required", http.StatusBadRequest)
		return
	}

	tdCallID, err := parseTdCallID(req.TdURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	agentTalkTime, err := parseSeconds(req.AgentTalkTime)
	if err != nil {
		http.Error(w, "agent_talk_time must be a number", http.StatusBadRequest)
		return
	}

	forwardDuration, err := parseSeconds(req.ForwardDuration)
	if err != nil {
		http.Error(w, "forward_duration must be a number", http.StatusBadRequest)
		return
	}

	c := call.Call{
		TdCallID:        &tdCallID,
		AudioFileURL:    req.Audio,
		TrackdriveURL:   &req.TdURL,
		AgentName:       &req.AgentName,
		OfferName:       req.OfferName,
		Disposition:     req.Disposition,
		AgentTalkTime:   agentTalkTime,
		ForwardDuration: forwardDuration,
	}

	id, err := app.store.CreateCall(r.Context(), c)
	if errors.Is(err, store.ErrDuplicateCall) {
		app.logger.DebugContext(r.Context(), "duplicate call ignored", "td_call_id", tdCallID)
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		app.logger.ErrorContext(r.Context(), "creating call", "error", err, "td_call_id", tdCallID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	app.logger.InfoContext(r.Context(), "call created", "call_id", id, "td_call_id", tdCallID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createCallResponse{ID: id})
}

func parseTdCallID(tdURL string) (string, error) {
	if tdURL == "" {
		return "", errors.New("td_url is required")
	}

	u, err := url.Parse(tdURL)
	if err != nil {
		return "", errors.New("td_url is not a valid URL")
	}

	id, err := uuid.Parse(path.Base(u.Path))
	if err != nil {
		return "", errors.New("td_url does not end in a call ID")
	}

	return id.String(), nil
}

func parseSeconds(s string) (*int, error) {
	if s == "" {
		return nil, nil
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}

	seconds := int(f)
	return &seconds, nil
}
