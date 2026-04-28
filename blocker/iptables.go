package blocker

import "os/exec"

func BlockIP(ip string) {
	// Check if rule already exists before adding to avoid duplicates
	err := exec.Command("iptables", "-C", "INPUT", "-s", ip, "-j", "DROP").Run()
	if err != nil {
		// Rule doesn't exist yet, add it
		exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP").Run()
	}
}

func UnblockIP(ip string) {
	// Remove all matching rules for this IP
	for {
		err := exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP").Run()
		if err != nil {
			break
		}
	}
}
