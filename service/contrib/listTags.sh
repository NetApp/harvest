#!/bin/bash
set -e
#set -x

IMAGE_NAME=harvest

checkForJq() {
  command -v jq >/dev/null 2>&1 || {
    echo >&2 "jq is required but not found. Exiting. See https://stedolan.github.io/jq/download/"
    exit 1
  }
}

getToken() {
  local response

  response=$(curl --silent "https://netappdownloads.jfrog.io/artifactory/api/docker/oss-docker/v2/token")
  TOKEN=$(echo "$response" | jq '.token' | xargs echo)
}

listTags() {
  local imageName=$1

  curl --silent --header "Authorization: Bearer $TOKEN" \
    https://netappdownloads.jfrog.io/artifactory/api/docker/oss-docker/v2/"$imageName"/tags/list | jq .
}

checkForJq
getToken
listTags $IMAGE_NAME
