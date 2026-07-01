#!/usr/bin/env bash
set -euo pipefail

REPO="Solr159/JavBoss"
VERSION="v1.8.2"
INSTALL_DIR=""
TMP_DIR=""

prefers_chinese() {
  local locale_text
  locale_text="${JAVBOSS_LANG:-} ${LANGUAGE:-} ${LC_ALL:-} ${LC_MESSAGES:-} ${LANG:-}"
  case "$locale_text" in
    *zh*|*ZH*) return 0 ;;
    *) return 1 ;;
  esac
}

log() {
  printf '[javboss] %s\n' "$*"
}

die() {
  printf '[javboss] ERROR: %s\n' "$*" >&2
  exit 1
}

cleanup() {
  if [[ -n "${TMP_DIR:-}" ]]; then
    rm -rf "$TMP_DIR"
  fi
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

download_file() {
  local url="$1"
  local dest="$2"
  if command_exists curl; then
    curl -fL --retry 3 -o "$dest" "$url"
    return
  fi
  if command_exists wget; then
    wget --tries=3 -O "$dest" "$url"
    return
  fi
  die "curl or wget is required"
}

detect_platform() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Linux)
      case "$arch" in
        x86_64|amd64) printf 'linux-x86_64' ;;
        *) die "unsupported Linux architecture: $arch" ;;
      esac
      ;;
    Darwin)
      case "$arch" in
        x86_64|amd64) printf 'macos-x86_64' ;;
        arm64|aarch64) printf 'macos-arm64' ;;
        *) die "unsupported macOS architecture: $arch" ;;
      esac
      ;;
    *)
      die "unsupported OS: $os. Use scripts/install.ps1 on Windows"
      ;;
  esac
}

default_install_dir() {
  case "$(uname -s)" in
    Darwin) printf '%s/Applications/JavBoss' "$HOME" ;;
    *) printf '%s/.local/share/javboss' "$HOME" ;;
  esac
}

version_file_path() {
  printf '%s/.version' "$1"
}

installed_version() {
  local install_dir="$1"
  local version_file
  version_file="$(version_file_path "$install_dir")"
  if [[ -f "$install_dir/javboss" && -f "$version_file" ]]; then
    head -n 1 "$version_file" | tr -d '[:space:]'
  fi
}

write_installed_version() {
  local install_dir="$1"
  local tag="$2"
  printf '%s\n' "$tag" >"$(version_file_path "$install_dir")"
}

canonical_path() {
  local target="$1"
  local dir base
  dir="$(dirname "$target")"
  base="$(basename "$target")"
  if [[ -d "$dir" ]]; then
    (cd "$dir" && printf '%s/%s' "$(pwd -P)" "$base")
  else
    printf '%s' "$target"
  fi
}

running_javboss_pids() {
  local install_dir="$1"
  local exe
  exe="$(canonical_path "$install_dir/javboss")"

  if [[ "$(uname -s)" == "Linux" && -d /proc ]]; then
    local pid proc_exe
    for pid in $(pgrep -x javboss 2>/dev/null || true); do
      proc_exe="$(readlink "/proc/$pid/exe" 2>/dev/null || true)"
      if [[ "$proc_exe" == "$exe" ]]; then
        printf '%s\n' "$pid"
      fi
    done
    return
  fi

  if command_exists pgrep; then
    pgrep -f "$exe" 2>/dev/null || true
  fi
}

ensure_not_running() {
  local install_dir="$1"
  local pids
  if [[ ! -e "$install_dir/javboss" ]]; then
    return
  fi
  pids="$(running_javboss_pids "$install_dir" | tr '\n' ' ' | sed 's/[[:space:]]*$//')"
  if [[ -n "$pids" ]]; then
    if prefers_chinese; then
      die "JavBoss 已经从 $install_dir 启动（PID：$pids）。安装或升级前请先退出 JavBoss。"
    fi
    die "JavBoss is already running from $install_dir (pid: $pids). Please exit JavBoss before installing or upgrading."
  fi
}

copy_release_files() {
  local src="$1"
  local dest="$2"
  local protected_config=""

  mkdir -p "$dest"
  if [[ -f "$dest/config.toml" ]]; then
    protected_config="$(mktemp)"
    cp "$dest/config.toml" "$protected_config"
  fi

  rm -rf "$dest/internal" "$dest/web" "$dest/modernz"
  rm -f "$dest/javboss" "$dest/javboss.exe" "$dest/javboss.command"
  cp -R "$src"/. "$dest"/

  if [[ -n "$protected_config" ]]; then
    cp "$protected_config" "$dest/config.toml"
    rm -f "$protected_config"
  fi
}

create_command_link() {
  local install_dir="$1"
  local link_dir="$HOME/.local/bin"
  
  # 优化 1：解决 ~/.local/bin 创建时的权限不足问题
  if [[ ! -d "$link_dir" ]]; then
    if ! mkdir -p "$link_dir" 2>/dev/null; then
      log "Creating $link_dir requires administrator privileges..."
      sudo mkdir -p "$link_dir"
      sudo chown -R "$(whoami)" "$HOME/.local"
    fi
  fi

  ln -sf "$install_dir/javboss" "$link_dir/javboss"
  log "command installed: $link_dir/javboss"
}

setup_path_env() {
  local link_dir="$HOME/.local/bin"
  
  # 优化 3：如果不在 PATH 中，自动帮用户写入 shell 配置文件
  case ":$PATH:" in
    *":$link_dir:"*) ;;
    *)
      local shell_rc=""
      if [[ "${SHELL:-}" == *"zsh"* ]]; then
        shell_rc="$HOME/.zshrc"
      elif [[ "${SHELL:-}" == *"bash"* ]]; then
        if [[ "$(uname -s)" == "Darwin" ]]; then
          shell_rc="$HOME/.bash_profile"
        else
          shell_rc="$HOME/.bashrc"
        fi
      fi

      if [[ -n "$shell_rc" ]]; then
        log "Adding $link_dir to PATH in $shell_rc"
        printf '\n# JavBoss PATH\nexport PATH="$HOME/.local/bin:$PATH"\n' >> "$shell_rc"
        if prefers_chinese; then
          log "已自动将路径写入 $shell_rc，新开终端窗口后即可在任意地方直接输入 'javboss' 启动。"
        else
          log "PATH updated. Please restart your terminal or run 'source $shell_rc' to use 'javboss' command anywhere."
        fi
      else
        log "add $link_dir to PATH if the javboss command is not found"
      fi
      ;;
  esac
}

start_javboss() {
  local install_dir="$1"
  
  # 优化 2：移除 open 命令，直接在当前终端前台用 exec 启动，保持会话复用
  if [[ -e /dev/tty ]] && { : </dev/tty; } 2>/dev/null; then
    exec "$install_dir/javboss" </dev/tty
  else
    exec "$install_dir/javboss"
  fi
}

main() {
  command_exists unzip || die "unzip is required"

  local platform tag filename url zip_file extract_dir release_dir
  platform="$(detect_platform)"
  INSTALL_DIR="$(default_install_dir)"
  ensure_not_running "$INSTALL_DIR"

  tag="$VERSION"
  filename="javboss-${tag}-${platform}.zip"
  url="https://github.com/${REPO}/releases/download/${tag}/${filename}"

  if [[ "$(installed_version "$INSTALL_DIR")" == "$tag" ]]; then
    log "JavBoss $tag is already installed; no update needed"
    # 如果已经安装，也直接原地启动
    start_javboss "$INSTALL_DIR"
    return
  fi

  TMP_DIR="$(mktemp -d)"
  trap cleanup EXIT

  zip_file="$TMP_DIR/$filename"
  extract_dir="$TMP_DIR/extract"
  mkdir -p "$extract_dir"

  log "downloading $url"
  download_file "$url" "$zip_file"

  log "extracting release package"
  unzip -q "$zip_file" -d "$extract_dir"
  release_dir="$(find "$extract_dir" -mindepth 1 -maxdepth 1 -type d -print -quit)"
  [[ -n "$release_dir" && -f "$release_dir/javboss" ]] || die "release package layout is invalid"

  log "installing to $INSTALL_DIR"
  copy_release_files "$release_dir" "$INSTALL_DIR"
  chmod +x "$INSTALL_DIR/javboss" 2>/dev/null || true
  [[ ! -f "$INSTALL_DIR/javboss.command" ]] || chmod +x "$INSTALL_DIR/javboss.command" 2>/dev/null || true

  if command_exists xattr; then
    xattr -dr com.apple.quarantine "$INSTALL_DIR" >/dev/null 2>&1 || true
  fi

  create_command_link "$INSTALL_DIR"
  setup_path_env
  write_installed_version "$INSTALL_DIR" "$tag"

  log "installed JavBoss $tag"
  log "starting JavBoss"
  start_javboss "$INSTALL_DIR"
}

main "$@"
