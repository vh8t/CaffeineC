#!/bin/bash
set -e

# Function to check and install packages using the package manager
install_package() {
    local package_manager=$1
    local package_name=$3
    local test_cmd=$2

    # Check if the package is already installed
    if ! command -v $test_cmd > /dev/null; then
        echo "Installing $package_name..."
        sudo $package_manager install -y $package_name
    else
        echo "$package_name is already installed."
    fi
}

# Auto-detect the OS and package manager
OS=$(uname -s | tr A-Z a-z)

case $OS in
  linux)
    source /etc/os-release
    case $ID in
      debian|ubuntu|mint)
        PACKAGE_MANAGER="apt-get"
        ;;

      fedora|rhel|centos)
        PACKAGE_MANAGER="yum"
        ;;

      opensuse*)
        PACKAGE_MANAGER="zypper"
        ;;

      arch)
        PACKAGE_MANAGER="pacman"
        ;;

      *)
        echo "Unsupported Linux distribution."
        exit 1
        ;;
    esac
    ;;

  darwin)
    PACKAGE_MANAGER="brew"
    ;;

  *)
    echo "Unsupported OS."
    exit 1
    ;;
esac

# Check and install LLVM and GCC
install_package "$PACKAGE_MANAGER" "llc" "llvm"
install_package "$PACKAGE_MANAGER" "gcc" "gcc"

# Directory to store user-installed binaries
install_dir="$HOME/.local/bin"

# Create the directory if it doesn't exist
mkdir -p $install_dir

# Download and install your compiler binary
latest_version=$(curl -sL https://github.com/vyPal/CaffeineC/releases/latest | grep -Eo 'tag/v[0-9\.]+' | head -n 1)
download_url="https://github.com/vyPal/CaffeineC/releases/latest/download/CaffeineC"

echo "Downloading CaffeineC version $latest_version..."
curl -sL $download_url -o $install_dir/CaffeineC

# Make the binary executable
chmod +x $install_dir/CaffeineC

# Determine the current shell
current_shell=$(basename "$SHELL")

# Set the autocomplete script URL based on the current shell
case $current_shell in
  bash)
    autocomplete_script_url="https://raw.githubusercontent.com/vyPal/CaffeineC/master/autocomplete/bash_autocomplete"
    shell_config_file="$HOME/.bashrc"
    ;;
  zsh)
    autocomplete_script_url="https://raw.githubusercontent.com/vyPal/CaffeineC/master/autocomplete/zsh_autocomplete"
    shell_config_file="$HOME/.zshrc"
    ;;
  *)
    echo "Unsupported shell for autocomplete. Skipping..."
    return
    ;;
esac

# If the shell is supported, continue with the rest of the script
if [ -n "$autocomplete_script_url" ]; then
  # Download the autocomplete script
  autocomplete_script_path="$install_dir/CaffeineC_autocomplete"
  echo "Downloading autocomplete script for $current_shell..."
  curl -sL $autocomplete_script_url -o $autocomplete_script_path

  # Source the downloaded script
  if [ "$current_shell" = "zsh" ]; then
    zsh -c "source $shell_config_file && source $autocomplete_script_path"
  else
    source $autocomplete_script_path
  fi

  # Add the source command to the shell's configuration file to make it persistent
  echo "source $autocomplete_script_path" >> $shell_config_file

  echo "Autocomplete script installed and sourced. It will be sourced automatically in new shell sessions."
fi

# Check if the install directory is in PATH
if [[ ":$PATH:" == *":$install_dir:"* ]]; then
    echo "The CaffeineC compiler is now installed and in your PATH."
else
    echo "Add the following line to your shell configuration file (e.g., .bashrc, .zshrc, .config/fish/config.fish):"
    echo "export PATH=\$PATH:$install_dir"
    echo "Then restart your terminal or run 'source <config-file>' to update the PATH."
fi

echo "Installation complete."