package callbackclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	downstreamcontract "github.com/dujiao-next/internal/modules/downstreamcallback/contract"
	"github.com/dujiao-next/internal/upstream"
)

func TestClientSendsSignedCallback(t *testing.T) {
	const secret = "downstream-secret"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read callback body: %v", err)
			return
		}
		if request.Method != http.MethodPost || request.Header.Get(upstream.HeaderApiKey) != "downstream-key" {
			t.Errorf("unexpected request: method=%s api_key=%q", request.Method, request.Header.Get(upstream.HeaderApiKey))
		}
		timestamp, err := strconv.ParseInt(request.Header.Get(upstream.HeaderTimestamp), 10, 64)
		if err != nil {
			t.Errorf("parse timestamp: %v", err)
			return
		}
		wantSignature := upstream.Sign(secret, http.MethodPost, signaturePath, timestamp, body)
		if request.Header.Get(upstream.HeaderSignature) != wantSignature {
			t.Errorf("signature mismatch: got=%q want=%q", request.Header.Get(upstream.HeaderSignature), wantSignature)
		}
		var payload downstreamcontract.CallbackPayload
		if err := json.Unmarshal(body, &payload); err != nil || payload.Event != "order.fulfilled" {
			t.Errorf("payload mismatch: payload=%#v err=%v", payload, err)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	client := New()
	err := client.Send(context.Background(), downstreamcontract.DeliveryRequest{
		URL:       server.URL,
		APIKey:    "downstream-key",
		APISecret: secret,
		Payload: downstreamcontract.CallbackPayload{
			Event:     "order.fulfilled",
			OrderID:   8,
			Timestamp: 1_700_000_000,
		},
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestClientRejectsNonSuccessContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
		_, _ = writer.Write([]byte("upstream unavailable"))
	}))
	t.Cleanup(server.Close)

	err := New().Send(context.Background(), downstreamcontract.DeliveryRequest{
		URL:     server.URL,
		Payload: downstreamcontract.CallbackPayload{Timestamp: 1_700_000_000},
	})
	if err == nil {
		t.Fatal("Send() should reject non-success response")
	}
}
