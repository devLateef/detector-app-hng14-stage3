package blocker

import "os/exec"

// BlockIP adds an iptables DROP rule for the given IP.
// It checks first to avoid duplicate rules.
func BlockIP(ip string) {
	// -C checks if the rule exists; non-zero exit means it doesn't
	err := exec.Command("iptables", "-C", "INPUT", "-s", ip, "-j", "DROP").Run()
	if err != nil {
		exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP").Run()
	}
}

// UnblockIP removes all iptables DROP rules for the given IP.
func UnblockIP(ip string) {
	// Loop until no more matching rules exist
	for {
		err := exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP").Run()
		if err != nil {
			break
		}
	}
}
