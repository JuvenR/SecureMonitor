package firewall

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

//  adds a deny rule for the given IP using ufw.
func BlockIP(ip string) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		log.Println("firewall: empty ip, skipping block")
		return
	}

	cmd := exec.Command("sudo", "/usr/sbin/ufw", "deny", "from", ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("firewall: failed to block %s: %v", ip, err)
		return
	}

	log.Printf("firewall: blocked %s", ip)
}

//  removes a deny rule for the given IP using ufw.
func UnblockIP(ip string) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		log.Println("firewall: empty ip, skipping unblock")
		return
	}

	cmd := exec.Command("sudo", "/usr/sbin/ufw", "delete", "deny", "from", ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("firewall: failed to unblock %s: %v", ip, err)
		return
	}

	log.Printf("firewall: unblocked %s", ip)
}
