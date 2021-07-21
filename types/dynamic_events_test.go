package types

import (
	"bytes"
	"crypto/rand"
	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func generateRandom32Array() [32]byte {
	var authorityId [32]byte
	slice := make([]byte, 32)
	_, err := rand.Read(slice)
	if err != nil {
		panic(err)
	}
	copy(authorityId[:], slice)
	return authorityId
}

func TestDynamicEventDecoder(t *testing.T) {
	dummyRawEventFields := map[string]interface{} {
		"first": uint64(1),
		"second": "str",
		"third": 1,
		"fourth": EventFieldRedirection("second"),
		"fifth": EventFieldRedirection("notexists"),
		"sixth": EventFieldRedirection("fourth"),
	}

	dynamicEventDecoder := NewDynamicEventDecoder(dummyRawEventFields, Phase{}, []Hash{})
	instance, err := dynamicEventDecoder.getRawField("first")
	assert.NoError(t, err)
	assert.Equal(t, dummyRawEventFields["first"], instance)

	// Redirection works
	instance, err = dynamicEventDecoder.getRawField("sixth")
	assert.NoError(t, err)
	assert.Equal(t, dummyRawEventFields["second"], instance)

	// Invalid redirection gives an error
	instance, err = dynamicEventDecoder.getRawField("fifth")
	assert.EqualError(t, err, "unable to find an instance of type: notexists")
	assert.Equal(t, nil, instance)

	dummyEventMetadata := EventMetadataV4{
		Name: "DummyEvent",
		Args: []Type{
			"First",
			"Vec<(Second, Third)>",
			"( Fourth )",
			"Option<(Second, Fourth)>",
		},
	}

	// We need to convert []Type to []string
	eventArguments := dummyEventMetadata.Args
	strEventArguments := make([]string, len(eventArguments))
	for i, eventArgument := range eventArguments {
		strEventArguments[i] = string(eventArgument)
	}

	typ, err := dynamicEventDecoder.createEventType("MyModule", string(dummyEventMetadata.Name), strEventArguments)

	assert.NoError(t, err)
	// Phase
	assert.Equal(t, reflect.Struct, typ.Field(0).Type.Kind())
	// First type
	assert.Equal(t, reflect.Uint64, typ.Field(1).Type.Kind())
	// Vec<(Second, Third)> type
	assert.Equal(t, reflect.Slice, typ.Field(2).Type.Kind())
	// Type of Second
	assert.Equal(t, reflect.String, typ.Field(2).Type.Elem().Field(0).Type.Kind())
	// Type of Third
	assert.Equal(t, reflect.Int, typ.Field(2).Type.Elem().Field(1).Type.Kind())
	// ( Fourth ) type
	assert.Equal(t, reflect.Struct, typ.Field(3).Type.Kind())
	// Option<(Second, Fourth)>
	assert.Equal(t, reflect.Struct, typ.Field(4).Type.Kind())
	// IsPresent field of Option
	assert.Equal(t, reflect.Bool, typ.Field(4).Type.Field(0).Type.Kind())
	// Main struct of Option
	assert.Equal(t, reflect.Struct, typ.Field(4).Type.Field(1).Type.Kind())
}

func TestDynamicEventDecoding(t *testing.T) {
	dynamicEventDecoder := NewDynamicEventDecoder(RawEventFields, Phase{}, []Hash{})
	testEvent := EventGrandpaNewAuthorities{
		Phase: Phase{
			IsApplyExtrinsic: true,
			AsApplyExtrinsic: 10,
			IsFinalization:   false,
			IsInitialization: false,
		},
		NewAuthorities: []struct {
			AuthorityID     AuthorityID
			AuthorityWeight U64
		}{
			{
				AuthorityID: generateRandom32Array(),
				AuthorityWeight: U64(3),
			},
			{
				AuthorityID: generateRandom32Array(),
				AuthorityWeight: U64(4),
			},
			{
				AuthorityID: generateRandom32Array(),
				AuthorityWeight: U64(5),
			},
		},
		Topics: []Hash{
			generateRandom32Array(),
			generateRandom32Array(),
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	encoder := scale.NewEncoder(buffer)
	err := encoder.Encode(testEvent)
	assert.NoError(t, err)
	testEventEncodedData := buffer.Bytes()

	decoder := scale.NewDecoder(bytes.NewReader(testEventEncodedData))

	eventMetadata := EventMetadataV4{
		Name: "NewAuthorities",
		Args: []Type{"Vec< ( AuthorityId, AuthorityWeight)>"},
	}
	dynamicEvent, err := dynamicEventDecoder.DecodeEvent(decoder, "Grandpa",  &eventMetadata)
	assert.NoError(t, err)

	buffer = bytes.NewBuffer([]byte{})
	encoder = scale.NewEncoder(buffer)
	err = encoder.Encode(dynamicEvent)
	assert.NoError(t, err)
	dynamicEventEncodedData := buffer.Bytes()
	assert.Equal(t, dynamicEventEncodedData, testEventEncodedData)


	optionalValue := generateRandom32Array()
	optionBytes := OptionBytes{}
	optionBytes.SetSome(optionalValue[:])
	testEvent2 := EventSchedulerDispatched{
		Phase: Phase{
			IsApplyExtrinsic: true,
			AsApplyExtrinsic: 111,
			IsFinalization:   false,
			IsInitialization: false,
		},
		Task: TaskAddress{
			Index: 1,
			When: 4,
		},
		ID: optionBytes,
		Result: DispatchResult{
			Ok: true,
			Error: DispatchError{},
		},
		Topics: []Hash{
			generateRandom32Array(),
			generateRandom32Array(),
		},
	}
	buffer = bytes.NewBuffer([]byte{})
	encoder = scale.NewEncoder(buffer)
	err = encoder.Encode(testEvent2)
	assert.NoError(t, err)
	testEventEncodedData = buffer.Bytes()

	decoder = scale.NewDecoder(bytes.NewReader(testEventEncodedData))

	eventMetadata = EventMetadataV4{
		Name: "Dispatched",
		Args: []Type{"TaskAddress<BlockNumber>", "Option<Bytes>", "DispatchResult"},
	}
	dynamicEvent, err = dynamicEventDecoder.DecodeEvent(decoder, "Scheduler",  &eventMetadata)
	assert.NoError(t, err)
	buffer = bytes.NewBuffer([]byte{})
	encoder = scale.NewEncoder(buffer)
	err = encoder.Encode(dynamicEvent)
	assert.NoError(t, err)
	dynamicEventEncodedData = buffer.Bytes()
	assert.Equal(t, testEventEncodedData, dynamicEventEncodedData)
}
