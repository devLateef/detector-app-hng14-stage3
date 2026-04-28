package blocker

import "os/exec"

func BlockIP(ip string) {
	exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP").Run()
}

func UnblockIP(ip string) {
	exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP").Run()
}
