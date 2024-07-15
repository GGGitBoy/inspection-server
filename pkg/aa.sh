#!/bin/bash

commands=(
    "Kubelet Health Check:curl -sS http://localhost:10248/healthz"
    "API Server Ready Check:kubectl get --raw='/readyz'"
    "API Server Live Check:kubectl get --raw='/livez'"
    "ETCD Ready Check:kubectl get --raw='/readyz/etcd'"
    "ETCD Live Check:kubectl get --raw='/livez/etcd'"
    "Test Error command:test-error"
)

results=()

for command in "${commands[@]}"; do
    description=$(echo "$command" | cut -d ':' -f 1)
    command_desc=$(echo "$command" | cut -d ':' -f 2-)

    result=$(eval "${command_desc}" 2>&1)
    status=$?

    if [ $status -ne 0 ]; then
        result="{\"description\": \"${description}\", \"command\": \"${command_desc}\", \"error\": \"$(echo "$result" | tail -n 1)\"}"
    else
        result="{\"description\": \"${description}\", \"command\": \"${command_desc}\", \"response\": \"${result}\"}"
    fi

    results+=("$result")
done

echo -n "["
for ((i=0; i<${#results[@]}; i++)); do
#    echo -n "${results[i]}" | jq -c .
    echo -n "${results[i]}"
    if [ $i -lt $((${#results[@]} - 1)) ]; then
        echo -n ","
    fi
done
echo -n "]"
