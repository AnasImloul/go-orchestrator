#!/bin/bash

# Conventional Commit Helper Script
# Usage: ./scripts/commit.sh

echo "Conventional Commit Helper"
echo "========================="
echo ""

# Get commit type
echo "Select commit type:"
echo "1) feat - New feature"
echo "2) fix - Bug fix"
echo "3) docs - Documentation changes"
echo "4) style - Code style changes"
echo "5) refactor - Code refactoring"
echo "6) perf - Performance improvements"
echo "7) test - Adding or updating tests"
echo "8) chore - Maintenance tasks"
echo "9) ci - CI/CD changes"
echo "10) build - Build system changes"
echo ""

read -p "Enter choice (1-10): " choice

case $choice in
    1) TYPE="feat" ;;
    2) TYPE="fix" ;;
    3) TYPE="docs" ;;
    4) TYPE="style" ;;
    5) TYPE="refactor" ;;
    6) TYPE="perf" ;;
    7) TYPE="test" ;;
    8) TYPE="chore" ;;
    9) TYPE="ci" ;;
    10) TYPE="build" ;;
    *) echo "Invalid choice"; exit 1 ;;
esac

# Get scope
echo ""
echo "Select scope (optional):"
echo "1) di - Dependency injection"
echo "2) lifecycle - Lifecycle management"
echo "3) orchestrator - Main orchestrator"
echo "4) logger - Logging"
echo "5) examples - Examples"
echo "6) docs - Documentation"
echo "7) ci - CI/CD"
echo "8) Skip scope"
echo ""

read -p "Enter choice (1-8): " scope_choice

case $scope_choice in
    1) SCOPE="di" ;;
    2) SCOPE="lifecycle" ;;
    3) SCOPE="orchestrator" ;;
    4) SCOPE="logger" ;;
    5) SCOPE="examples" ;;
    6) SCOPE="docs" ;;
    7) SCOPE="ci" ;;
    8) SCOPE="" ;;
    *) echo "Invalid choice"; exit 1 ;;
esac

# Get subject
echo ""
read -p "Enter commit subject (brief description): " SUBJECT

# Get body
echo ""
read -p "Enter commit body (optional, press Enter to skip): " BODY

# Get breaking change info
echo ""
read -p "Is this a breaking change? (y/n): " BREAKING

# Build commit message
if [ -n "$SCOPE" ]; then
    COMMIT_MSG="$TYPE($SCOPE): $SUBJECT"
else
    COMMIT_MSG="$TYPE: $SUBJECT"
fi

if [ "$BREAKING" = "y" ] || [ "$BREAKING" = "Y" ]; then
    COMMIT_MSG="$TYPE!: $SUBJECT"
fi

if [ -n "$BODY" ]; then
    COMMIT_MSG="$COMMIT_MSG

$BODY"
fi

if [ "$BREAKING" = "y" ] || [ "$BREAKING" = "Y" ]; then
    COMMIT_MSG="$COMMIT_MSG

BREAKING CHANGE: $SUBJECT"
fi

echo ""
echo "Commit message:"
echo "==============="
echo "$COMMIT_MSG"
echo ""

read -p "Proceed with commit? (y/n): " CONFIRM

if [ "$CONFIRM" = "y" ] || [ "$CONFIRM" = "Y" ]; then
    git commit -m "$COMMIT_MSG"
    echo "Commit created successfully!"
else
    echo "Commit cancelled."
fi
