#!/bin/bash
# This script is used to start the services for the project and log their output.

# Create logs directory in wall-e-go/ if it doesn’t exist
mkdir -p logs

# Start auth service in a subshell to avoid changing the script’s directory
(cd auth && go run main.go serve > ../logs/auth.log 2>&1) &

# Start broker service in a subshell
(cd broker && go run main.go serve > ../logs/broker.log 2>&1) &

# Wait briefly
sleep 2

# Display all logs in the current terminal
multitail -f logs/*.log