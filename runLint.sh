#/bin/bash -x

if [[ -x /usr/bin/docker ]]; then
    CONTAINER_CMD=docker
    PRIVATE_UNSHARED=:Z
elif [[ -x /usr/bin/podman ]]; then
    CONTAINER_CMD=podman
    PRIVATE_UNSHARED=:Z
else
    dnf install podman
    CONTAINER_CMD=podman
    PRIVATE_UNSHARED=:Z
fi

SLINGSHOT_BUILD_CONTAINER=arti.hpc.amslabs.hpecorp.net/docker-remote/golangci/golangci-lint:latest 

${CONTAINER_CMD} run -v ${PWD}:${HOME}${PRIVATE_UNSHARED} -w ${HOME} ${SLINGSHOT_BUILD_CONTAINER} /bin/sh -c \
                 "cd ${HOME}; \
                  make -f Makefile.slingshot lint;"
