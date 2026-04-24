package bedrock

import (
	"encoding/json"
	"testing"
)

func TestTitanEmbedRequest_Marshal(t *testing.T) {
	req := titanEmbedRequest{InputText: "hello world"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"inputText":"hello world"}`
	if string(data) != want {
		t.Errorf("got %s, want %s", data, want)
	}
}

func TestTitanEmbedResponse_Unmarshal(t *testing.T) {
	raw := `{"embedding":[0.1,0.2,0.3],"inputTextTokenCount":3}`
	var resp titanEmbedResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Embedding) != 3 {
		t.Fatalf("embedding len = %d, want 3", len(resp.Embedding))
	}
	if resp.Embedding[0] != 0.1 {
		t.Errorf("embedding[0] = %f, want 0.1", resp.Embedding[0])
	}
}
