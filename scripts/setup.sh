#!/bin/bash
set -o nounset # Treat unset variables as an error
# Required to fix an issue with DNS in k3d
export K3D_FIX_DNS=1
#-----------------------------------------------------
# FUNCTIONS
#-----------------------------------------------------
# function for yes or no prompts when script requires user interaction
yes_no() {
    while true; do
        read -p "$1 (yes/no): " yn
        case $yn in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}
# Function to clean up the background process
cleanup() {
    if [[ -n "${PF_PID:-}" ]]; then
        kill "$PF_PID"
        wait "$PF_PID" 2>/dev/null
    fi
}
# Trap signals for clean exit
trap cleanup EXIT INT TERM
# function to check whether docker is running and k3d is installed
requirement_check() {
    # Check prerequisites: docker, k3d, helm, and kubectl
    prerequisites=(docker k3d helm kubectl)

    for app in ${prerequisites[@]}; do
        if ! hash $app 2>/dev/null; then
                echo >&2 "Required application '$app' is not found. Please install it to proceed."
            exit 1
        fi
    done
    if ! docker info &> /dev/null; then
        echo "Docker service is inactive. Please start Docker to continue."
        exit 1
    fi
}

# cluster_setup is a function to create a k3d cluster named "hello-world"
cluster_setup(){
    # Define local variables
    local cluster_name="hello-world"
    local configfile="./config/k3d.yaml"
    
    # Check if the cluster is already running
    if k3d cluster list | grep -qw "${cluster_name}"; then
        echo "Cluster "${cluster_name}" is already running"
        return 
    fi

    # Check if the config file exists
    if [[ ! -f ${configfile} ]]; then
        echo "Error: Configuration file does not exist."
        # Return with error code 1 if config file does not exist
        return 1
    fi

    # Create the cluster 
    echo "Initiating creation of the "${cluster_name}" k3d cluster.."
    # If the cluster creation fails, print an error message and return with error code 1
    if ! k3d cluster create --config "${configfile}"; then
        echo "Error: Failed to create "${cluster_name}" k3d cluster."
        return 1
    fi
    # Confirm current kube context before we proceed with any installs
    current_context=$(kubectl config current-context 2>/dev/null)
    if [ -z "$current_context" ]; then
        echo "Kubernetes context is not set."
        exit 1
    else
        echo "Current Kubernetes context is: $current_context"
        if yes_no "Proceed with installation of ArgoCD?"; then
          echo "installing argo........."
          install_argo
        else 
          echo "exiting"
          exit 1
        fi
    fi
}
# Function to install ArgoCD and follow the app of app approach
# https://argo-cd.readthedocs.io/en/stable/operator-manual/cluster-bootstrapping/
install_argo(){
    local namespace="argocd"
    helm dep update bootstrap/argocd >> /dev/null 2>&1 
    helm upgrade --wait -i -n "${namespace}" "${namespace}" bootstrap/argocd --create-namespace
    kubectl apply -f bootstrap/argocd.yml
    kubectl apply -f bootstrap/root.yml
    grafana
}
# Function that checks that a namepsace exists. We do this to prevent the script failing if Argo hasn't sync'd
check_namespace_exists() {
    local namespace=$1
    local timeout=$2
    local start_time=$(date +%s)  # Capture start time in seconds

    # Loop until the timeout is reached
    while true; do
        # Calculate elapsed time
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))

        # Check if timeout has been reached
        if [[ $elapsed_time -ge $timeout ]]; then
            echo "Timeout reached. Namespace '$namespace' does not exist after ${timeout} seconds."
            return 1
        fi
        # Query Kubernetes for the specified namespace
        if kubectl get namespace "$namespace" > /dev/null 2>&1; then
            return 0
        else
            echo "Namespace '$namespace' not found. Checking again in 10 seconds..."
            sleep 10
        fi
    done
}

# Function to handle Grafana access and port-forwarding
grafana() {
    local namespace="grafana"
    local service_name="grafana"
    local secret_name="grafana"
    local secret_key="admin-password"
    local local_port=3000
    local remote_port=80
    local admin_password
    local os
    local open_cmd

    if check_namespace_exists "$namespace" 120; then
        kubectl wait --for=jsonpath='{.status.phase}=Running' pods -l app.kubernetes.io/name=$namespace -n $namespace
        admin_password=$(kubectl get secret $secret_name --namespace $namespace --output=jsonpath="{.data.$secret_key}" | base64 --decode)
        printf "Grafana Admin Password: $admin_password\nGrafana User: admin\n"

        # Start port-forwarding in the background and capture its PID
        kubectl port-forward --pod-running-timeout=1m0s svc/$service_name $local_port:$remote_port --namespace $namespace &
        PF_PID=$!
        sleep 5 # wait for port forwarding to establish

        # Detect the operating system
        os=$(uname)
        case "$os" in
            Darwin) open_cmd="open" ;;
            Linux) open_cmd="xdg-open" ;;
            *) echo "Unable to automatically open browser. Please visit http://localhost:$local_port/login to view the dashboards."
               exit 1 ;;
        esac

        # Use the appropriate command to open the browser
        $open_cmd "http://localhost:$local_port/login"
    else
        echo "Unable to find $namespace namespace on cluster"
        exit 1
    fi
}

# Execute the function to check all requirements
main(){
    if requirement_check; then
        cluster_setup
    fi
    # Prevent the script from exiting immediately after setup
    read -n 1 -s -r -p 'Setup complete. Press any key to stop port-forwarding..\n'
    echo
}

main
