package downstreamcallback_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/modules/downstreamcallback"
	"github.com/dujiao-next/internal/queue"
	"github.com/dujiao-next/internal/upstream"

	"github.com/hibiken/asynq"
)

type refRepositoryStub struct {
	byID      map[uint]*models.DownstreamOrderRef
	byOrderID map[uint]*models.DownstreamOrderRef
	created   *models.DownstreamOrderRef
	updated   []*models.DownstreamOrderRef
}

func (r *refRepositoryStub) GetByID(id uint) (*models.DownstreamOrderRef, error) {
	return r.byID[id], nil
}

func (r *refRepositoryStub) GetByOrderID(orderID uint) (*models.DownstreamOrderRef, error) {
	return r.byOrderID[orderID], nil
}

func (r *refRepositoryStub) Create(ref *models.DownstreamOrderRef) error {
	r.created = ref
	return nil
}

func (r *refRepositoryStub) Update(ref *models.DownstreamOrderRef) error {
	r.updated = append(r.updated, ref)
	return nil
}

type orderReaderStub struct {
	orders map[uint]*models.Order
}

func (r orderReaderStub) GetByID(id uint) (*models.Order, error) {
	return r.orders[id], nil
}

type credentialReaderStub struct {
	credentials map[uint]*models.ApiCredential
}

func (r credentialReaderStub) GetByID(id uint) (*models.ApiCredential, error) {
	return r.credentials[id], nil
}

type callbackQueueStub struct {
	payloads []queue.DownstreamCallbackPayload
}

func (q *callbackQueueStub) EnqueueDownstreamCallback(payload queue.DownstreamCallbackPayload, _ ...asynq.Option) error {
	q.payloads = append(q.payloads, payload)
	return nil
}

func TestCreateRefInitializesPendingStatus(t *testing.T) {
	references := &refRepositoryStub{}
	service := downstreamcallback.NewService(references, orderReaderStub{}, credentialReaderStub{}, nil)
	ref := &models.DownstreamOrderRef{OrderID: 42, CallbackStatus: constants.CallbackStatusFailed}

	if err := service.CreateRef(ref); err != nil {
		t.Fatalf("CreateRef() error = %v", err)
	}
	if references.created != ref || ref.CallbackStatus != constants.CallbackStatusPending {
		t.Fatalf("created ref mismatch: %#v", references.created)
	}
	if err := service.CreateRef(nil); err == nil {
		t.Fatal("CreateRef(nil) should fail")
	}
}

func TestEnqueueCallbackResolvesParentAndResetsRetryState(t *testing.T) {
	parentID := uint(17)
	ref := &models.DownstreamOrderRef{
		ID:                 9,
		OrderID:            parentID,
		CallbackURL:        "https://callback.example.test",
		CallbackStatus:     constants.CallbackStatusFailed,
		CallbackRetryCount: 4,
	}
	references := &refRepositoryStub{byOrderID: map[uint]*models.DownstreamOrderRef{parentID: ref}}
	orders := orderReaderStub{orders: map[uint]*models.Order{21: {ID: 21, ParentID: &parentID}}}
	queued := &callbackQueueStub{}
	service := downstreamcallback.NewService(references, orders, credentialReaderStub{}, queued)

	service.EnqueueCallback(21)

	if len(queued.payloads) != 1 || queued.payloads[0].DownstreamOrderRefID != ref.ID {
		t.Fatalf("queued payloads = %#v", queued.payloads)
	}
	if ref.CallbackStatus != constants.CallbackStatusPending || ref.CallbackRetryCount != 0 {
		t.Fatalf("ref was not reset: %#v", ref)
	}
	if len(references.updated) != 1 {
		t.Fatalf("updates = %d, want 1", len(references.updated))
	}
}

func TestSendCallbackSignsRequestAndPersistsSentStatus(t *testing.T) {
	const apiSecret = "downstream-secret"
	var received bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		received = true
		if request.Method != http.MethodPost || request.Header.Get(upstream.HeaderApiKey) != "downstream-key" {
			t.Errorf("unexpected callback request: method=%s api_key=%q", request.Method, request.Header.Get(upstream.HeaderApiKey))
		}
		var body map[string]interface{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode callback body: %v", err)
		}
		if body["event"] != "order.fulfilled" || request.Header.Get(upstream.HeaderSignature) == "" {
			t.Errorf("callback contract mismatch: body=%#v signature=%q", body, request.Header.Get(upstream.HeaderSignature))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	ref := &models.DownstreamOrderRef{
		ID:                3,
		OrderID:           8,
		ApiCredentialID:   5,
		DownstreamOrderNo: "remote-8",
		CallbackURL:       server.URL,
		CallbackStatus:    constants.CallbackStatusPending,
	}
	references := &refRepositoryStub{byID: map[uint]*models.DownstreamOrderRef{ref.ID: ref}}
	orders := orderReaderStub{orders: map[uint]*models.Order{8: {
		ID:      8,
		OrderNo: "DJ-8",
		Status:  constants.OrderStatusCompleted,
	}}}
	credentials := credentialReaderStub{credentials: map[uint]*models.ApiCredential{5: {
		ID:        5,
		ApiKey:    "downstream-key",
		ApiSecret: apiSecret,
	}}}
	service := downstreamcallback.NewService(references, orders, credentials, nil)

	if err := service.SendCallback(ref.ID); err != nil {
		t.Fatalf("SendCallback() error = %v", err)
	}
	if !received || ref.CallbackStatus != constants.CallbackStatusSent || ref.LastCallbackAt == nil {
		t.Fatalf("callback result mismatch: received=%v ref=%#v", received, ref)
	}
	if len(references.updated) != 1 {
		t.Fatalf("updates = %d, want 1", len(references.updated))
	}
}

func TestSendCallbackReturnsStableNotFoundSentinel(t *testing.T) {
	service := downstreamcallback.NewService(&refRepositoryStub{}, orderReaderStub{}, credentialReaderStub{}, nil)
	if err := service.SendCallback(404); !errors.Is(err, downstreamcallback.ErrRefNotFound) {
		t.Fatalf("error = %v, want ErrRefNotFound", err)
	}
}
