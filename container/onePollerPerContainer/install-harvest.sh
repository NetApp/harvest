#!/bin/bash -e

# Define usage message
usage() {
  echo "Usage: $(basename "$0")"
  echo "  --config <harvest.yml>    Path to the Harvest configuration file (required)"
  echo "  --action <start|stop|upgrade>   Action to perform (required)"
  echo "  [--image <harvest_image>]   Docker image to use (default: ghcr.io/netapp/harvest:latest)"
  echo ""
  echo "  start    Start the Harvest containers"
  echo "  stop     Stop the Harvest containers"
  echo "  upgrade  Upgrade the Harvest containers"
  exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -c|--config)
      shift
      harvest_config=$1
      ;;
    -a|--action)
      shift
      action=$1
      ;;
    -i|--image)
      shift
      harvest_image=$1
      ;;
    *)
      usage
      ;;
  esac
  shift
done


# Check if required arguments are present
if [[ -z $harvest_config || -z $action ]]; then
  usage
fi

# Check if the harvest configuration file exists
if [ ! -f "$harvest_config" ]; then
  echo "The specified Harvest configuration file does not exist: $harvest_config"
  exit 1
fi

# Check if the action is valid
if [[ $action != "start" && $action != "stop" && $action != "upgrade" ]]; then
  echo "The specified action is not valid. Valid actions are 'start', 'stop', or 'upgrade'."
  exit 1
fi

# Check if Docker and Docker Compose are installed
if ! command -v docker &> /dev/null; then
  echo "Docker could not be found. Please install Docker and try again."
  exit 1
fi

if ! command -v docker-compose &> /dev/null; then
  echo "Docker Compose could not be found. Please install Docker Compose and try again."
  exit 1
fi

# Set default image if not provided
image="${harvest_image:-ghcr.io/netapp/harvest:latest}"

# Function to start the Harvest containers
start_harvest() {
    # Remove any existing Docker container named "harvest_dummy" if it's running
    if [[ $(docker ps -aqf "name=harvest_dummy" -f "status=running") ]]; then
        echo "Removing existing running container named harvest_dummy"
        docker rm -f harvest_dummy
    fi

    # Start a new Docker container with the provided harvest.yml file and mount the current directory as a volume
    docker run --entrypoint "sh" --volume "$(realpath "$1"):/opt/harvest.yml" --name harvest_dummy -d -t "$2"

    # Create a cert directory to store certificates
    mkdir -p cert

    # Copy the conf directory from the Docker container to the local directory
    docker cp harvest_dummy:/opt/harvest/conf .

    # Generate the harvest-compose.yml file
    docker container exec -it harvest_dummy bin/harvest generate docker full --port --output harvest-compose.yml --config /opt/harvest.yml \
    --image "$harvest_image" --defaultMounts=false \
    --volume "$(realpath "./conf"):/opt/harvest/conf" --volume "$(realpath "./cert"):/opt/harvest/cert" --volume "$(realpath "$1"):/opt/harvest.yml"

    # Copy necessary files from the Docker container to the local directory
    docker cp harvest_dummy:/opt/harvest/prom-stack.yml ./prom-stack.yml
    docker cp harvest_dummy:/opt/harvest/grafana .
    docker cp harvest_dummy:/opt/harvest/container .
    docker cp harvest_dummy:/opt/harvest/harvest-compose.yml .
    if [[ $(docker ps -aqf "name=harvest_dummy") ]]; then
        docker rm -f harvest_dummy
    fi
}

# Check for the second argument and perform the corresponding action
if [[ $action == "start" ]]; then
    start_harvest "$harvest_config" "$image"
    docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
    echo "Harvest started successfully."
elif [[ $action == "stop" ]]; then
    docker-compose -f prom-stack.yml -f harvest-compose.yml down
    echo "Harvest stopped successfully."
elif [[ $action == "upgrade" ]]; then
    start_harvest "$harvest_config" "$image"
    docker-compose -f prom-stack.yml -f harvest-compose.yml up -d --remove-orphans
     echo "Harvest upgraded successfully."
fi