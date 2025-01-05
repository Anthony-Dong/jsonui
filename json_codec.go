package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthony-dong/jsonui/internal/orderedmap"
)

func newJsonDecoder(data []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	return decoder
}

func decodeOrderMap(data []byte) (*orderedmap.OrderedMap, error) {
	orderedMap := orderedmap.New()
	orderedMap.SetUseNumber(true)
	orderedMap.SetEscapeHTML(false)
	if err := newJsonDecoder(data).Decode(&orderedMap); err != nil {
		return nil, err
	}
	return orderedMap, nil
}

func decodeInterface(data []byte) (interface{}, error) {
	var result interface{}
	if err := newJsonDecoder(data).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func decodeRawMessage(data json.RawMessage) (interface{}, error) {
	if len(data) >= 2 && data[0] == '{' {
		if orderMap, err := decodeOrderMap(data); err == nil {
			return orderMap, nil
		}
	}
	if len(data) >= 2 && data[0] == '[' {
		if array, err := decodeArray(data); err == nil {
			return array, nil
		}
	}
	return decodeInterface(data)
}

func decodeArray(data []byte) ([]interface{}, error) {
	array := make([]json.RawMessage, 0)
	if err := newJsonDecoder(data).Decode(&array); err != nil {
		return nil, err
	}
	result := make([]interface{}, len(array))
	for index, elem := range array {
		message, err := decodeRawMessage(elem)
		if err != nil {
			return nil, err
		}
		result[index] = message
	}
	return result, nil
}

func decodeJsonData(b []byte) (interface{}, error) {
	return decodeRawMessage(b)
}

func toOrderMap(data interface{}) (*orderedmap.OrderedMap, error) {
	switch value := data.(type) {
	case map[string]interface{}:
		result := orderedmap.NewWithSize(len(value))
		for k, v := range value {
			result.Set(k, v)
		}
		return result, nil
	case *orderedmap.OrderedMap:
		return value, nil
	case orderedmap.OrderedMap:
		return &value, nil
	}
	return nil, fmt.Errorf(`unexpected data type %T`, data)
}

func fromBytes(b []byte) (treeNode, error) {
	b = bytes.TrimSpace(b)
	value, err := decodeJsonData(b)
	if err != nil {
		return nil, err
	}
	return newTree(value)
}

func encodeJson(v interface{}, indent int) string {
	switch vv := v.(type) {
	case *orderedmap.OrderedMap:
		vv.SetUseNumber(true)
		vv.SetEscapeHTML(false)
	case orderedmap.OrderedMap:
		vv.SetUseNumber(true)
		vv.SetEscapeHTML(false)
	}
	out := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)
	if indent > 0 {
		indentStr := strings.Repeat(" ", indent)
		encoder.SetIndent("", indentStr)
	}
	if err := encoder.Encode(v); err != nil {
		return err.Error()
	}
	return strings.TrimSpace(out.String())
}
