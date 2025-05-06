#!/bin/bash
# This script is used to start the services for the project and log their output.

# Create logs directory in wall-e-go/ if it doesn’t exist
mkdir -p logs

# Start auth service in a subshell to avoid changing the script’s directory
(cd auth && go run main.go serve > ../logs/auth.log 2>&1) &

# Start the wallet service in a subshell
(cd wallet && go run main.go serve > ../logs/wallet.log 2>&1) &

# Wait briefly
sleep 1

# Start broker service in a subshell
(cd broker && go run main.go serve > ../logs/broker.log 2>&1) &


# Display all logs in the current terminal
tail -f logs/*.log