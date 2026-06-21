#!/usr/bin/env bash
set -euo pipefail

# Import GPG private key from env var
if [ -z "${GPG_PRIVATE_KEY:-}" ]; then
  echo "::error::GPG_PRIVATE_KEY is not set"
  exit 1
fi

# GPG_PASSPHRASE may be unset or empty if the key has no passphrase.
# Default to empty string so gpg --passphrase works either way.
GPG_PASSPHRASE="${GPG_PASSPHRASE:-}"

# Import the key
echo "$GPG_PRIVATE_KEY" | gpg --batch --yes --pinentry-mode loopback \
  --passphrase "$GPG_PASSPHRASE" --import

# Extract fingerprint and key ID
FINGERPRINT=$(gpg --list-keys --with-colons | grep '^fpr' | head -1 | cut -d: -f10)
KEYID=$(echo "$FINGERPRINT" | tail -c 17)

echo "Fingerprint: $FINGERPRINT"
echo "KeyID:       $KEYID"

# Verify expected key
EXPECTED_KEYID="2B11E3055D7BAED4"
if [ "$KEYID" != "$EXPECTED_KEYID" ]; then
  echo "::error::Expected signing key $EXPECTED_KEYID, got $KEYID"
  exit 1
fi

# Export for GitHub Actions step outputs
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "fingerprint=$FINGERPRINT" >> "$GITHUB_OUTPUT"
  echo "keyid=$KEYID" >> "$GITHUB_OUTPUT"
fi
