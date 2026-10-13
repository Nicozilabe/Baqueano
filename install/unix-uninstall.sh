#!/bin/bash

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
EXECUTABLE_NAME=sshm
EXECUTABLE_PATH="$INSTALL_DIR/$EXECUTABLE_NAME"
USE_SUDO="false"
OS=""
FORCE_UNINSTALL="${FORCE_UNINSTALL:-false}"
PURGE_CONFIG="${PURGE_CONFIG:-false}"

RED='\033[0;31m'
PURPLE='\033[0;35m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
    printf "${PURPLE}SSHM Uninstallation Script${NC}\n\n"
    printf "Usage:\n"
    printf "  Default:                     ${GREEN}bash unix-uninstall.sh${NC}\n"
    printf "  Force (no prompts):          ${GREEN}FORCE_UNINSTALL=true bash unix-uninstall.sh${NC}\n"
    printf "  Also remove config/backups:  ${GREEN}PURGE_CONFIG=true bash unix-uninstall.sh${NC}\n"
    printf "  Custom install directory:    ${GREEN}INSTALL_DIR=/opt/bin bash unix-uninstall.sh${NC}\n\n"
    printf "Environment variables:\n"
    printf "  FORCE_UNINSTALL - Skip confirmation prompts (default: false)\n"
    printf "  PURGE_CONFIG    - Also delete ~/.config/sshm (history, backups, credentials)\n"
    printf "  INSTALL_DIR     - Installation directory (default: /usr/local/bin)\n\n"
}

setSystem() {
    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

    # Use sudo if we're in linux or darwin (macOS) and sudo exists
    if [[ ( "$OS" = "linux" || "$OS" = "darwin" ) && -n "$(command -v sudo)" ]]; then
        USE_SUDO="true"
    fi
}

runAsRoot() {
    local CMD="$*"
    if [ "$USE_SUDO" = "true" ]; then
        printf "${PURPLE}We need sudo access to remove SSHM from $INSTALL_DIR ${NC}\n"
        CMD="sudo $CMD"
    fi
    $CMD
}

# configDir prints the SSHM configuration directory, mirroring the app logic
configDir() {
    if [ -n "$XDG_CONFIG_HOME" ]; then
        printf "%s" "$XDG_CONFIG_HOME/sshm"
    else
        printf "%s" "$HOME/.config/sshm"
    fi
}

# resolveBinary finds the installed sshm binary, preferring the configured
# INSTALL_DIR and falling back to whatever is on PATH
resolveBinary() {
    if [ -f "$EXECUTABLE_PATH" ]; then
        return 0
    fi

    local onPath
    onPath=$(command -v "$EXECUTABLE_NAME" 2>/dev/null)
    if [ -n "$onPath" ]; then
        EXECUTABLE_PATH="$onPath"
        return 0
    fi

    return 1
}

uninstall() {
    if ! resolveBinary; then
        printf "${YELLOW}SSHM does not appear to be installed (checked $INSTALL_DIR and PATH)${NC}\n"
    else
        printf "${YELLOW}Removing binary: $EXECUTABLE_PATH${NC}\n"
        runAsRoot rm -f "$EXECUTABLE_PATH"
        if [ $? -ne 0 ] || [ -e "$EXECUTABLE_PATH" ]; then
            printf "${RED}Failed to remove $EXECUTABLE_PATH${NC}\n"
            exit 1
        fi
        printf "${GREEN}Removed $EXECUTABLE_PATH${NC}\n"
    fi

    if [ "$PURGE_CONFIG" = "true" ]; then
        local dir
        dir=$(configDir)
        if [ -d "$dir" ]; then
            printf "${YELLOW}Removing configuration directory: $dir${NC}\n"
            rm -rf "$dir"
            printf "${GREEN}Removed $dir${NC}\n"
        else
            printf "${YELLOW}Configuration directory not found: $dir${NC}\n"
        fi
    fi
}

confirm() {
    if [ "$FORCE_UNINSTALL" = "true" ]; then
        return 0
    fi

    # Running via pipe (stdin is not a terminal): proceed without prompting
    if [ ! -t 0 ]; then
        printf "${YELLOW}Running via pipe - automatically proceeding with uninstallation...${NC}\n"
        return 0
    fi

    printf "${YELLOW}Do you want to uninstall SSHM? [y/N]: ${NC}"
    read -r response
    case "$response" in
        [yY][eE][sS]|[yY]) return 0 ;;
        *) return 1 ;;
    esac
}

main() {
    if [ "$1" = "-h" ] || [ "$1" = "--help" ] || [ "$1" = "help" ]; then
        usage
        exit 0
    fi

    printf "${PURPLE}Uninstalling SSHM - SSH Connection Manager${NC}\n\n"

    setSystem

    if ! confirm; then
        printf "${GREEN}Uninstallation cancelled.${NC}\n"
        exit 0
    fi

    uninstall

    printf "\n${GREEN}✅ SSHM was uninstalled successfully.${NC}\n"
    if [ "$PURGE_CONFIG" != "true" ]; then
        local dir
        dir=$(configDir)
        if [ -d "$dir" ]; then
            printf "${YELLOW}Your configuration was kept at $dir${NC}\n"
            printf "${YELLOW}Run with PURGE_CONFIG=true to remove it as well.${NC}\n"
        fi
    fi
}

main "$@"
