#!/usr/bin/env bash

NBR_CONODES=3

for pop in $( seq $NBR_CONODES ); do
	./pop -c running/pop$pop $@
done
