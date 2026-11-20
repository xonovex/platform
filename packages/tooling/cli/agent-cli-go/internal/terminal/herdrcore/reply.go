package herdrcore

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// The replies below are the JSON that hc prints with --json: the result object
// of the matching method of the herdr-core API (hc schema), with no envelope,
// and an error object on standard error. The replies of hc run and hc wait are
// known from that schema alone, so each parser reads only the fields the schema
// documents, ignores every other field and refuses a reply without the field it
// needs.

// waitResult is the part of a WaitResult that the wait loop reads.
type waitResult struct {
	Ended    bool   `json:"ended"`
	Reason   string `json:"reason"`
	ExitCode *int   `json:"exit_code"`
	Signal   *int   `json:"signal"`
	Status   string `json:"status"`
	Turn     uint64 `json:"turn"`
}

// unchanged reports whether the program reports the same status in the same
// turn as in the earlier result.
func (r waitResult) unchanged(earlier waitResult) bool {
	return r.Reason == earlier.Reason && r.Turn == earlier.Turn && r.Status == earlier.Status
}

// terminalState is the part of a Terminal that tells whether it has finished.
type terminalState struct {
	State    string `json:"state"`
	ExitCode *int   `json:"exit_code"`
	Signal   *int   `json:"signal"`
}

// failure is an error hc reported: its kind, empty when hc named none, and the
// text that explains it.
type failure struct {
	kind string
	text string
}

// parseAddress returns the address of the RunResult that hc run prints.
func parseAddress(stdout string) (string, error) {
	var reply struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal([]byte(stdout), &reply); err != nil || reply.Address == "" {
		return "", fmt.Errorf("hc run named no terminal: %s", stdout)
	}
	return reply.Address, nil
}

// parseWait reads the WaitResult that hc wait prints. A reply without a reason
// is not a WaitResult, because the schema requires one.
func parseWait(stdout string) (waitResult, error) {
	var result waitResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil || result.Reason == "" {
		return waitResult{}, fmt.Errorf("not a wait result: %s", stdout)
	}
	return result, nil
}

// parseTerminal reads the Terminal that hc get prints.
func parseTerminal(stdout string) (terminalState, error) {
	var state terminalState
	if err := json.Unmarshal([]byte(stdout), &state); err != nil || state.State == "" {
		return terminalState{}, fmt.Errorf("hc get named no terminal state: %s", stdout)
	}
	return state, nil
}

// parseFailure reads the error object of a failed hc command from its standard
// error, then from its standard output. The command line client prints the kind
// beside the message; the API nests it under data, so both places are read.
func parseFailure(stdout, stderr string, code int) failure {
	for _, output := range []string{stderr, stdout} {
		var reply struct {
			Error *struct {
				Kind    string `json:"kind"`
				Message string `json:"message"`
				Data    struct {
					Kind string `json:"kind"`
				} `json:"data"`
			} `json:"error"`
		}
		if json.Unmarshal([]byte(output), &reply) != nil || reply.Error == nil {
			continue
		}
		kind := reply.Error.Kind
		if kind == "" {
			kind = reply.Error.Data.Kind
		}
		return failure{kind: kind, text: firstOf(output, reply.Error.Message, kind)}
	}
	return failure{text: firstOf("exit code "+strconv.Itoa(code), stderr, stdout)}
}

// firstOf returns the first value that is not empty, or fallback.
func firstOf(fallback string, values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return fallback
}
