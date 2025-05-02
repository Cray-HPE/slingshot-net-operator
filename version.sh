#!/bin/bash

if [[ ! -z "${BUILD_NUMBER}" ]]; then

    if [[ $GIT_BRANCH == "release"* ]]; then
        DOCKER_REGISTRY='slingshot-docker-stable-local'
    elif [[ $GIT_BRANCH == "main" ]]; then
        DOCKER_REGISTRY='slingshot-docker-master-local'
    else
        DOCKER_REGISTRY='slingshot-docker-unstable-local'
    fi

    # If default branch is named main instead of master, checkout master branch from slingshot-version 
    CURRENT_BRANCH=`git branch --show-current`
    if [[ $CURRENT_BRANCH == "main" ]]; then
        CURRENT_BRANCH='master'
    fi

    git clone https://$HPE_GITHUB_TOKEN@github.hpe.com/hpe/hpc-sshot-slingshot-version.git

    cd hpc-sshot-slingshot-version

    if ! git checkout $CURRENT_BRANCH > /dev/null ; then
        echo "INFO: Branch "$CURRENT_BRANCH" is not an official Slingshot branch, using version string from master branch" >&2
    fi

    cd - > /dev/null

    PRODUCT_VERSION=$(cat hpc-sshot-slingshot-version/slingshot-version)

    if [[ -z "${PRODUCT_VERSION}" ]]; then
        echo "Version: ${PRODUCT_VERSION} is Empty"
        exit 1
    fi

    sed -i s/999.999.999/${PRODUCT_VERSION}-${BUILD_NUMBER}/g .version
    sed -i s/999.999.999/${PRODUCT_VERSION}-${BUILD_NUMBER}/g kubernetes/slingshot-net-operator/Chart.yaml
    sed -i s/@slingshot_docker_registry@/${DOCKER_REGISTRY}/g kubernetes/slingshot-net-operator/values.yaml

    echo "$(cat .version)"
fi
