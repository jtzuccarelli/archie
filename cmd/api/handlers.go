package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/google/uuid"
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

func (app *application) createCallsHandler(w http.ResponseWriter, r *http.Request) {
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
		app.logger.Info("duplicate call ignored", "td_call_id", tdCallID)
		w.WriteHeader(http.StatusOK)
		return
	}
	if err != nil {
		app.logger.Error("creating call", "error", err, "td_call_id", tdCallID)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	app.logger.Info("call created", "id", id, "td_call_id", tdCallID)

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
