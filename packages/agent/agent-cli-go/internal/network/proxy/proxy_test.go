package proxy

import "testing"

func TestEnv_NilWithoutURL(t *testing.T) {
	if env := (Options{EgressAllowlist: []string{"example.com"}}).Env(); env != nil {
		t.Errorf("Env() = %v, want nil when no proxy URL is set", env)
	}
}

func TestEnv_RoutesAllEgressAndCarriesAllowlist(t *testing.T) {
	env := Options{URL: "http://127.0.0.1:3128", EgressAllowlist: []string{"a.com", "b.com"}}.Env()
	if env["HTTPS_PROXY"] != "http://127.0.0.1:3128" || env["NO_PROXY"] != "" {
		t.Errorf("proxy env = %v, want all egress routed and NO_PROXY empty", env)
	}
	if env[EgressAllowEnvVar] != "a.com,b.com" {
		t.Errorf("%s = %q, want the allowlist joined", EgressAllowEnvVar, env[EgressAllowEnvVar])
	}
}
