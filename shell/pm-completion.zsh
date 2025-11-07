#compdef pm

# Zsh completion script for pm (Project Manager)
# Place this file in a directory in your $fpath, e.g.:
#   ~/.oh-my-zsh/completions/_pm
# or source it from your .zshrc

_pm_projects() {
  local -a proj_list
  proj_list=(${(f)"$(pm ls 2>/dev/null | tail -n +2 | awk '{print $1":"$2}')"})
  if [[ ${#proj_list[@]} -gt 0 ]]; then
    _describe 'projects' proj_list
  fi
}

_pm() {
  local curcontext="$curcontext" state line
  typeset -A opt_args

  _arguments -C \
    '1: :_pm_projects' \
    '*:: :->args'

  case $state in
    args)
      local project=$words[1]

      # Check if project exists
      if pm ls 2>/dev/null | grep -q "^$project"; then
        local -a all_opts
        all_opts=()

        # Add project commands (starting with :)
        local -a cmds
        cmds=(${(f)"$(pm $project :help 2>/dev/null | grep -E '^\s+:' | awk '{print $1":"substr($0, index($0,$2))}')"})
        all_opts+=($cmds)

        # Add docker groups (starting with @)
        local -a groups
        groups=(${(f)"$(pm $project :help 2>/dev/null | grep -E '^\s+@' | awk '{print $1":"substr($0, index($0,$2))}')"})
        all_opts+=($groups)

        if [[ ${#all_opts[@]} -gt 0 ]]; then
          _describe 'commands and groups' all_opts
        else
          _files
        fi
      else
        # If not a project, complete as normal command
        _files
      fi
      ;;
  esac
}

_pm "$@"
