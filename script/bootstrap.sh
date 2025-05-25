#!/bin/bash
CURDIR=$(cd $(dirname $0); pwd)

if [ ! -f $CURDIR/settings.py ]; then
    echo "there is no settings.py exist."
    exit -1
fi

MODULE=$(cd $CURDIR; python -c "import settings; print settings.MODULE")

SVC_NAME=${MODULE}

BinaryName=pokerservice

export PSM=$SVC_NAME
CONF_DIR=$CURDIR/conf/

args="-psm=$SVC_NAME -conf-dir=$CONF_DIR"

echo "$CURDIR/bin/${BinaryName} $args"

exec $CURDIR/bin/${BinaryName} $args
