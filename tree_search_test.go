package main

import (
	"testing"
)

func TestSearchSimpleKeys(t *testing.T) {
	raw := []byte(`{
		"alma": "mu",
		"name": "barack",
		"age": 12,
		"tags": ["good", "excellent", "ceu"]
	}`)
	tree, err := fromBytes(raw)
	if err != nil {
		t.Fatalf("failed to convert JSON to tree")
	}
	v, ok := tree.(*complexNode)
	if !ok {
		t.Fatalf("failed to convert tree to complexNode")
	}

	if len(v.keys()) != 4 {
		t.Fatalf("root element should have 4 children")
	}

	subtree, err := tree.search("alma")
	if err != nil {
		t.Fatalf("failed to search tree: %q", err.Error())
	}
	if subtree == nil {
		t.Fatalf("subtree returned nil")
	}
	v2, ok := subtree.(*complexNode)
	if !ok {
		t.Fatalf("failed to convert tree to complexNode")
	}

	if len(v2.keys()) != 1 {
		t.Fatalf("searched subtree element should have 1 children")
	}
	anode, ok := v2.get("alma")
	if !ok {
		t.Fatalf("first node should be alma node")
	}
	str := anode.String(0)
	if str != "\"mu\"" {
		t.Fatalf("searched subtree element should be mu. Instead it was %q", anode.String(0))
	}
}

func TestNode(t *testing.T) {
	raw := []byte(`{"Data":[{"Data":null,"Name":"1","JsonRaw":"[\"1\",\"2\",\"3\"]","Age":1,"DataMap":null},{"Data":null,"Name":"1","JsonRaw":"[\"1\",\"2\",\"3\"]","Age":1,"DataMap":null}],"Name":"2","JsonRaw":"[\"1\",\"2\",\"3\"]","Age":1,"DataMap":{"1":{"Data":null,"Name":"1","JsonRaw":"[\"1\",\"2\",\"3\"]","Age":1,"DataMap":null},"2":{"Data":null,"Name":"1","JsonRaw":"[\"1\",\"2\",\"3\"]","Age":1,"DataMap":null}}}`)
	tree, err := fromBytes(raw)
	if err != nil {
		t.Fatalf("failed to convert JSON to tree")
	}

	t.Log(tree.find([]string{"Data", "0", "JsonRaw"}).String(2))

	data := tree.find([]string{"Data", "0", "JsonRaw"}).String(2)
	t.Log(FormatData(data))
}
