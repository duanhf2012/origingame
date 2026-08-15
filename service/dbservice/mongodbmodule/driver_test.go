package mongodbmodule

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	rpcapi "origingame/protocol/rpc"
)

func TestMarshalBSONValuePreservesTypeAndRawValue(t *testing.T) {
	value, err := marshalBSONValue("player-1")
	if err != nil {
		t.Fatal(err)
	}
	if bson.Type(value.Type) != bson.TypeString {
		t.Fatalf("type=%v, want string", bson.Type(value.Type))
	}
	decoded := bson.RawValue{Type: bson.Type(value.Type), Value: value.Value}
	if got := decoded.StringValue(); got != "player-1" {
		t.Fatalf("decoded=%q", got)
	}
}

func TestClassifyFailureReturnsStableKinds(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		kind         rpcapi.MongoFailureKind
		stateUnknown bool
	}{
		{name: "duplicate", err: mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000}}}, kind: rpcapi.MongoFailureKindDuplicateKey},
		{name: "deadline", err: context.DeadlineExceeded, kind: rpcapi.MongoFailureKindTimeout, stateUnknown: true},
		{name: "canceled", err: context.Canceled, kind: rpcapi.MongoFailureKindCanceled, stateUnknown: true},
		{name: "unknown", err: errors.New("boom"), kind: rpcapi.MongoFailureKindUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			failure := classifyFailure(test.err)
			if failure.Kind != test.kind || failure.StateUnknown != test.stateUnknown {
				t.Fatalf("failure=%+v", failure)
			}
		})
	}
}
