package types

import (
	"fmt"
	"github.com/centrifuge/go-substrate-rpc-client/v3/scale"
	"reflect"
	"strconv"
	"strings"
)

type EventField interface {
	RustType() string
	GetInstance() interface{}
	GetType() reflect.Type
}

type PlainEventField struct {
	rustType string
	instance interface{}
}

func (p *PlainEventField) RustType() string {
	return p.rustType
}

func (p *PlainEventField) GetInstance() interface{} {
	return p.instance
}

func (p *PlainEventField) GetType() reflect.Type {
	return reflect.TypeOf(p.instance)
}

type SliceEventField struct {
	UnderlyingElem EventField
}

func (a *SliceEventField) RustType() string {
	return "Vec" + "<" + a.UnderlyingElem.RustType() + ">"
}

func (a *SliceEventField) GetInstance() interface{} {
	myType := a.GetType()
	return reflect.Zero(myType).Interface()
}

func (a *SliceEventField) GetType() reflect.Type {
	typ := a.UnderlyingElem.GetType()
	return reflect.SliceOf(typ)
}

type TupleEventField struct {
	UnderlyingElems []EventField
}

func (t *TupleEventField) GetType() reflect.Type {
	stFields := make([]reflect.StructField, len(t.UnderlyingElems))
	for i, underlyingElem := range t.UnderlyingElems {
		typ := underlyingElem.GetType()
		stFields[i] = reflect.StructField{
			Type: typ,
			Name: "Field" + "_" + strconv.Itoa(i),
		}
	}

	return reflect.StructOf(stFields)
}

func (t *TupleEventField) GetInstance() interface{} {
	myType := t.GetType()
	return reflect.Zero(myType).Interface()
}

func (t *TupleEventField) RustType() string {
	var builder strings.Builder
	builder.WriteString("(")
	for i, underlyingElem := range t.UnderlyingElems {
		builder.WriteString(underlyingElem.RustType())
		if i != len(t.UnderlyingElems) - 1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString(")")
	return builder.String()
}

type OptionEventField struct {
	UnderlyingElem EventField
}

func (o *OptionEventField) GetType() reflect.Type {
	stFields := make([]reflect.StructField, 2)
	stFields[0] = reflect.StructField{
		Name:      "IsPresent",
		Type:      reflect.TypeOf(true),
	}
	stFields[1] = reflect.StructField{
		Name: "Field",
		Type: o.UnderlyingElem.GetType(),
	}
	return reflect.StructOf(stFields)
}

func (o *OptionEventField) GetInstance() interface{} {
	myType := o.GetType()
	return reflect.Zero(myType).Interface()
}

func (o *OptionEventField) RustType() string {
	return "Option" + "<" + o.UnderlyingElem.RustType() + ">"
}


type DynamicEventDecoder struct {
	eventTypeCache map[string]reflect.Type
	rawEventFields map[string]interface{}
	phaseType  reflect.Type
	topicsType  reflect.Type
}

func NewDynamicEventDecoder(rawEventFields map[string]interface{}, phaseInstance Phase, topicInstance []Hash) DynamicEventDecoder {
	return DynamicEventDecoder{
		eventTypeCache: make(map[string]reflect.Type),
		rawEventFields: rawEventFields,
		phaseType:  reflect.TypeOf(phaseInstance),
		topicsType:  reflect.TypeOf(topicInstance),
	}
}

func (d *DynamicEventDecoder) getRawField(argStr string) (interface{}, error) {
	argStr = strings.ToLower(argStr)
	instance, ok := d.rawEventFields[argStr]
	if !ok {
		return nil, fmt.Errorf("unable to find an instance of type: %s", argStr)
	}
	switch instance.(type){
	case EventFieldRedirection:
		eventField, err := d.parseEventField(string(instance.(EventFieldRedirection)))
		if err != nil {
			return nil, err
		}
		return eventField.GetInstance(), nil
	default:
		return instance, nil
	}
}

func (d *DynamicEventDecoder) parseEventField(arg string) (EventField, error) {
	argStr := strings.TrimSpace(strings.ToLower(arg))
	if strings.HasPrefix(argStr, "vec<") {
		underlyingArg := strings.TrimSuffix(strings.TrimPrefix(argStr, "vec<"), ">")
		typ, err := d.parseEventField(underlyingArg)
		if err != nil {
			return nil, err
		}
		return &SliceEventField{
			UnderlyingElem: typ,
		}, nil
	} else if strings.HasPrefix(argStr, "option<") {
		underlyingArg := strings.TrimSuffix(strings.TrimPrefix(argStr, "option<"), ">")
		typ, err := d.parseEventField(underlyingArg)
		if err != nil {
			return nil, err
		}
		return &OptionEventField{
			UnderlyingElem: typ,
		}, nil
	} else if strings.HasPrefix(argStr, "(") {
		tuple := strings.TrimSuffix(strings.TrimPrefix(argStr, "("), ")")
		tupleTypes := strings.Split(tuple, ",")
		tupleEventFields := make([]EventField, len(tupleTypes))
		for i, tupleType := range tupleTypes {
			tupleEventField, err := d.parseEventField(tupleType)
			if err != nil {
				return nil, err
			}
			tupleEventFields[i] = tupleEventField
		}
		return &TupleEventField{
			UnderlyingElems: tupleEventFields,
		}, nil
	} else {
		instance, err := d.getRawField(argStr)
		if err != nil {
			return nil, err
		}
		return &PlainEventField{
			rustType: argStr,
			instance: instance,
		}, nil
	}
}

func (d *DynamicEventDecoder) createArgType(arg string) (reflect.Type, error) {
	eventField, err := d.parseEventField(arg)
	if err != nil {
		return nil, err
	}
	return eventField.GetType(), nil
}

func (d *DynamicEventDecoder) DecodeEvent(decoder *scale.Decoder, moduleName Text, eventMetadata EventMetadata) (interface{}, error) {
	// We need to convert []Type to []string
	eventArguments := eventMetadata.EventArguments()
	strEventArguments := make([]string, len(eventArguments))
	for i, eventArgument := range eventArguments {
		strEventArguments[i] = string(eventArgument)
	}

	eventType, err := d.createEventType(string(moduleName), string(eventMetadata.EventName()), strEventArguments)
	if err != nil {
		return nil, err
	}
	eventInstance := reflect.New(eventType).Interface()
	err = decoder.Decode(eventInstance)
	if err != nil {
		return nil, err
	}
	return eventInstance, nil
}

func (d *DynamicEventDecoder) createEventType(moduleName string, eventName string, eventArguments []string) (reflect.Type, error) {
	eventType, ok := d.eventTypeCache[moduleName + "_" + eventName]
	if ok {
		return eventType, nil
	} else {
		stFields := make([]reflect.StructField, len(eventArguments) + 2)
		stFields[0] = reflect.StructField{
			Name: "PhaseField" + "_" + strconv.Itoa(0),
			Type: d.phaseType,
		}
		for i, arg := range eventArguments {
			argType, err := d.createArgType(arg)
			if err != nil {
				return nil, fmt.Errorf("type instance creation error: %v", err)
			}
			stFields[i + 1] = reflect.StructField{
				Name:      "Field" + "_" + strconv.Itoa(i + 1),
				Type:      argType,
			}
		}

		stFields[len(stFields) - 1] = reflect.StructField{
			Name: "TopicsField" + "_" + strconv.Itoa(len(stFields) - 1),
			Type: d.topicsType,
		}

		eventType := reflect.StructOf(stFields)
		d.eventTypeCache[moduleName + "_" + eventName] = eventType
		return eventType, nil
	}
}

