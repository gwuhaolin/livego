#/bin/bash

if [ $# -ne 1 ]
then
	CHANNEL="${RANDOM}-${RANDOM}-${RANDOM}"
else
	CHANNEL="${1}"
fi

echo ${CHANNEL}
curl "http://localhost:8090/control/get?room=${CHANNEL}"
echo
