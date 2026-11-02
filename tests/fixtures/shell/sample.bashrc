# ~/.bashrc: executed by bash(1) for non-login shells.

export EDITOR="nvim"
export VISUAL="nvim"
export BROWSER=firefox

alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'
alias gs="git status"

export PATH="$HOME/.local/bin:$HOME/bin:$PATH"

# Node Version Manager
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

# Starship prompt
eval "$(starship init bash)"
