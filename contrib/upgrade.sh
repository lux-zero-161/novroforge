#!/usr/bin/env bash
# This is an update script for novroforge installed via the binary distribution
# from codeberg.org/novroforge/novroforge on linux as systemd service. It
# performs a backup and updates NovroForge in place.
# NOTE: This adds the GPG Signing Key of the NovroForge maintainers to the keyring.
# Depends on: bash, curl, xz, sha256sum. optionally jq, gpg
#   See section below for available environment vars.
#   When no version is specified, updates to the latest release.
# Examples:
#   upgrade.sh 1.15.10
#   novroforgehome=/opt/novroforge novroforgeconf=$novroforgehome/app.ini upgrade.sh

# Check if novroforge service is running
if ! pidof novroforge &> /dev/null; then
  echo "Error: novroforge is not running."
  exit 1
fi

# Continue with rest of the script if novroforge is running
echo "NovroForge is running. Continuing with rest of script..."

# apply variables from environment
: "${novroforgebin:="/usr/local/bin/novroforge"}"
: "${novroforgehome:="/var/lib/novroforge"}"
: "${novroforgeconf:="/etc/novroforge/app.ini"}"
: "${novroforgeuser:="git"}"
: "${sudocmd:="sudo"}"
: "${arch:="linux-amd64"}"
: "${service_start:="$sudocmd systemctl start novroforge"}"
: "${service_stop:="$sudocmd systemctl stop novroforge"}"
: "${service_status:="$sudocmd systemctl status novroforge"}"
: "${backupopts:=""}" # see `novroforge dump --help` for available options

function novroforgecmd {
  if [[ $sudocmd = "su" ]]; then
    # `-c` only accept one string as argument.
    "$sudocmd" - "$novroforgeuser" -c "$(printf "%q " "$novroforgebin" "--config" "$novroforgeconf" "--work-path" "$novroforgehome" "$@")"
  else
    "$sudocmd" --user "$novroforgeuser" "$novroforgebin" --config "$novroforgeconf" --work-path "$novroforgehome" "$@"
  fi
}

function require {
  for exe in "$@"; do
    command -v "$exe" &>/dev/null || (echo "missing dependency '$exe'"; exit 1)
  done
}

# parse command line arguments
while true; do
  case "$1" in
    -v | --version ) novroforgeversion="$2"; shift 2 ;;
    -y | --yes ) no_confirm="yes"; shift ;;
    --ignore-gpg) ignore_gpg="yes"; shift ;;
    "" | -- ) shift; break ;;
    * ) echo "Usage:  [<environment vars>] upgrade.sh [-v <version>] [-y] [--ignore-gpg]"; exit 1;; 
  esac
done

# exit once any command fails. this means that each step should be idempotent!
set -euo pipefail

if [[ -f /etc/os-release ]]; then
  os_release=$(cat /etc/os-release)

  if [[ "$os_release" =~ "OpenWrt" ]]; then
    sudocmd="su"
    service_start="/etc/init.d/novroforge start"
    service_stop="/etc/init.d/novroforge stop"
    service_status="/etc/init.d/novroforge status"
  else
    require systemctl
  fi
fi

require curl xz sha256sum "$sudocmd"

# select version to install
if [[ -z "${novroforgeversion:-}" ]]; then
  require jq
  novroforgeversion=$(curl --connect-timeout 10 -sL 'https://codeberg.org/api/v1/repos/novroforge/novroforge/releases?draft=false&pre-release=false&limit=1' -H 'accept: application/json' | jq -r '.[0].tag_name | sub("v"; "")')
  echo "Latest available version is $novroforgeversion"
fi

# confirm update
echo "Checking currently installed version..."
current=$(novroforgecmd --version | cut -d ' ' -f 3)
[[ "$current" == "$novroforgeversion" ]] && echo "$current is already installed, stopping." && exit 1
if [[ -z "${no_confirm:-}"  ]]; then
  echo "Make sure to read the changelog first: https://codeberg.org/novroforge/novroforge/src/branch/novroforge/CHANGELOG.md"
  echo "Are you ready to update novroforge from ${current} to ${novroforgeversion}? (y/N)"
  read -r confirm
  [[ "$confirm" == "y" ]] || [[ "$confirm" == "Y" ]] || exit 1
fi

echo "Upgrading novroforge from $current to $novroforgeversion ..."

pushd "$(pwd)" &>/dev/null
cd "$novroforgehome" # needed for novroforge dump later

# download new binary
binname="novroforge-${novroforgeversion}-${arch}"
binurl="https://codeberg.org/novroforge/novroforge/releases/download/v${novroforgeversion}/${binname}.xz"
echo "Downloading $binurl..."
curl --connect-timeout 10 --silent --show-error --fail --location -O "$binurl{,.sha256,.asc}"

# validate checksum & gpg signature
sha256sum -c "${binname}.xz.sha256"
if [[ -z "${ignore_gpg:-}" ]]; then
  require gpg
  gpg --keyserver keys.openpgp.org --recv EB114F5E6C0DC2BCDD183550A4B61A2DC5923710
  gpg --verify "${binname}.xz.asc" "${binname}.xz" || { echo 'Signature does not match'; exit 1; }
fi
rm "${binname}".xz.{sha256,asc}

# unpack binary + make executable
xz --decompress --force "${binname}.xz"
chown "$novroforgeuser" "$binname"
chmod +x "$binname"

# stop novroforge, create backup, replace binary, restart novroforge
echo "Flushing novroforge queues at $(date)"
novroforgecmd manager flush-queues
echo "Stopping novroforge at $(date)"
$service_stop
echo "Creating backup in $novroforgehome"
novroforgecmd dump $backupopts
echo "Updating binary at $novroforgebin"
cp -f "$novroforgebin" "$novroforgebin.bak" && mv -f "$binname" "$novroforgebin"
$service_start
$service_status

echo "Upgrade to $novroforgeversion successful!"

popd
