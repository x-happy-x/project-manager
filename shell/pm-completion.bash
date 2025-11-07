#!/usr/bin/env bash

# Bash completion script for pm (Project Manager)
# Source this file from your .bashrc:
#   source /path/to/pm-completion.bash
# or copy to /etc/bash_completion.d/pm

_pm_completion() {
  local cur prev words cword
  _init_completion || return

  local projects commands docker_groups

  # If we're on the first argument, complete project names
  if [[ $cword -eq 1 ]]; then
    # Get list of projects
    projects=$(pm ls 2>/dev/null | tail -n +2 | awk '{print $1}')
    COMPREPLY=($(compgen -W "$projects" -- "$cur"))
    return 0
  fi

  # If we have a project name, complete commands/groups for that project
  if [[ $cword -ge 2 ]]; then
    local project="${words[1]}"

    # Check if the project exists
    if pm ls 2>/dev/null | grep -q "^$project"; then
      local help_output
      help_output=$(pm "$project" :help 2>/dev/null)

      # Extract commands (lines starting with whitespace followed by :)
      local cmds
      cmds=$(echo "$help_output" | grep -E '^\s+:' | awk '{print $1}')

      # Extract docker groups (lines starting with whitespace followed by @)
      local groups
      groups=$(echo "$help_output" | grep -E '^\s+@' | awk '{print $1}')

      # Combine all completions
      local all_completions="$cmds $groups"

      if [[ -n "$all_completions" ]]; then
        COMPREPLY=($(compgen -W "$all_completions" -- "$cur"))
        return 0
      fi
    fi

    # Fallback to file completion
    _filedir
    return 0
  fi
}

# Register the completion function
complete -F _pm_completion pm
