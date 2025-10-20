# Install sow CLI

You are helping the user install the `sow` CLI tool on their local machine. Follow these steps carefully to ensure a successful installation.

## Your Task

Guide the user through installing the sow CLI, preferring Homebrew when available, falling back to direct binary download when necessary.

## Step 1: Detect Platform and Environment

First, gather information about the user's system:

1. **Detect platform**: Run `uname -s` to determine OS (Darwin=macOS, Linux, MINGW/MSYS=Windows)
2. **Detect architecture**: Run `uname -m` to determine architecture (x86_64, arm64, aarch64)
3. **Check Homebrew availability**: Run `which brew` to see if Homebrew is installed (macOS/Linux only)
4. **Fetch latest version**: Get the latest release from GitHub API:
   ```bash
   curl -s https://api.github.com/repos/jmgilman/sow/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
   ```

## Step 2: Present Recommendation

Based on your detection, present the user with:

- Detected platform and architecture
- Available installation methods
- **Recommended method**:
  - If Homebrew is available → **Homebrew installation** (easier, handles updates)
  - Otherwise → **Direct binary download** (works everywhere)

**Important**: Ask the user to confirm they want to proceed with the recommended method before continuing.

## Step 3: Execute Installation

### Option A: Homebrew Installation (Preferred)

If Homebrew is available and user confirmed:

1. Run the installation command:
   ```bash
   brew install jmgilman/apps/sow
   ```

2. Wait for completion and show output to user

3. Skip to Step 4 (Verification)

### Option B: Direct Binary Download (Fallback)

If Homebrew is not available or user prefers direct download:

1. **Construct download URL** based on detected platform:

   Format: `https://github.com/jmgilman/sow/releases/download/{VERSION}/sow_{VERSION}_{OS}_{ARCH}.{EXT}`

   Where:
   - `{VERSION}` = latest version (e.g., `v1.0.0`)
   - `{OS}` = Platform name:
     - macOS: `Darwin`
     - Linux: `Linux`
     - Windows: `Windows`
   - `{ARCH}` = Architecture:
     - x86_64: `x86_64`
     - arm64/aarch64: `arm64`
   - `{EXT}` = Archive extension:
     - macOS/Linux: `tar.gz`
     - Windows: `zip`

   Example URLs:
   - macOS Intel: `https://github.com/jmgilman/sow/releases/download/v1.0.0/sow_v1.0.0_Darwin_x86_64.tar.gz`
   - Linux ARM: `https://github.com/jmgilman/sow/releases/download/v1.0.0/sow_v1.0.0_Linux_arm64.tar.gz`
   - Windows: `https://github.com/jmgilman/sow/releases/download/v1.0.0/sow_v1.0.0_Windows_x86_64.zip`

2. **Download and extract**:
   ```bash
   # Create temp directory
   TEMP_DIR=$(mktemp -d)
   cd $TEMP_DIR

   # Download (replace URL with constructed URL)
   curl -L -o archive.tar.gz "https://github.com/jmgilman/sow/releases/download/{VERSION}/{FILENAME}"

   # Extract
   tar -xzf archive.tar.gz  # Use 'unzip archive.zip' for Windows
   ```

3. **Install to ~/.local/bin**:
   ```bash
   # Create directory if it doesn't exist
   mkdir -p ~/.local/bin

   # Move binary
   mv sow ~/.local/bin/sow

   # Make executable (Unix-like systems)
   chmod +x ~/.local/bin/sow

   # Clean up
   cd ~
   rm -rf $TEMP_DIR
   ```

4. **Check PATH and update if needed**:

   Check if `~/.local/bin` is in PATH:
   ```bash
   echo $PATH | grep -q "$HOME/.local/bin" && echo "In PATH" || echo "Not in PATH"
   ```

   If NOT in PATH, ask the user which shell configuration file(s) to update:
   - `.zshrc` (Zsh - default on modern macOS)
   - `.bashrc` (Bash on Linux)
   - `.bash_profile` (Bash on macOS)
   - `.profile` (Generic POSIX shell)

   For each selected file:
   ```bash
   # Add to shell config (only if not already present)
   if ! grep -q 'export PATH="$HOME/.local/bin:$PATH"' ~/.zshrc; then
       echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
   fi
   ```

   **Inform the user** they need to either:
   - Restart their terminal, OR
   - Run `source ~/.zshrc` (or whichever file was modified)

## Step 4: Verify Installation

1. Run the version command:
   ```bash
   sow version
   ```

2. If successful, display:
   - ✅ Installation successful!
   - Installed version: `v1.0.0` (or whatever version)
   - Location: `/path/to/sow` (from `which sow`)

3. Show **Next Steps**:
   ```
   Next steps:
   1. Navigate to your git repository
   2. Run: sow init
   3. Follow the prompts to initialize sow in your repository

   For more information, visit: https://github.com/jmgilman/sow
   ```

## Error Handling

Handle these common issues gracefully:

- **Network failures**: Suggest checking internet connection, trying again
- **Permission errors**: Suggest using appropriate permissions or alternative install location
- **Unsupported platform**: Inform user and suggest building from source
- **Binary not found after install**: Check PATH configuration, suggest manual verification
- **GitHub API rate limit**: Suggest waiting or using a specific version number

## Important Notes

- Always show commands before running them (for transparency)
- Use the Bash tool to execute commands, not echo
- Check command exit codes and handle errors appropriately
- Be conversational and explain what you're doing at each step
- If something fails, explain clearly and suggest alternatives

---

**Remember**: You're helping the user install software on their machine. Be careful, transparent, and always confirm before making system changes.
