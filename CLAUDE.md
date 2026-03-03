# Validation

Run `make` to execute tests, linting, etc. This should be run to validate any change.

# Manual Validation with tmux

Manually validate by running examples in a tmux session and capturing the output.

## Basic workflow

```bash
# Start the example in a detached tmux session
tmux new-session -d -s kl -x 120 -y 40 "kl"

# Wait for the app to start, then capture the pane
sleep 2
tmux capture-pane -t kl -p -e    # with ANSI color/style escape sequences

# Clean up
tmux kill-session -t kl
```

### Sending key input

Use `tmux send-keys` to interact with the running example. Add a short sleep before capturing to let the UI update.

```bash
# Scroll down one line (j or Down)
tmux send-keys -t kl j && sleep 0.5 && tmux capture-pane -t kl -p

# Jump to bottom (shift+g, sent as literal G)
tmux send-keys -t kl G && sleep 0.5 && tmux capture-pane -t kl -p

# Jump to top
tmux send-keys -t kl g && sleep 0.5 && tmux capture-pane -t kl -p

# Page down
tmux send-keys -t kl f && sleep 0.5 && tmux capture-pane -t kl -p

# Toggle wrapping
tmux send-keys -t kl w && sleep 0.5 && tmux capture-pane -t kl -p

# Toggle selection
tmux send-keys -t kl s && sleep 0.5 && tmux capture-pane -t kl -p
```
