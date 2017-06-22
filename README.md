# demo_17_bcss

## Cmd-reference

### Installation:
```bash
go get github.com/dedis/demo_17_bcss/pop
```

Config:
```bash
wget https://pop.dedis.ch/config_bcss.toml
pop client join config_bcss.toml private_key
```

Sign:
```bash
pop client sign message context
```

Verify:
```bash
pop client verify message context signature
```
