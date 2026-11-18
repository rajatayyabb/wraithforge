package triage

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// The structs below mirror the subset of the nmap -oX schema that is useful
// for triage. Fields nmap does not always emit are optional.

type nmapRun struct {
	XMLName xml.Name   `xml:"nmaprun"`
	Hosts   []nmapHost `xml:"host"`
}

type nmapHost struct {
	Addresses []nmapAddress `xml:"address"`
	Hostnames []nmapName    `xml:"hostnames>hostname"`
	Ports     []nmapPort    `xml:"ports>port"`
}

type nmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type nmapName struct {
	Name string `xml:"name,attr"`
}

type nmapPort struct {
	Protocol string       `xml:"protocol,attr"`
	PortID   int          `xml:"portid,attr"`
	State    nmapState    `xml:"state"`
	Service  nmapService  `xml:"service"`
	Scripts  []nmapScript `xml:"script"`
}

type nmapState struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

type nmapService struct {
	Name      string `xml:"name,attr"`
	Product   string `xml:"product,attr"`
	Version   string `xml:"version,attr"`
	ExtraInfo string `xml:"extrainfo,attr"`
	Tunnel    string `xml:"tunnel,attr"`
}

type nmapScript struct {
	ID     string `xml:"id,attr"`
	Output string `xml:"output,attr"`
}

// ParseNmapXML reads nmap XML output (nmap -oX) and returns one Finding per
// open port. Every finding is informational "exposure" context: it records
// that a service is reachable, together with any version banner and NSE script
// output nmap captured. It makes no judgement about exploitability and takes no
// action against any host — it only restructures a scan the operator already
// ran into normalized findings.
func ParseNmapXML(data []byte) ([]Finding, error) {
	var run nmapRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parse nmap xml: %w", err)
	}

	var findings []Finding
	for _, h := range run.Hosts {
		host := hostLabel(h)
		for _, p := range h.Ports {
			if strings.ToLower(p.State.State) != "open" {
				continue
			}
			asset := fmt.Sprintf("%s:%d/%s", host, p.PortID, strings.ToLower(p.Protocol))
			svc := serviceLabel(p.Service)

			title := fmt.Sprintf("Open %s port %d", strings.ToLower(p.Protocol), p.PortID)
			if svc != "" {
				title = fmt.Sprintf("%s (%s)", title, svc)
			}

			findings = append(findings, Finding{
				SourceTool:  "nmap",
				Asset:       asset,
				Title:       title,
				Severity:    SeverityInfo,
				Category:    "exposure",
				Description: exposureDescription(host, p),
				Evidence:    scriptEvidence(p.Scripts),
			})
		}
	}
	return findings, nil
}

func hostLabel(h nmapHost) string {
	// Prefer a resolved hostname, fall back to the first IP address.
	for _, n := range h.Hostnames {
		if n.Name != "" {
			return n.Name
		}
	}
	for _, a := range h.Addresses {
		if a.AddrType == "ipv4" || a.AddrType == "ipv6" {
			return a.Addr
		}
	}
	if len(h.Addresses) > 0 {
		return h.Addresses[0].Addr
	}
	return "unknown-host"
}

func serviceLabel(s nmapService) string {
	parts := make([]string, 0, 3)
	if s.Name != "" {
		parts = append(parts, s.Name)
	}
	if s.Product != "" {
		parts = append(parts, s.Product)
	}
	if s.Version != "" {
		parts = append(parts, s.Version)
	}
	return strings.Join(parts, " ")
}

func exposureDescription(host string, p nmapPort) string {
	svc := serviceLabel(p.Service)
	if svc == "" {
		svc = "an unidentified service"
	}
	reason := p.State.Reason
	if reason == "" {
		reason = "open"
	}
	return fmt.Sprintf(
		"nmap reported %s/%d on %s as open (%s), serving %s. "+
			"Confirm this exposure is intended and the service is patched.",
		strings.ToLower(p.Protocol), p.PortID, host, reason, svc,
	)
}

func scriptEvidence(scripts []nmapScript) string {
	if len(scripts) == 0 {
		return ""
	}
	var b strings.Builder
	for i, s := range scripts {
		if i > 0 {
			b.WriteString("\n")
		}
		out := strings.TrimSpace(s.Output)
		b.WriteString(fmt.Sprintf("[%s] %s", s.ID, out))
	}
	return b.String()
}
