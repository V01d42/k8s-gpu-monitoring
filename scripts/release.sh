#!/bin/bash

set -e  # Exit on error

# Color constants for output messages
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Show help message
show_help() {
    echo "Usage: $0 <version>"
    echo ""
    echo "Example:"
    echo "  $0 1.0.2"
    echo ""
    echo "This script will:"
    echo "  1. Update Helm chart versions"
    echo "  2. Commit changes"
    echo "  3. Create and push git tag"
    echo "  4. Push changes to main branch"
    echo ""
}

# Validate version format
validate_version() {
    if [[ ! "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}Error: Version must be in format x.y.z (e.g., 1.0.2)${NC}"
        exit 1
    fi
}

# Check Git repository status
check_git_status() {
    # Check if current branch is main
    current_branch=$(git symbolic-ref --short HEAD)
    if [ "$current_branch" != "main" ]; then
        echo -e "${RED}Error: You must be on 'main' branch. Current branch: $current_branch${NC}"
        exit 1
    fi
    
    # Check for uncommitted changes
    if ! git diff-index --quiet HEAD --; then
        echo -e "${RED}Error: You have uncommitted changes. Please commit or stash them first.${NC}"
        git status --short
        exit 1
    fi
    
    # Check if branch is up to date with remote
    echo -e "${BLUE}Fetching latest changes...${NC}"
    git fetch origin
    
    if [ $(git rev-list HEAD...origin/main --count) -ne 0 ]; then
        echo -e "${RED}Error: Your branch is not up to date with origin/main${NC}"
        echo "Please run: git pull origin main"
        exit 1
    fi
}

# Update chart files
update_files() {
    local version=$1
    
    echo -e "${BLUE}Updating Helm chart files...${NC}"
    
    # Update Chart.yaml
    sed -i "s/version: .*/version: $version/" ./charts/k8s-gpu-monitoring-dev/Chart.yaml
    sed -i "s/appVersion: .*/appVersion: \"$version\"/" ./charts/k8s-gpu-monitoring-dev/Chart.yaml
    
    # Update values.yaml
    sed -i "s/tag: .*/tag: \"$version\"/g" ./charts/k8s-gpu-monitoring-dev/values.yaml
    
    echo -e "${GREEN}Updated files:${NC}"
    echo "  - charts/k8s-gpu-monitoring-dev/Chart.yaml"
    echo "  - charts/k8s-gpu-monitoring-dev/values.yaml"
}

# Show changes and confirm
show_changes() {
    echo -e "\n${YELLOW}Changes to be committed:${NC}"
    git diff --color=always ./charts/k8s-gpu-monitoring-dev/Chart.yaml ./charts/k8s-gpu-monitoring-dev/values.yaml
    
    echo -e "\n${YELLOW}Do you want to proceed with these changes? (y/N)${NC}"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo -e "${RED}Release canceled${NC}"
        # Revert changes
        git checkout -- ./charts/k8s-gpu-monitoring-dev/Chart.yaml ./charts/k8s-gpu-monitoring-dev/values.yaml
        exit 1
    fi
}

# Commit and push changes
commit_and_push() {
    local version=$1
    
    echo -e "${BLUE}Committing changes...${NC}"
    git add ./charts/k8s-gpu-monitoring-dev/Chart.yaml ./charts/k8s-gpu-monitoring-dev/values.yaml
    git commit -m "chore: release v$version"
    
    echo -e "${BLUE}Creating tag v$version...${NC}"
    git tag "v$version"
    
    echo -e "${BLUE}Pushing to origin...${NC}"
    git push origin main
    git push origin "v$version"
    
    echo -e "${GREEN}Successfully released v$version!${NC}"
    echo -e "${GREEN}Tag v$version has been pushed and GitHub Actions will build/deploy automatically${NC}"
}

# Main processing
main() {
    # Check arguments
    if [ $# -eq 0 ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_help
        exit 0
    fi
    
    local version=$1
    
    echo -e "${BLUE}Starting release process for version $version${NC}"
    
    # Run checks
    validate_version "$version"
    check_git_status
    
    # Update files
    update_files "$version"
    
    # Confirm changes
    show_changes
    
    # Commit and push
    commit_and_push "$version"
}

# Execute script
main "$@" 