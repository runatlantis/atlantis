// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/models"
)

type publicationRecoveryRequest struct {
	Repository                    string `json:"repository"`
	Hostname                      string `json:"hostname"`
	PullNum                       int    `json:"pull_num"`
	ExpectedOwner                 string `json:"expected_owner"`
	OperationStoppedAndReconciled bool   `json:"operation_stopped_and_reconciled"`
}

// PublicationLease inspects an exceptional publication record. Recovery uses the
// same API authentication and repository allowlist as other administrative APIs.
func (a *APIController) PublicationLease(w http.ResponseWriter, r *http.Request) {
	a.publicationLease(w, r, false)
}

// RecoverPublicationLease releases only an expired exact owner after the
// operator has established that its remote request cannot still complete.
// This never modifies the accepted PullStatus or cancels a Terraform process.
func (a *APIController) RecoverPublicationLease(w http.ResponseWriter, r *http.Request) {
	a.publicationLease(w, r, true)
}

func (a *APIController) publicationLease(w http.ResponseWriter, r *http.Request, recoverLease bool) {
	middleware := a.getAPIMiddleware()
	if !middleware.RequireAuth(w, r) {
		return
	}
	responder := middleware.Responder
	if a.PublicationRecovery == nil {
		responder.Error(w, r, http.StatusServiceUnavailable, NewAPIError(ErrCodeServiceUnavailable, "publication recovery is unavailable"))
		return
	}
	var request publicationRecoveryRequest
	if recoverLease {
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			responder.ValidationFailed(w, r, "invalid publication recovery request")
			return
		}
		if request.ExpectedOwner == "" || !request.OperationStoppedAndReconciled {
			responder.ValidationFailed(w, r, "expected_owner and operation_stopped_and_reconciled=true are required; first establish that the original remote operation cannot still complete")
			return
		}
	} else {
		request.Repository = r.URL.Query().Get("repository")
		request.Hostname = r.URL.Query().Get("hostname")
		request.PullNum, _ = strconv.Atoi(r.URL.Query().Get("pull_num"))
	}
	if request.Repository == "" || request.Hostname == "" || request.PullNum <= 0 {
		responder.ValidationFailed(w, r, "repository, hostname, and a positive pull_num are required")
		return
	}
	if a.RepoAllowlistChecker == nil || !a.RepoAllowlistChecker.IsAllowlisted(request.Repository, request.Hostname) {
		responder.Forbidden(w, r, "repository is not in the allowlist")
		return
	}
	pull := models.PullRequest{Num: request.PullNum, BaseRepo: models.Repo{FullName: request.Repository, VCSHost: models.VCSHost{Hostname: request.Hostname}}}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if a.PublicationShutdown != nil {
		stop := context.AfterFunc(a.PublicationShutdown, cancel)
		defer stop()
		if a.PublicationShutdown.Err() != nil {
			cancel()
		}
	}
	if recoverLease {
		err := a.PublicationRecovery.RecoverPublicationLease(ctx, pull, db.PublicationFence{Owner: request.ExpectedOwner})
		if err != nil {
			code := http.StatusInternalServerError
			errorCode := ErrCodeInternal
			if errors.Is(err, db.ErrPublicationOwnerLost) || errors.Is(err, db.ErrPublicationBusy) || errors.Is(err, db.ErrPublicationRecoveryNotNeeded) {
				code, errorCode = http.StatusConflict, ErrCodeConflict
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				code, errorCode = http.StatusServiceUnavailable, ErrCodeServiceUnavailable
			}
			responder.Error(w, r, code, NewAPIError(errorCode, err.Error()))
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	lease, err := a.PublicationRecovery.GetPublicationLease(ctx, pull)
	if err != nil {
		responder.InternalError(w, r, err)
		return
	}
	if lease == nil {
		responder.NotFound(w, r, "no publication lease")
		return
	}
	responder.writeJSON(w, http.StatusOK, struct {
		Owner             string `json:"owner"`
		DeadlineUnixMilli int64  `json:"deadline_unix_milli"`
		Publishing        bool   `json:"publishing"`
	}{lease.Owner, lease.DeadlineUnixMilli, lease.Publishing})
}
