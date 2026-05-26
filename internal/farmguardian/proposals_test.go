package farmguardian

import "testing"

func TestMatchAlertToolIntent(t *testing.T) {
	tool, ok := matchAlertToolIntent("Please acknowledge the humidity alert in Flower Room")
	if !ok || tool != "ack_alert" {
		t.Fatalf("got tool=%q ok=%v want ack_alert", tool, ok)
	}
	tool, ok = matchAlertToolIntent("mark alert #12 as read")
	if !ok || tool != "mark_alert_read" {
		t.Fatalf("got tool=%q ok=%v want mark_alert_read", tool, ok)
	}
	if _, ok := matchAlertToolIntent("what is the weather"); ok {
		t.Fatal("expected no intent for unrelated question")
	}
}

func TestPickAlertForIntent(t *testing.T) {
	details := []UnreadAlertDetail{
		{ID: 1, Subject: "OHN batch below minimum"},
		{ID: 4, Subject: "Humidity high — Flower Room"},
	}
	got := pickAlertForIntent("acknowledge the humidity alert", details)
	if got.ID != 4 {
		t.Fatalf("pick by keyword: got id %d want 4", got.ID)
	}
	got = pickAlertForIntent("ack alert #1", details)
	if got.ID != 1 {
		t.Fatalf("pick by id: got %d want 1", got.ID)
	}
}
