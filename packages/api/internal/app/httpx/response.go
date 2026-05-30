package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// pathInt64 reads a path parameter and parses it as int64. Returns the parsed
// value and ok=true on success; on failure writes a 400 to w and returns ok=false.
func pathInt64(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	raw := r.PathValue(name)
	id, err := strconv.ParseInt(raw, 10, 64)

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid %s: %q", name, raw))

		return 0, false
	}

	return id, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		fmt.Fprintf(os.Stderr, "json encode error: %v\n", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeSSE(w http.ResponseWriter, rc *http.ResponseController, event string, data any) error {
	payload, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload); err != nil {
		return err
	}

	return rc.Flush()
}
