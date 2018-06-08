package test

import (
	"testing"
)

// TestSaveData
func TestSaveData(t *testing.T) {
	t.Log("TestSaveData")
	err := httpDo("POST", "http://127.0.0.1:8888/factor/saveData", "", getSaveDataRequest("222"), t)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	t.Log("TestSaveData success")
}

// TestDslQuery
func TestDslQuery(t *testing.T) {
	t.Log("TestDslQuery")
	request := "{\"selector\":{\"sender\":\"222\"}}"
	err := httpDo("POST", "http://127.0.0.1:8888/factor/dslQuery/", "", []byte(request), t)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	t.Log("TestDslQuery success")
}

// TestBlockQuery
func TestBlockQuery(t *testing.T) {
	t.Log("TestBlockQuery")
	txId := ""
	err := httpDo("GET", "http://127.0.0.1:8888/factor/"+txId+"/block/", "", nil, t)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	t.Log("TestBlockQuery success")
}
