// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

func TestIsLeftoverE2EHook(t *testing.T) {
	cases := map[string]bool{
		"https://77a8cc2b0877.ngrok.app/events":      true,
		"https://abc.ngrok-free.app/events":          true,
		"https://ABC.NGROK.IO/events":                true,
		"https://abc.ngrok.dev/events":               true,
		"https://ngrok.app/events":                   false, // no tunnel subdomain
		"https://abc.ngrok.app/other":                false,
		"https://abc.ngrok.app/events/":              false,
		"http://abc.ngrok.app/events":                false,
		"https://atlantis.example.com/events":        false,
		"https://abc.ngrok.app.example.com/events":   false,
		"https://evilngrok.app/events":               false,
		"https://hooks.slack.com/services/T/B/token": false,
		"::not a url":                                false,
		"":                                           false,
	}
	for in, want := range cases {
		if got := isLeftoverE2EHook(in); got != want {
			t.Errorf("isLeftoverE2EHook(%q) = %v, want %v", in, got, want)
		}
	}
}

type fakeHookClient struct {
	VCSClient
	hooks     []Webhook
	listErr   error
	deleteErr error
	deleted   []int64
}

func (f *fakeHookClient) ListAtlantisHooks(context.Context) ([]Webhook, error) {
	return f.hooks, f.listErr
}

func (f *fakeHookClient) DeleteAtlantisHook(_ context.Context, id int64) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, id)
	return nil
}

func TestDeleteLeftoverHooks(t *testing.T) {
	f := &fakeHookClient{hooks: []Webhook{
		{ID: 1, URL: "https://a1.ngrok.app/events"},
		{ID: 2, URL: "https://atlantis.example.com/events"},
		{ID: 3, URL: "https://b2.ngrok-free.app/events"},
		{ID: 4, URL: "https://hooks.slack.com/x"},
	}}
	if err := deleteLeftoverHooks(context.Background(), f); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(f.deleted, []int64{1, 3}) {
		t.Fatalf("deleted %v, want [1 3]", f.deleted)
	}
}

func TestDeleteLeftoverHooksErrors(t *testing.T) {
	boom := errors.New("boom")
	if err := deleteLeftoverHooks(context.Background(), &fakeHookClient{listErr: boom}); !errors.Is(err, boom) {
		t.Fatalf("list error not returned: %v", err)
	}
	f := &fakeHookClient{hooks: []Webhook{{ID: 1, URL: "https://a1.ngrok.app/events"}}, deleteErr: boom}
	if err := deleteLeftoverHooks(context.Background(), f); !errors.Is(err, boom) {
		t.Fatalf("delete error not returned: %v", err)
	}
}

func TestSleepReturnsWhenContextIsDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	sleep(ctx, time.Minute)
	if time.Since(start) > time.Second {
		t.Fatal("sleep ignored a cancelled context")
	}
}
