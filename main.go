package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

var knownMIBs = map[string]string{
	"sysDescr":    ".1.3.6.1.2.1.1.1.0",
	"sysObjectID": ".1.3.6.1.2.1.1.2.0",
	"sysUpTime":   ".1.3.6.1.2.1.1.3.0",
	"sysContact":  ".1.3.6.1.2.1.1.4.0",
	"sysName":     ".1.3.6.1.2.1.1.5.0",
	"sysLocation": ".1.3.6.1.2.1.1.6.0",
}

func main() {
	var (
		target    string
		community string
		mib       string
		version   string
		port      int
		timeout   time.Duration
		retries   int
	)

	flag.StringVar(&target, "target", "", "ターゲットのIPアドレス (例: 192.0.2.1)")
	flag.StringVar(&community, "community", "public", "SNMP community string")
	flag.StringVar(&mib, "mib", "sysDescr", "取得するMIB名またはOID (例: sysDescr, .1.3.6.1.2.1.1.1.0)")
	flag.StringVar(&version, "version", "2c", "SNMPバージョン: 1 または 2c")
	flag.IntVar(&port, "port", 161, "SNMPポート")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "タイムアウト時間")
	flag.IntVar(&retries, "retries", 1, "リトライ回数")
	flag.Parse()

	if target == "" {
		fmt.Fprintln(os.Stderr, "ERROR: --target を指定してください")
		flag.Usage()
		os.Exit(2)
	}
	if mib == "" {
		fmt.Fprintln(os.Stderr, "ERROR: --mib を指定してください")
		flag.Usage()
		os.Exit(2)
	}

	if net.ParseIP(target) == nil {
		fmt.Fprintf(os.Stderr, "ERROR: 不正なIPアドレスです: %s\n", target)
		os.Exit(2)
	}

	oid, err := resolveMIB(mib)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(2)
	}

	gs, err := newSNMPClient(target, community, version, port, timeout, retries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: SNMPクライアントの初期化に失敗しました: %v\n", err)
		os.Exit(2)
	}
	defer gs.Conn.Close()

	fmt.Fprintf(os.Stdout, "SNMP walk %s @ %s (%s)\n", mib, target, version)
	err = walkOID(gs, oid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: SNMP walk に失敗しました: %v\n", err)
		os.Exit(1)
	}
}

func resolveMIB(mib string) (string, error) {
	if strings.HasPrefix(mib, ".") {
		return mib, nil
	}
	if strings.Contains(mib, ":") {
		return mib, nil
	}
	if oid, ok := knownMIBs[mib]; ok {
		return oid, nil
	}
	if strings.HasPrefix(mib, "SNMPv2-MIB::") || strings.HasPrefix(mib, "IF-MIB::") {
		return mib, nil
	}
	if isNumericOID(mib) {
		if !strings.HasPrefix(mib, ".") {
			return "." + mib, nil
		}
		return mib, nil
	}
	return "", fmt.Errorf("不明なMIB名またはOIDです: %s", mib)
}

func isNumericOID(s string) bool {
	parts := strings.Split(strings.TrimPrefix(s, "."), ".")
	if len(parts) == 0 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func newSNMPClient(target, community, version string, port int, timeout time.Duration, retries int) (*gosnmp.GoSNMP, error) {
	var snmpVersion gosnmp.SnmpVersion
	switch version {
	case "1":
		snmpVersion = gosnmp.Version1
	case "2c", "2C":
		snmpVersion = gosnmp.Version2c
	default:
		return nil, fmt.Errorf("サポートされていないSNMPバージョンです: %s", version)
	}

	gs := &gosnmp.GoSNMP{
		Target:    target,
		Port:      uint16(port),
		Community: community,
		Version:   snmpVersion,
		Timeout:   timeout,
		Retries:   retries,
		MaxOids:   gosnmp.MaxOids,
	}
	if err := gs.Connect(); err != nil {
		return nil, err
	}
	return gs, nil
}

func walkOID(gs *gosnmp.GoSNMP, oid string) error {
	return gs.Walk(oid, func(pdu gosnmp.SnmpPDU) error {
		value := formatValue(pdu)
		fmt.Printf("%s = %s\n", pdu.Name, value)
		return nil
	})
}

func formatValue(pdu gosnmp.SnmpPDU) string {
	switch pdu.Type {
	case gosnmp.OctetString:
		if b, ok := pdu.Value.([]byte); ok {
			return string(b)
		}
	}
	return fmt.Sprintf("%v", pdu.Value)
}
