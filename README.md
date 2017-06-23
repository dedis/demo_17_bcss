# Proof-of-Personhood demo at Blockchain-Summer-School 2017

Proof-of-personhood is a novel system to attribute exactly one cryptographic
token to exactly one physical person. It uses linkable ring
 signatures to identify yourself as being part of the group.
 This gives you an anonymity-set the size of the participating
 group.

## Attendee - participant

If you participated in the pop-party at blockchain summer school
2017 in EPFL, you can use your tokens in the following way:

- Installation - we suppose you already have a running go 1.8 version or later, according to https://golang.org/doc/install
```bash
go get github.com/dedis/demo_17_bcss/pop
```

- Linking your private key to the transcript
```bash
wget https://pop.dedis.ch/transcript_bcss.toml
pop client join transcript_bcss.toml private_key
```

- Signing a message coming from a service
```bash
pop client sign message context
```

- Verifying on the service side
```bash
pop client verify message context signature
```

## Organizer

If you want to organize your own pop-party, please have a look at the
following readme: [pop/README.md]

## Papers

For further reading, here some papers that have been written by the DEDIS-group:

* First ideas of pseudonym parties: http://www.brynosaurus.com/log/2007/0327-PseudonymParties.pdf
* Proof-of-Personhood
* Proof-of-Personhood: Redemocratizing Permissionless Cryptocurrencies: https://zerobyte.io/publications/2017-BKJGGF-pop.pdf