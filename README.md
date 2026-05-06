# go-snmpwalk

Go 1.21 以上で動作する、SNMP walk 用の簡易 CLI ツールです。

## 使い方

```bash
go run . --target 192.0.2.1 --mib sysDescr
```

### サポート項目

- `--target`: 取得対象のIPアドレス
- `--community`: SNMP community (デフォルト: `public`)
- `--mib`: MIB名またはOID。`sysDescr` などの既知MIB名、または `.1.3.6.1.2.1.1.1.0` のようなOID
- `--version`: `1` または `2c` (デフォルト: `2c`)
- `--port`: SNMPポート (デフォルト: `161`)
- `--timeout`: タイムアウト (デフォルト: `5s`)
- `--retries`: リトライ回数 (デフォルト: `1`)

## 例

```bash
go run . --target 192.0.2.1 --mib sysDescr

# snmpd側に exec <チェック名> <チェックコマンド> で定義されている監視をチェックするコマンドライン
go run . --target 192.0.2.1 --mib ET-SNMP-EXTEND-MIB::nsExtendOutput1Line."mycheck"
```
