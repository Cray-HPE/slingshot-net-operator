#!/bin/sh

# Check if Operator mode is enabled
if [ x$ADMINISTRATIVE_STATE = xenable ]; then
    echo "Operator mode is enabled"
    # Start the Operator
    /root/sshot-net-operator
else
    echo "Slingshot network operator is stopped since ADMINISTRATIVE STATE is disabled. To enable, set the ADMINISTRATIVE_STATE variable to enable in values.yaml in slingshot network operator helm chart and reinstall the helm chart."

    # loop to wait
    while true; do
        sleep 3600
    done
fi