#!/usr/bin/env bash
KEYS=""
for k in $@; do
	KEYS="$KEYS, \"$( echo $k | sed -e "s/ed25519pub://" )\""
done
echo [${KEYS:2}]
