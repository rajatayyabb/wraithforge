package triage

import "testing"

func TestParseSeverity(t *testing.T) {
	cases := map[string]Severity{
		"CRITICAL": SeverityCritical,
		"crit":     SeverityCritical,
		"High":     SeverityHigh,
		"moderate": SeverityMedium,
		"warning":  SeverityMedium,
		"low":      SeverityLow,
		"":         SeverityInfo,
		"nonsense": SeverityInfo,
	}
	for in, want := range cases {
		if got := ParseSeverity(in); got != want {
			t.Errorf("ParseSeverity(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeDedupKeepsMostSevere(t *testing.T) {
	in := []Finding{
		{SourceTool: "scanner", Asset: "pkg:lodash", Title: "CVE-1", Category: "dep", Severity: SeverityLow},
		{SourceTool: "scanner", Asset: "pkg:lodash", Title: "CVE-1", Category: "dep", Severity: SeverityHigh, Evidence: "fuller"},
		{SourceTool: "nmap", Asset: "host:22/tcp", Title: "Open tcp port 22", Category: "exposure", Severity: SeverityInfo},
	}
	out := Normalize(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 findings after dedup, got %d", len(out))
	}
	// Most severe first.
	if out[0].Severity != SeverityHigh {
		t.Errorf("expected High first, got %q", out[0].Severity)
	}
	if out[0].Evidence != "fuller" {
		t.Errorf("dedup should keep the richer/most-severe copy, got evidence %q", out[0].Evidence)
	}
}

func TestNormalizeFillsFingerprintAndSeverity(t *testing.T) {
	out := Normalize([]Finding{{SourceTool: "x", Asset: "a", Title: "t"}})
	if len(out) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(out))
	}
	if out[0].Fingerprint == "" {
		t.Error("fingerprint should be filled")
	}
	if out[0].Severity != SeverityInfo {
		t.Errorf("missing severity should default to info, got %q", out[0].Severity)
	}
}

func TestCounts(t *testing.T) {
	c := Counts([]Finding{
		{Severity: SeverityHigh}, {Severity: SeverityHigh}, {Severity: SeverityLow},
	})
	if c[SeverityHigh] != 2 || c[SeverityLow] != 1 {
		t.Errorf("unexpected counts: %+v", c)
	}
}

const sampleNmapXML = `<?xml version="1.0"?>
<nmaprun scanner="nmap">
  <host>
    <address addr="192.0.2.10" addrtype="ipv4"/>
    <hostnames><hostname name="web.example.test"/></hostnames>
    <ports>
      <port protocol="tcp" portid="22">
        <state state="open" reason="syn-ack"/>
        <service name="ssh" product="OpenSSH" version="8.9"/>
      </port>
      <port protocol="tcp" portid="80">
        <state state="open" reason="syn-ack"/>
        <service name="http" product="nginx" version="1.24.0"/>
        <script id="http-title" output="Example"/>
      </port>
      <port protocol="tcp" portid="81">
        <state state="closed" reason="reset"/>
        <service name="hosts2-ns"/>
      </port>
    </ports>
  </host>
</nmaprun>`

func TestParseNmapXML(t *testing.T) {
	findings, err := ParseNmapXML([]byte(sampleNmapXML))
	if err != nil {
		t.Fatalf("ParseNmapXML error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 open-port findings (closed port skipped), got %d", len(findings))
	}
	for _, f := range findings {
		if f.SourceTool != "nmap" || f.Category != "exposure" || f.Severity != SeverityInfo {
			t.Errorf("unexpected finding shape: %+v", f)
		}
		if f.Asset == "" || f.Title == "" {
			t.Errorf("finding missing asset/title: %+v", f)
		}
	}
}

func TestParseScannerJSONObject(t *testing.T) {
	doc := `{"tool":"trivy","findings":[
		{"target":"pkg:npm/lodash@4.17.11","id":"CVE-2019-10744","severity":"HIGH","type":"vulnerable-dependency","description":"prototype pollution"}
	]}`
	findings, err := ParseScannerJSON([]byte(doc))
	if err != nil {
		t.Fatalf("ParseScannerJSON error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.SourceTool != "trivy" || f.Severity != SeverityHigh || f.Asset == "" {
		t.Errorf("unexpected finding: %+v", f)
	}
}

func TestParseScannerJSONArray(t *testing.T) {
	doc := `[{"asset":"login.php","title":"Reflected XSS","severity":"medium","category":"web"}]`
	findings, err := ParseScannerJSON([]byte(doc))
	if err != nil {
		t.Fatalf("ParseScannerJSON error: %v", err)
	}
	if len(findings) != 1 || findings[0].Severity != SeverityMedium {
		t.Errorf("unexpected: %+v", findings)
	}
}

func TestParseScannerJSONEmpty(t *testing.T) {
	if _, err := ParseScannerJSON([]byte("   ")); err == nil {
		t.Error("expected error on empty input")
	}
}
